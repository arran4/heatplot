package main

import (
	"errors"
	"image"
	"image/color"
	"image/gif"
	"log"
	"math"
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
	Speed           = 100 * time.Millisecond
	Step            = .1
)

type State interface {
	CurX() float64
	CurY() float64
	CurT() int
}

type RealState struct {
	X,Y float64
	T int
	AccessedX,AccessedY,AccessedT bool
}

func (rs *RealState) CurX() float64 {
	rs.AccessedX = true
	return rs.X
}

func (rs *RealState) CurY() float64 {
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

func (v Function) Evaluate(X, Y float64, T int) (weight float64, TUsed bool, err error) {
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
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
		}
	}()
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

type Plus struct {
	LHS Expression
	RHS Expression
}

func (v Plus) Evaluate(state State) float64 {
	return v.RHS.Evaluate(state) + v.LHS.Evaluate(state)
}

type Subtract struct {
	LHS Expression
	RHS Expression
}

func (v Subtract) Evaluate(state State) float64 {
	return v.RHS.Evaluate(state) - v.LHS.Evaluate(state)
}

type Multiply struct {
	LHS Expression
	RHS Expression
}

func (v Multiply) Evaluate(state State) float64 {
	return v.RHS.Evaluate(state) * v.LHS.Evaluate(state)
}

type Divide struct {
	LHS Expression
	RHS Expression
}

func (v Divide) Evaluate(state State) float64 {
	return v.RHS.Evaluate(state) / v.LHS.Evaluate(state)
}

type Power struct {
	LHS Expression
	RHS Expression
}

func (v Power) Evaluate(state State) float64 {
	return math.Pow(v.LHS.Evaluate(state), v.RHS.Evaluate(state))
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
		&Function{ // 0
			Equals: &Equals{
				LHS: &Var{
					Var: "Y",
				},
				RHS: &Const{
					Value: 4,
				},
			},
		},
		&Function{ // 1
			Equals: &Equals{
				LHS: &Var{
					Var: "Y",
				},
				RHS: &Const{
					Value: 4.5,
				},
			},
		},
		&Function{ // 2
			Equals: &Equals{
				LHS: &Var{
					Var: "Y",
				},
				RHS: &Var{
					Var: "X",
				},
			},
		},
		&Function{ // 3
			Equals: &Equals{
				LHS: &Var{
					Var: "Y",
				},
				RHS: &Multiply{
					LHS: &Var{
						Var: "X",
					},
					RHS: &Const{
						Value: 0.7,
					},
				},
			},
		},
		&Function{ // 4
			Equals: &Equals{
				LHS: &Var{
					Var: "Y",
				},
				RHS: &Multiply{
					LHS: &Var{
						Var: "X",
					},
					RHS: &Var{
						Var: "T",
					},
				},
			},
		},
		&Function{ // 5
			Equals: &Equals{
				LHS:  &Var{
					Var: "T",
				},
				RHS: &Plus{
					LHS: &Power{
						LHS: &Var{
							Var: "Y",
						},
						RHS: &Const{
							Value: 2,
						},
					},
					RHS: &Power{
						LHS: &Var{
							Var: "X",
						},
						RHS: &Const{
							Value: 2,
						},
					},
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
		if TUsed, err = plotFunction(img, imageSize, colours, functions[len(functions) - 1], t); err != nil {
			log.Panic(err)
		}
		if err := drawPlane(img, imageSize, colours); err != nil {
			log.Panic(err)
		}
		imgs = append(imgs, ConvertImageAxis(img, graphicSize, colours))
		delays = append(delays, int(Speed / (time.Millisecond * 10)))
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
			result.Set(x - size.Min.X , size.Dy() - (y - size.Min.Y) - 1, img.At(x, y))
		}
	}
	return result
}

func plotFunction(img *image.Paletted, size image.Rectangle, colours []color.Color, function *Function, t int) (TUsed bool, err error) {
	for x := size.Min.X; x < size.Max.X; x++ {
		for y := size.Min.Y; y < size.Max.Y; y++ {
			var w float64
			w, TUsed, err = function.Evaluate(float64(x) * Step, float64(y) * Step, t)
			if err != nil {
				return false, err
			}
			c := MakeHeatColour(w)
			if c != nil {
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
