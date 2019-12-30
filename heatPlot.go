package heatPlot

import (
	"errors"
	"fmt"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/gif"
	"io"
	"log"
	"math"
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
	goregularfnt *truetype.Font
)

type State interface {
	CurX() float64
	CurY() float64
	CurT() int
}

type RealState struct {
	X, Y                            float64
	T                               int
	AccessedX, AccessedY, AccessedT bool
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
	String() string
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
		return 0, false, errors.New("no such formula")
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

func (v Function) String() string {
	return v.Equals.String()
}

type Equals struct {
	LHS Expression
	RHS Expression
}

func (v Equals) Evaluate(state State) float64 {
	return v.RHS.Evaluate(state) - v.LHS.Evaluate(state)
}

func (v Equals) String() string {
	return fmt.Sprintf("%s = %s", v.LHS.String(), v.RHS.String())
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

func (v Var) String() string {
	return v.Var
}

type Const struct {
	Value float64
}

func (c Const) Evaluate(state State) float64 {
	return c.Value
}

func (v Const) String() string {
	return fmt.Sprintf("%g", v.Value)
}

type Plus struct {
	LHS Expression
	RHS Expression
}

func (v Plus) Evaluate(state State) float64 {
	return v.RHS.Evaluate(state) + v.LHS.Evaluate(state)
}

func (v Plus) String() string {
	return fmt.Sprintf("%s + %s", v.LHS.String(), v.RHS.String())
}

type Subtract struct {
	LHS Expression
	RHS Expression
}

func (v Subtract) Evaluate(state State) float64 {
	return v.RHS.Evaluate(state) - v.LHS.Evaluate(state)
}

func (v Subtract) String() string {
	return fmt.Sprintf("%s - %s", v.LHS.String(), v.RHS.String())
}

type Multiply struct {
	LHS Expression
	RHS Expression
}

func (v Multiply) Evaluate(state State) float64 {
	return v.RHS.Evaluate(state) * v.LHS.Evaluate(state)
}

func (v Multiply) String() string {
	return fmt.Sprintf("%s * %s", v.LHS.String(), v.RHS.String())
}

type Divide struct {
	LHS Expression
	RHS Expression
}

func (v Divide) Evaluate(state State) float64 {
	return v.RHS.Evaluate(state) / v.LHS.Evaluate(state)
}

func (v Divide) String() string {
	return fmt.Sprintf("%s / %s", v.LHS.String(), v.RHS.String())
}

type Power struct {
	LHS Expression
	RHS Expression
}

func (v Power) Evaluate(state State) float64 {
	return math.Pow(v.LHS.Evaluate(state), v.RHS.Evaluate(state))
}

func (v Power) String() string {
	return fmt.Sprintf("%s ^ %s", v.LHS.String(), v.RHS.String())
}

type Negate struct {
	Expr Expression
}

func (v Negate) Evaluate(state State) float64 {
	return -v.Expr.Evaluate(state)
}

func (v Negate) String() string {
	return fmt.Sprintf("-(%s)", v.Expr.String())
}

type Brackets struct {
	Expr Expression
}

func (v Brackets) Evaluate(state State) float64 {
	return v.Expr.Evaluate(state)
}

func (v Brackets) String() string {
	return fmt.Sprintf("(%s)", v.Expr.String())
}

func init() {
	if fnt, err := truetype.Parse(goregular.TTF); err != nil {
		log.Panic(err)
	} else {
		goregularfnt = fnt
	}
}

func RunFunction(functionString string, w io.Writer, size, timeLowerBound, timeUpperBound, scale, heatColourCount int, pixelSize float64, speed time.Duration) {
	plotSize := image.Rect(-size, -size, size, size)
	colours := []color.Color{
		lineColor,
		color.White,
		color.Black,
	}
	colours = append(colours, HeatColours(heatColourCount)...)
	imgs := []*image.Paletted{}
	delays := []int{}
	TUsed := false
	function := parseFunction(functionString)
	for t := (timeLowerBound); t < (timeUpperBound) && TUsed || t == (timeLowerBound); t++ {
		img := image.NewPaletted(plotSize, colours)
		if err := paintWhite(img, plotSize); err != nil {
			log.Panic(err)
		}
		var err error
		if TUsed, err = plotFunction(img, plotSize, function, t, heatColourCount, pixelSize); err != nil {
			log.Panic(err)
		}
		if err := drawPlane(img, plotSize); err != nil {
			log.Panic(err)
		}
		img = FlipAndMoveImage(img)
		if img, err = AddHeaderAndFooter(img, function, t, timeUpperBound); err != nil {
			log.Panic(err)
		}
		img = ScaleImage(img, scale)
		imgs = append(imgs, img)
		delays = append(delays, int((speed)/(time.Millisecond*10)))
	}
	if err := gif.EncodeAll(w, &gif.GIF{
		Image: imgs,
		Delay: delays,
	}); err != nil {
		log.Panic(err)
	}
}

func parseFunction(arg string) *Function {
	if r := yyParse(NewCalcLexer(arg)); r != 0 {
		log.Panic("Invalid formula: ", arg)
	}
	return yyResult
}

func AddHeaderAndFooter(img *image.Paletted, function *Function, t, timeUpperBound int) (*image.Paletted, error) {
	borderSizes := image.Pt(20, 20)
	newRect := image.Rect(img.Rect.Min.X, img.Rect.Min.Y, img.Rect.Max.X+borderSizes.X*2, img.Rect.Max.Y+borderSizes.Y*2)
	result := image.NewPaletted(newRect, img.Palette)
	if err := paintWhite(result, newRect); err != nil {
		return nil, err
	}
	for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
		for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
			result.Set(x+borderSizes.X, y+borderSizes.Y, img.At(x, y))
		}
	}
	if err := AddText(fmt.Sprintf("%s", function.String()), result, newRect.Min.X, newRect.Min.Y+borderSizes.Y); err != nil {
		return nil, err
	}
	if err := AddText(fmt.Sprintf("T: %d/%d - https://github.com/arran4/", t, (timeUpperBound)), result, newRect.Min.X, newRect.Max.Y-10); err != nil {
		return nil, err
	}
	return result, nil
}

func AddText(s string, img *image.Paletted, x int, y int) error {
	face := truetype.NewFace(goregularfnt, &truetype.Options{
		Size:       12 * 2,
		DPI:        40,
		Hinting:    0,
		SubPixelsX: 0,
		SubPixelsY: 0,
	})
	d := &font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(s)
	return nil
}

func ScaleImage(img *image.Paletted, scale int) *image.Paletted {
	result := image.NewPaletted(image.Rect(0, 0, img.Rect.Dx()*scale, img.Rect.Dy()*scale), img.Palette)
	for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
		for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
			for xs := 0; xs < scale; xs++ {
				for ys := 0; ys < scale; ys++ {
					result.Set(x*scale+xs, y*scale+ys, img.At(x, y))
				}
			}
		}
	}
	return result
}

func FlipAndMoveImage(img *image.Paletted) *image.Paletted {
	result := image.NewPaletted(image.Rect(0, 0, img.Rect.Dx(), img.Rect.Dy()), img.Palette)
	for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
		for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
			result.Set(x-img.Rect.Min.X, img.Rect.Dy()-(y-img.Rect.Min.Y)-1, img.At(x, y))
		}
	}
	return result
}

func plotFunction(img *image.Paletted, size image.Rectangle, function *Function, t, heatColourCount int, pixelSize float64) (TUsed bool, err error) {
	for x := size.Min.X; x < size.Max.X; x++ {
		for y := size.Min.Y; y < size.Max.Y; y++ {
			var w float64
			w, TUsed, err = function.Evaluate(float64(x)*(pixelSize), float64(y)*(pixelSize), t)
			if err != nil {
				return false, err
			}
			c := MakeHeatColour(heatColourCount, w)
			if c != nil {
				img.Set(x, y, c)
			}
		}
	}
	return
}

func HeatColours(heatColourCount int) []color.Color {
	result := make([]color.Color, (heatColourCount)*2-1, (heatColourCount)*2-1)
	for i := 1; i < (heatColourCount)*2; i++ {
		v := MakeHeatColour(heatColourCount, float64(i)/float64(heatColourCount)-1)
		if v == nil {
			log.Panic("Got nil heat colour....", i)
		}
		result[i-1] = v
	}
	return result
}

func MakeHeatColour(heatColourCount int, i float64) color.Color {
	if i <= -1 || i >= 1 {
		return nil
	}
	c := int((i * 100.0) * (1.0 / float64(heatColourCount) * 100.0))
	if c == 0 {
		return color.Black
	}
	r, b := uint8(255), uint8(255)
	if i > 0 {
		r = uint8(255 - int(float64(c)*256.0/float64(heatColourCount)))
	} else {
		b = uint8(255 + int(float64(c)*256.0/float64(heatColourCount)))
	}
	return &color.RGBA{
		R: r,
		G: 0,
		B: b,
		A: 0xFF,
	}
}

type Image interface {
	Set(x, y int, c color.Color)
	image.Image
}

func paintWhite(img Image, size image.Rectangle) error {
	for x := size.Min.X; x < size.Max.X; x++ {
		for y := size.Min.Y; y < size.Max.Y; y++ {
			img.Set(x, y, color.White)
		}
	}
	return nil
}

func drawPlane(img Image, size image.Rectangle) error {
	for x := size.Min.X; x < size.Max.X; x++ {
		y := 0
		img.Set(x, y, lineColor)
	}
	for y := size.Min.Y; y < size.Max.Y; y++ {
		x := 0
		img.Set(x, y, lineColor)
	}
	return nil
}
