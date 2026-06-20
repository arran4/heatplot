package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"bitbucket.org/arran4/heatplot"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	sizeevent "golang.org/x/mobile/event/size"
)

var (
	testUI    = flag.String("test-ui", "", "Path to a JSON file containing key events to test UI headlessly")
	testUIOut = flag.String("test-ui-out", "ui-test-out.png", "Path to output the final UI image in test-ui mode")
)

type UIState struct {
	Formula    string
	T          int
	Typing     bool
	TypingText string
	ShowHelp   bool
	Quit       bool
}

func (s *UIState) handleKeyEvent(code key.Code, r rune) bool {
	changed := false
	if s.Typing {
		if code == key.CodeReturnEnter || code == key.CodeKeypadEnter {
			s.Formula = s.TypingText
			s.Typing = false
			changed = true
		} else if code == key.CodeEscape {
			s.Typing = false
			changed = true
		} else if code == key.CodeDeleteBackspace {
			runes := []rune(s.TypingText)
			if len(runes) > 0 {
				s.TypingText = string(runes[:len(runes)-1])
				changed = true
			}
		} else if r != 0 {
			s.TypingText += string(r)
			changed = true
		}
	} else {
		if code == key.CodeRightArrow {
			s.T++
			changed = true
		} else if code == key.CodeLeftArrow {
			s.T--
			changed = true
		} else if r == 't' || r == 'T' {
			s.Typing = true
			s.TypingText = s.Formula
			changed = true
		} else if code == key.CodeF1 || r == '?' {
			s.ShowHelp = !s.ShowHelp
			changed = true
		} else if code == key.CodeEscape || r == 'q' || r == 'Q' {
			s.Quit = true
			changed = true
		}
	}
	return changed
}

func runUI() {
	if *testUI != "" {
		runUITest()
		return
	}

	driver.Main(func(sc screen.Screen) {
		w, err := sc.NewWindow(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()

		formula := "y = x * sin(t/10)"
		if flag.NArg() > 1 {
			formula = flag.Arg(1)
		}

		state := &UIState{
			Formula: formula,
			T:       *timeLowerBound,
		}

		var sz sizeevent.Event
		var b screen.Buffer

		for {
			switch e := w.NextEvent().(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
			case key.Event:
				if e.Direction == key.DirPress || e.Direction == key.DirNone { // sometimes repeat is DirNone
					if state.handleKeyEvent(e.Code, e.Rune) {
						if state.Quit {
							return
						}
						w.Send(paint.Event{})
					}
				}
			case paint.Event:
				w.Fill(sz.Bounds(), color.White, draw.Src)

				img, err := generateImageForUI(state)
				if err != nil {
					img = errorImage(err.Error())
				}

				// Upload to shiny buffer
				if img != nil {
					if b == nil || b.Bounds().Max != img.Bounds().Max {
						if b != nil {
							b.Release()
						}
						var err error
						b, err = sc.NewBuffer(img.Bounds().Max)
						if err != nil {
							b = nil
						}
					}

					if b != nil {
						draw.Draw(b.RGBA(), b.Bounds(), img, image.Point{}, draw.Src)
						w.Upload(image.Point{}, b, b.Bounds())
					}
				}

				w.Publish()
			case sizeevent.Event:
				sz = e
				w.Send(paint.Event{})
			}
		}
	})
}

type TestEvent struct {
	Code string `json:"Code"`
	Rune string `json:"Rune"`
}

func parseKeyCode(codeStr string) key.Code {
	switch codeStr {
	case "CodeReturnEnter":
		return key.CodeReturnEnter
	case "CodeEscape":
		return key.CodeEscape
	case "CodeDeleteBackspace":
		return key.CodeDeleteBackspace
	case "CodeRightArrow":
		return key.CodeRightArrow
	case "CodeLeftArrow":
		return key.CodeLeftArrow
	case "CodeF1":
		return key.CodeF1
	}
	return key.CodeUnknown
}

func runUITest() {
	data, err := os.ReadFile(*testUI)
	if err != nil {
		log.Fatalf("Failed to read test UI file: %v", err)
	}

	var events []TestEvent
	if err := json.Unmarshal(data, &events); err != nil {
		log.Fatalf("Failed to parse test UI JSON: %v", err)
	}

	formula := "y = x * sin(t/10)"
	if flag.NArg() > 1 {
		formula = flag.Arg(1)
	}

	state := &UIState{
		Formula: formula,
		T:       *timeLowerBound,
	}

	for _, ev := range events {
		code := parseKeyCode(ev.Code)
		r := rune(0)
		if len(ev.Rune) > 0 {
			r = []rune(ev.Rune)[0]
		}
		state.handleKeyEvent(code, r)
		if state.Quit {
			break
		}
	}

	img, err := generateImageForUI(state)
	if err != nil {
		img = errorImage(err.Error())
	}

	outFile, err := os.Create(*testUIOut)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer func() { _ = outFile.Close() }()

	if err := png.Encode(outFile, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
}

func errorImage(msg string) image.Image {
	colours := []color.Color{color.White, color.Black}
	img := image.NewPaletted(image.Rect(0, 0, 800, 200), colours)
	_ = heatPlot.AddText(msg, img, 20, 100, 2)
	return img
}

func generateImageForUI(state *UIState) (img image.Image, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error parsing formula: %v", r)
		}
	}()

	plotSize := image.Rect(-*size, -*size, *size, *size)
	function := heatPlot.ParseFunction(state.Formula)

	plot, tUsed, err2 := function.PlotForT(plotSize, state.T, *pointSize)
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
	draw.Draw(pImg, pImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

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

	hintText := ""
	if state.Typing {
		hintText = "Typing: " + state.TypingText + "_"
	} else if !state.ShowHelp {
		hintText = "[? or F1 for Help]"
	}

	footerUrl := ""
	if *testUI != "" { // Only display footer url in export mode
		footerUrl = *footerText
	}

	pImg, err2 = heatPlot.AddHeaderAndFooterWithHint(pImg, function, plot.T, *timeUpperBound, *scale, tUsed, footerUrl, hintText)
	if err2 != nil {
		return nil, err2
	}

	if state.ShowHelp {
		// Draw a semi-transparent overlay box for help text readability
		helpText := []string{
			"Commands (Esc / q / F1 / ? to dismiss):",
			"  t       : Edit formula",
			"  Left    : Decrease T",
			"  Right   : Increase T",
		}

		// Draw a semi-transparent overlay box for help text readability, centered.
		// Measure width of longest string to bound the box precisely.
		longestWidth := 0
		for _, text := range helpText {
			width := len(text) * 8 * (*scale) // Approximation
			if width > longestWidth {
				longestWidth = width
			}
		}

		boxWidth := longestWidth + 40*(*scale) // Extra padding
		boxHeight := (20 * (*scale)) + (len(helpText) * 20 * (*scale))

		// Center the box horizontally
		imgWidth := pImg.Bounds().Dx()
		startX := (imgWidth - boxWidth) / 2
		if startX < 10*(*scale) {
			startX = 10*(*scale)
		}

		overlayRect := image.Rect(startX, 30*(*scale), startX+boxWidth, 30*(*scale)+boxHeight)

		// Draw border
		borderRect := image.Rect(startX-2*(*scale), 30*(*scale)-2*(*scale), startX+boxWidth+2*(*scale), 30*(*scale)+boxHeight+2*(*scale))
		draw.Draw(pImg, borderRect, &image.Uniform{C: color.Black}, image.Point{}, draw.Src)

		// Draw gray background - make it lighter by increasing alpha transparency or using lighter gray
		draw.DrawMask(pImg, overlayRect, &image.Uniform{C: color.RGBA{220, 220, 220, 255}}, image.Point{}, &image.Uniform{C: color.Alpha{220}}, image.Point{}, draw.Over)

		for i, text := range helpText {
			_ = heatPlot.AddText(text, pImg, startX+20*(*scale), 50*(*scale)+i*20*(*scale), *scale)
		}
	}

	return pImg, nil
}
