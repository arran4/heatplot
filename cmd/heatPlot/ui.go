package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"

	"bitbucket.org/arran4/heatplot"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	sizeevent "golang.org/x/mobile/event/size"
)

func runUI() {
	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()

		formula := "y = x * sin(t/10)"
		if flag.NArg() > 1 {
			formula = flag.Arg(1)
		}

		t := *timeLowerBound
		typing := false
		typingText := ""

		var sz sizeevent.Event

		for {
			switch e := w.NextEvent().(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
			case key.Event:
				if e.Direction == key.DirPress || e.Direction == key.DirNone { // sometimes repeat is DirNone
					if typing {
						if e.Code == key.CodeReturnEnter || e.Code == key.CodeKeypadEnter {
							formula = typingText
							typing = false
							w.Send(paint.Event{})
						} else if e.Code == key.CodeEscape {
							typing = false
							w.Send(paint.Event{})
						} else if e.Code == key.CodeDeleteBackspace {
							if len(typingText) > 0 {
								typingText = typingText[:len(typingText)-1]
								w.Send(paint.Event{})
							}
						} else if e.Rune != 0 {
							typingText += string(e.Rune)
							w.Send(paint.Event{})
						}
					} else {
						if e.Code == key.CodeRightArrow {
							t++
							w.Send(paint.Event{})
						} else if e.Code == key.CodeLeftArrow {
							t--
							w.Send(paint.Event{})
						} else if e.Rune == 't' || e.Rune == 'T' {
							typing = true
							typingText = formula
							w.Send(paint.Event{})
						}
					}
				}
			case paint.Event:
				w.Fill(sz.Bounds(), color.White, draw.Src)

				img, err := generateImageForUI(formula, t, typing, typingText)
				if err != nil {
					img = errorImage(err.Error())
				}

				// Upload to shiny buffer
				if img != nil {
					b, err := s.NewBuffer(img.Bounds().Max)
					if err == nil {
						draw.Draw(b.RGBA(), b.Bounds(), img, image.Point{}, draw.Src)
						w.Upload(image.Point{}, b, b.Bounds())
						b.Release()
					}
				}

				w.Publish()
			case sizeevent.Event:
				sz = e
			}
		}
	})
}

func errorImage(msg string) image.Image {
	colours := []color.Color{color.White, color.Black}
	img := image.NewPaletted(image.Rect(0, 0, 800, 200), colours)
	for x := 0; x < 800; x++ {
		for y := 0; y < 200; y++ {
			img.Set(x, y, color.White)
		}
	}
	_ = heatPlot.AddText(msg, img, 20, 100, 2)
	return img
}

func generateImageForUI(formula string, t int, typing bool, typingText string) (img image.Image, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error parsing formula: %v", r)
		}
	}()

	plotSize := image.Rect(-*size, -*size, *size, *size)
	function := heatPlot.ParseFunction(formula)

	plot, tUsed, err2 := function.PlotForT(plotSize, t, *pointSize)
	if err2 != nil {
		return nil, err2
	}

	colours := []color.Color{
		color.RGBA{R: 0x0F, G: 0x0F, B: 0x0F, A: 0xFF}, // lineColor
		color.White,
		color.Black,
	}
	colours = append(colours, heatPlot.HeatColours(*heatColourCount)...)

	pImg := image.NewPaletted(plotSize, colours)

	// paintWhite
	for x := plotSize.Min.X; x < plotSize.Max.X; x++ {
		for y := plotSize.Min.Y; y < plotSize.Max.Y; y++ {
			pImg.Set(x, y, color.White)
		}
	}

	// plot.Draw
	_ = plot.Draw(pImg, *heatColourCount)

	// drawPlane
	for x := plotSize.Min.X; x < plotSize.Max.X; x++ {
		pImg.Set(x, 0, colours[0])
	}
	for y := plotSize.Min.Y; y < plotSize.Max.Y; y++ {
		pImg.Set(0, y, colours[0])
	}

	pImg = heatPlot.FlipAndMoveImage(pImg)
	pImg = heatPlot.ScaleImage(pImg, *scale)

	footer := *footerText
	if typing {
		footer = "Typing: " + typingText
	} else {
		footer = "[Press 't' to type formula, Left/Right to change T] " + footer
	}

	// heatPlot.AddHeaderAndFooter panics if function is nil, but ParseFunction panics before returning nil, so function won't be nil.
	pImg, err2 = heatPlot.AddHeaderAndFooter(pImg, function, plot.T, *timeUpperBound, *scale, tUsed, footer)
	if err2 != nil {
		return nil, err2
	}

	return pImg, nil
}
