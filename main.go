package main

import (
	"errors"
	"image"
	"image/color"
	"image/gif"
	"log"
	"os"
	"strings"
	"time"
)

var (
	lineColor = color.RGBA{
		R: 0x0F,
		G: 0x0F,
		B: 0x0F,
		A: 0xFF,
	}
)

const (
	HeatColourCount = 126
)

type State interface {
	CurX() int
	CurY() int
	CurT() int
}

type RealState struct {
	X,Y,T int
	AccessedX,AccessedY,AccessedT bool
}

func (rs *RealState) CurX() int {
	rs.AccessedX = true
	return rs.X
}

func (rs *RealState) CurY() int {
	rs.AccessedY = true
	return rs.Y
}

func (rs *RealState) CurT() int {
	rs.AccessedT = true
	return rs.T
}

type Expression interface {
	Evaluate(state State) float64
}

type Function struct {
	Equals *Equals
}

func (v Function) Evaluate(X, Y, T int) (weight float64, TUsed bool, err error) {
	state := &RealState{
		X:         X,
		Y:         Y,
		T:         T,
		AccessedX: false,
		AccessedY: false,
		AccessedT: false,
	}
	if v.Equals == nil {
		return 0,false, errors.New("no such formula")
	}
	weight = v.Equals.Evaluate(state)
	TUsed = state.AccessedT
	return
}

type Equals struct {
	LHS Expression
	RHS Expression
}

func (v Equals) Evaluate(state State) float64 {
	return v.RHS.Evaluate(state) - v.LHS.Evaluate(state)
}

type Var struct {
	Var string
}

func (v Var) Evaluate(state State) float64 {
	switch strings.ToUpper(v.Var) {
	case "X":
		return float64(state.CurX())
	case "Y":
		return float64(state.CurY())
	case "T":
		return float64(state.CurT())
	default:
		return 0
	}
}

type Const struct {
	Value float64
}

func (c Const) Evaluate(state State) float64 {
	return c.Value
}

func main() {
	log.SetFlags(log.Flags()|log.Lshortfile)
	graphicSize := image.Rect(-120,-120, 120, 120)
	imageSize := image.Rect(-100,-100, 100, 100)
	colours := []color.Color{
		lineColor,
		color.White,
		color.Black,
	}
	colours = append(colours, HeatColours()...)
	functions := []*Function {
		&Function{
			Equals: &Equals{
				LHS: &Var{
					Var: "Y",
				},
				RHS: &Const{
					Value: 4,
				},
			},
		},
		&Function{
			Equals: &Equals{
				LHS: &Var{
					Var: "Y",
				},
				RHS: &Const{
					Value: 4.5,
				},
			},
		},
		&Function{
			Equals: &Equals{
				LHS: &Var{
					Var: "Y",
				},
				RHS: &Var{
					Var: "X",
				},
			},
		},
	}
	imgs := []*image.Paletted{}
	delays := []int{}
	TUsed := false
	for t := 0; t < 360 && TUsed || t == 0; t++ {
		img := image.NewPaletted(graphicSize, colours)
		if err := paintWhite(img, graphicSize, colours); err != nil {
			log.Panic(err)
		}
		var err error
		if TUsed, err = plotFunction(img, imageSize, colours, functions[2], t); err != nil {
			log.Panic(err)
		}
		if err := drawPlane(img, imageSize, colours); err != nil {
			log.Panic(err)
		}
		imgs = append(imgs, ConvertImageAxis(img, graphicSize, colours))
		delays = append(delays, int(1 * time.Second / (time.Millisecond * 10)))
	}
	w, err := os.OpenFile("./out.gif", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer w.Close()
	if err := gif.EncodeAll(w, &gif.GIF{
		Image:           imgs,
		Delay:           delays,
		//Config:          image.Config{},
	}); err != nil {
		log.Panic(err)
	}
}

func ConvertImageAxis(img *image.Paletted, size image.Rectangle, colours []color.Color) *image.Paletted {
	result := image.NewPaletted(image.Rect(0,0, size.Dx(), size.Dy()), colours)
	for x := size.Min.X; x < size.Max.X; x++ {
		for y := size.Min.Y; y < size.Max.Y; y++ {
			result.Set(x - size.Min.X , y - size.Min.Y, img.At(x, y))
		}
	}
	return result
}

func plotFunction(img *image.Paletted, size image.Rectangle, colours []color.Color, function *Function, t int) (TUsed bool, err error) {
	for x := size.Min.X; x < size.Max.X; x++ {
		for y := size.Min.Y; y < size.Max.Y; y++ {
			var w float64
			w, TUsed, err = function.Evaluate(x,y,t)
			if err != nil {
				return false, err
			}
			c := MakeHeatColour(w)
			if c != nil {
				log.Print(x, y, c)
				img.Set(x, y, c)
			}
		}
	}
	return
}

func HeatColours() []color.Color {
	result := make([]color.Color, HeatColourCount*2-1, HeatColourCount*2-1)
	for i := 1; i < HeatColourCount*2; i++ {
		v := MakeHeatColour(float64(i)/HeatColourCount - 1)
		if v == nil {
			log.Panic("Got nil heat colour....", i)
		}
		result[i-1] = v
	}
	return result
}

func MakeHeatColour(i float64) color.Color {
	if i <= -1 || i >= 1 {
		return nil
	}
	c := int((i * 100.0) * (1.0 / float64(HeatColourCount) * 100.0))
	log.Print(i, c)
	if c == 0 {
		return color.Black
	}
	r, b := uint8(255), uint8(255)
	if i > 0 {
		r = uint8(255 - int(float64(c) * 256.0 / float64(HeatColourCount)))
	} else {
		b = uint8(255 + int(float64(c) * 256.0 / float64(HeatColourCount)))
	}
	return &color.RGBA{
		R: r,
		G: 255,
		B: b,
		A: 0xFF,
	}
}

type Image interface {
	Set(x, y int, c color.Color)
	image.Image
}

func paintWhite(img Image, size image.Rectangle, colours []color.Color) error {
	for x := size.Min.X; x < size.Max.X; x++ {
		for y := size.Min.Y; y < size.Max.Y; y++ {
			img.Set(x,y,color.White)
		}
	}
	return nil
}

func drawPlane(img Image, size image.Rectangle, colours []color.Color) error {
	for x := size.Min.X; x < size.Max.X; x++ {
		y := 0
		img.Set(x,y, lineColor)
	}
	for y := size.Min.Y; y < size.Max.Y; y++ {
		x := 0
		img.Set(x,y, lineColor)
	}
	return nil
}
