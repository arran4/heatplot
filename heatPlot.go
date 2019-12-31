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

type Modulus struct {
	LHS Expression
	RHS Expression
}

func (v Modulus) Evaluate(state State) float64 {
	return math.Mod(v.LHS.Evaluate(state), v.RHS.Evaluate(state))
}

func (v Modulus) String() string {
	return fmt.Sprintf("%s %% %s", v.LHS.String(), v.RHS.String())
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

type SingleFunction struct {
	Name string
	Expr Expression
}

func (v SingleFunction) Evaluate(state State) float64 {
	var r = v.Expr.Evaluate(state)
	switch strings.ToUpper(v.Name) {
	case "ABS":
		r = math.Abs(r)
	case "ACOS":
		r = math.Acos(r)
	case "ACOSH":
		r = math.Acosh(r)
	case "ASIN":
		r = math.Asin(r)
	case "ASINH":
		r = math.Asinh(r)
	case "ATAN":
		r = math.Atan(r)
	case "ATANH":
		r = math.Atanh(r)
	case "CBRT":
		r = math.Cbrt(r)
	case "CEIL":
		r = math.Ceil(r)
	case "COS":
		r = math.Cos(r)
	case "COSH":
		r = math.Cosh(r)
	case "ERF":
		r = math.Erf(r)
	case "ERFC":
		r = math.Erfc(r)
	case "ERFCINV":
		r = math.Erfcinv(r)
	case "ERFINV":
		r = math.Erfinv(r)
	case "EXP":
		r = math.Exp(r)
	case "EXP2":
		r = math.Exp2(r)
	case "EXPM1":
		r = math.Expm1(r)
	case "FLOOR":
		r = math.Floor(r)
	case "GAMMA":
		r = math.Gamma(r)
	case "J0":
		r = math.J0(r)
	case "J1":
		r = math.J1(r)
	case "LOG":
		r = math.Log(r)
	case "LOG10":
		r = math.Log10(r)
	case "LOG1P":
		r = math.Log1p(r)
	case "LOG2":
		r = math.Log2(r)
	case "LOGB":
		r = math.Logb(r)
	case "ROUND":
		r = math.Round(r)
	case "ROUNDTOEVEN":
		r = math.RoundToEven(r)
	case "SIN":
		r = math.Sin(r)
	case "SINH":
		r = math.Sinh(r)
	case "SQRT":
		r = math.Sqrt(r)
	case "TAN":
		r = math.Tan(r)
	case "TANH":
		r = math.Tanh(r)
	case "TRUNC":
		r = math.Trunc(r)
	case "Y0":
		r = math.Y0(r)
	case "Y1":
		r = math.Y1(r)
	}
	return r
}

func (v SingleFunction) String() string {
	return fmt.Sprintf("%s(%s)", v.Name, v.Expr.String())
}

type DoubleFunction struct {
	Name  string
	Expr1 Expression
	Expr2 Expression
	Infix bool
}

func (v DoubleFunction) Evaluate(state State) float64 {
	var r1 = v.Expr1.Evaluate(state)
	var r2 = v.Expr2.Evaluate(state)
	switch strings.ToUpper(v.Name) {
	case "ATAN2":
		r1 = math.Atan2(r1, r2)
	case "COPYSIGN":
		r1 = math.Copysign(r1, r2)
	case "HYPOT":
		r1 = math.Hypot(r1, r2)
	case "NEXTAFTER":
		r1 = math.Nextafter(r1, r2)
	case "POW":
		r1 = math.Pow(r1, r2)
	case "LDEXP":
		r1 = math.Ldexp(r1, int(r2))
	case "MAX":
		r1 = math.Max(r1, r2)
	case "MIN":
		r1 = math.Min(r1, r2)
	case "MOD":
		r1 = math.Mod(r1, r2)
	case "REMAINDER":
		r1 = math.Remainder(r1, r2)
	case "DIM":
		r1 = math.Dim(r1, r2)
	}
	return r1
}

func (v DoubleFunction) String() string {
	if v.Infix {
		return fmt.Sprintf("%s %s %s", v.Expr1.String(), v.Name, v.Expr2.String())
	} else {
		return fmt.Sprintf("%s(%s, %s)", v.Name, v.Expr1.String(), v.Expr2.String())
	}
}

func init() {
	if fnt, err := truetype.Parse(goregular.TTF); err != nil {
		log.Panic(err)
	} else {
		goregularfnt = fnt
	}
}

func RunFunction(functionString string, w io.Writer, size, timeLowerBound, timeUpperBound, scale, heatColourCount int, pixelSize float64, speed time.Duration, footerText string) {
	plotSize := image.Rect(-size, -size, size, size)
	colours := []color.Color{
		lineColor,
		color.White,
		color.Black,
	}
	colours = append(colours, HeatColours(heatColourCount)...)
	imgs := []*image.Paletted{}
	delays := []int{}
	tUsed := false
	function := parseFunction(functionString)
	for t := (timeLowerBound); t < (timeUpperBound) && tUsed || t == (timeLowerBound); t++ {
		img := image.NewPaletted(plotSize, colours)
		if err := paintWhite(img, plotSize); err != nil {
			log.Panic(err)
		}
		var err error
		if tUsed, err = plotFunction(img, plotSize, function, t, heatColourCount, pixelSize); err != nil {
			log.Panic(err)
		}
		if err := drawPlane(img, plotSize); err != nil {
			log.Panic(err)
		}
		img = FlipAndMoveImage(img)
		img = ScaleImage(img, scale)
		if img, err = AddHeaderAndFooter(img, function, t, timeUpperBound, scale, tUsed, footerText); err != nil {
			log.Panic(err)
		}
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

func AddHeaderAndFooter(img *image.Paletted, function *Function, t, timeUpperBound, scale int, tUsed bool, footerText string) (*image.Paletted, error) {
	borderSizes := image.Pt(20*scale, 20*scale)
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
	if err := AddText(fmt.Sprintf("%s", function.String()), result, newRect.Min.X+10, newRect.Min.Y+borderSizes.Y, scale); err != nil {
		return nil, err
	}
	if tUsed {
		if err := AddText(fmt.Sprintf("T: %d/%d - %s", t, (timeUpperBound), footerText), result, newRect.Min.X+10, newRect.Max.Y-10, scale); err != nil {
			return nil, err
		}
	} else {
		if err := AddText(fmt.Sprintf("%s", footerText), result, newRect.Min.X+10, newRect.Max.Y-10, scale); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func AddText(s string, img *image.Paletted, x, y, scale int) error {
	face := truetype.NewFace(goregularfnt, &truetype.Options{
		Size:       12 * 2 * float64(scale),
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
