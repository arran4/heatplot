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
	goregularfnt    *truetype.Font
	SingleFunctions map[string]SingleFunctionDef
	DoubleFunctions map[string]DoubleFunctionDef
	FunctionNames   []string
)

type SingleFunctionDef func(float64) float64
type DoubleFunctionDef func(float64, float64) float64

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
	Depth() int
	Simplify() Expression
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

func (v Function) Simplify() *Function {
	e := v.Equals.Simplify().(Equals)
	v.Equals = &e
	return &v
}

type Equals struct {
	LHS Expression
	RHS Expression
}

func (v Equals) Evaluate(state State) float64 {
	return v.RHS.Evaluate(state) - v.LHS.Evaluate(state)
}

func (v Equals) Depth() int {
	l, r := v.LHS.Depth(), v.RHS.Depth()
	if l > r {
		return l + 1
	}
	return r + 1
}

func (v Equals) String() string {
	return fmt.Sprintf("%s = %s", v.LHS.String(), v.RHS.String())
}

func (v Equals) Simplify() Expression {
	v.RHS = v.RHS.Simplify()
	v.LHS = v.LHS.Simplify()
	return v
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

func (v Var) Depth() int {
	return 1
}

func (v Var) String() string {
	return v.Var
}

func (v Var) Simplify() Expression {
	return v
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

func (v Const) Simplify() Expression {
	return v
}

func (v Const) Depth() int {
	return 1
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

func (v Plus) Simplify() Expression {
	v.RHS = v.RHS.Simplify()
	v.LHS = v.LHS.Simplify()
	return v
}

func (v Plus) Depth() int {
	l, r := v.LHS.Depth(), v.RHS.Depth()
	if l > r {
		return l + 1
	}
	return r + 1
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

func (v Subtract) Simplify() Expression {
	v.RHS = v.RHS.Simplify()
	v.LHS = v.LHS.Simplify()
	return v
}

func (v Subtract) Depth() int {
	l, r := v.LHS.Depth(), v.RHS.Depth()
	if l > r {
		return l + 1
	}
	return r + 1
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

func (v Multiply) Simplify() Expression {
	v.RHS = v.RHS.Simplify()
	v.LHS = v.LHS.Simplify()
	return v
}

func (v Multiply) Depth() int {
	l, r := v.LHS.Depth(), v.RHS.Depth()
	if l > r {
		return l + 1
	}
	return r + 1
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

func (v Divide) Simplify() Expression {
	v.RHS = v.RHS.Simplify()
	v.LHS = v.LHS.Simplify()
	return v
}

func (v Divide) Depth() int {
	l, r := v.LHS.Depth(), v.RHS.Depth()
	if l > r {
		return l + 1
	}
	return r + 1
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

func (v Power) Simplify() Expression {
	v.RHS = v.RHS.Simplify()
	v.LHS = v.LHS.Simplify()
	return v
}

func (v Power) Depth() int {
	l, r := v.LHS.Depth(), v.RHS.Depth()
	if l > r {
		return l + 1
	}
	return r + 1
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

func (v Modulus) Simplify() Expression {
	v.RHS = v.RHS.Simplify()
	v.LHS = v.LHS.Simplify()
	return v
}

func (v Modulus) Depth() int {
	l, r := v.LHS.Depth(), v.RHS.Depth()
	if l > r {
		return l + 1
	}
	return r + 1
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

func (v Negate) Simplify() Expression {
	if nn, ok := v.Expr.(*Negate); ok {
		return nn.Expr.Simplify()
	}
	if nb, ok := v.Expr.(*Brackets); ok {
		if nn, ok := nb.Expr.(*Negate); ok {
			return nn.Expr.Simplify()
		}
	}
	v.Expr = v.Expr.Simplify()
	return v
}

func (v Negate) Depth() int {
	return v.Expr.Depth() + 1
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

func (v Brackets) Simplify() Expression {
	v.Expr = v.Expr.Simplify()
	return v
}

func (v Brackets) Depth() int {
	return v.Expr.Depth() + 1
}

type SingleFunction struct {
	Name string
	Expr Expression
}

func (v SingleFunction) Evaluate(state State) float64 {
	var r = v.Expr.Evaluate(state)
	if f, ok := SingleFunctions[strings.ToUpper(v.Name)]; ok {
		r = f(r)
	}
	return r
}

func (v SingleFunction) String() string {
	return fmt.Sprintf("%s(%s)", v.Name, v.Expr.String())
}

func (v SingleFunction) Simplify() Expression {
	v.Expr = v.Expr.Simplify()
	return v
}

func (v SingleFunction) Depth() int {
	return v.Expr.Depth() + 1
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
	if f, ok := DoubleFunctions[strings.ToUpper(v.Name)]; ok {
		r1 = f(r1, r2)
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

func (v DoubleFunction) Simplify() Expression {
	v.Expr1 = v.Expr1.Simplify()
	v.Expr2 = v.Expr2.Simplify()
	return v
}

func (v DoubleFunction) Depth() int {
	l, r := v.Expr1.Depth(), v.Expr2.Depth()
	if l > r {
		return l + 1
	}
	return r + 1
}

func init() {
	if fnt, err := truetype.Parse(goregular.TTF); err != nil {
		log.Panic(err)
	} else {
		goregularfnt = fnt
	}
	SingleFunctions = map[string]SingleFunctionDef{}
	DoubleFunctions = map[string]DoubleFunctionDef{}
	FunctionNames = []string{}
	for name, f := range map[string]interface{}{
		"Abs":             math.Abs,
		"Acos":            math.Acos,
		"Acosh":           math.Acosh,
		"Asin":            math.Asin,
		"Asinh":           math.Asinh,
		"Atan":            math.Atan,
		"Atan2":           math.Atan2,
		"Atanh":           math.Atanh,
		"Cbrt":            math.Cbrt,
		"Ceil":            math.Ceil,
		"Copysign":        math.Copysign,
		"Cos":             math.Cos,
		"Cosh":            math.Cosh,
		"Dim":             math.Dim,
		"Erf":             math.Erf,
		"Erfc":            math.Erfc,
		"Erfcinv":         math.Erfcinv,
		"Erfinv":          math.Erfinv,
		"Exp":             math.Exp,
		"Exp2":            math.Exp2,
		"Expm1":           math.Expm1,
		"Float32bits":     math.Float32bits,
		"Float32frombits": math.Float32frombits,
		"Float64bits":     math.Float64bits,
		"Float64frombits": math.Float64frombits,
		"Floor":           math.Floor,
		"Frexp":           math.Frexp,
		"Gamma":           math.Gamma,
		"Hypot":           math.Hypot,
		"Ilogb":           math.Ilogb,
		"Inf":             math.Inf,
		"IsInf":           math.IsInf,
		"IsNaN":           math.IsNaN,
		"J0":              math.J0,
		"J1":              math.J1,
		"Jn":              math.Jn,
		"Ldexp":           math.Ldexp,
		"Lgamma":          math.Lgamma,
		"Log":             math.Log,
		"Log10":           math.Log10,
		"Log1p":           math.Log1p,
		"Log2":            math.Log2,
		"Logb":            math.Logb,
		"Max":             math.Max,
		"Min":             math.Min,
		"Mod":             math.Mod,
		"Modf":            math.Modf,
		"NaN":             math.NaN,
		"Nextafter":       math.Nextafter,
		"Nextafter32":     math.Nextafter32,
		"Pow":             math.Pow,
		"Pow10":           math.Pow10,
		"Remainder":       math.Remainder,
		"Round":           math.Round,
		"RoundToEven":     math.RoundToEven,
		"Signbit":         math.Signbit,
		"Sin":             math.Sin,
		"Sincos":          math.Sincos,
		"Sinh":            math.Sinh,
		"Sqrt":            math.Sqrt,
		"Tan":             math.Tan,
		"Tanh":            math.Tanh,
		"Trunc":           math.Trunc,
		"Y0":              math.Y0,
		"Y1":              math.Y1,
		"Yn":              math.Yn,
	} {
		switch f := f.(type) {
		case func(float64) float64:
			SingleFunctions[strings.ToUpper(name)] = f
		case func(float64, float64) float64:
			DoubleFunctions[strings.ToUpper(name)] = f
		case func(int, float64) float64:
			DoubleFunctions[strings.ToUpper(name)] = func(f1 float64, f2 float64) float64 {
				return f(int(f1), f2)
			}
		case func(float64, int) float64:
			DoubleFunctions[strings.ToUpper(name)] = func(f1 float64, f2 float64) float64 {
				return f(f1, int(f2))
			}
		case func(int) float64:
			SingleFunctions[strings.ToUpper(name)] = func(f1 float64) float64 {
				return f(int(f1))
			}
		case func(float64) int:
			SingleFunctions[strings.ToUpper(name)] = func(f1 float64) float64 {
				return float64(f(f1))
			}
		default:
			continue
		}
		FunctionNames = append(FunctionNames, name)
	}
}

func ParseRunAndDrawFunction(functionString string, w io.Writer, size, timeLowerBound, timeUpperBound, scale, heatColourCount int, pointSize float64, speed time.Duration, footerText string) {
	function := ParseFunction(functionString)
	function.PlotAndDraw(w, size, timeLowerBound, timeUpperBound, scale, heatColourCount, pointSize, speed, footerText)
}

func (function *Function) PlotAndDraw(w io.Writer, size, timeLowerBound, timeUpperBound, scale, heatColourCount int, pointSize float64, speed time.Duration, footerText string) {
	plotSize := image.Rect(-size, -size, size, size)
	tUsed, plots := function.Plot(timeLowerBound, timeUpperBound, plotSize, pointSize)
	RenderPlots(heatColourCount, plots, plotSize, scale, function, timeUpperBound, tUsed, footerText, speed, w)
}

func RenderPlots(heatColourCount int, plots []*Plot, plotSize image.Rectangle, scale int, function *Function, timeUpperBound int, tUsed bool, footerText string, speed time.Duration, w io.Writer) {
	delays := []int{}
	colours := []color.Color{
		lineColor,
		color.White,
		color.Black,
	}
	colours = append(colours, HeatColours(heatColourCount)...)
	imgs := []*image.Paletted{}
	for _, plot := range plots {
		img := image.NewPaletted(plotSize, colours)
		if err := paintWhite(img, plotSize); err != nil {
			log.Panic(err)
		}
		if err := plot.Draw(img, heatColourCount); err != nil {
			log.Panic(err)
		}
		if err := drawPlane(img, plotSize); err != nil {
			log.Panic(err)
		}
		img = FlipAndMoveImage(img)
		img = ScaleImage(img, scale)
		var err error
		if img, err = AddHeaderAndFooter(img, function, plot.T, timeUpperBound, scale, tUsed, footerText); err != nil {
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

func (function *Function) Plot(timeLowerBound int, timeUpperBound int, plotSize image.Rectangle, pointSize float64) (tUsed bool, plots []*Plot) {
	for t := (timeLowerBound); t < (timeUpperBound) && tUsed || t == (timeLowerBound); t++ {
		var err error
		var plot *Plot
		if plot, tUsed, err = function.PlotForT(plotSize, t, pointSize); err != nil {
			log.Panic(err)
		}
		plots = append(plots, plot)
	}
	return
}

func ParseFunction(arg string) *Function {
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

type Plot struct {
	Size   image.Rectangle
	Values []float64
	Sets   int
	T      int
}

func (plot *Plot) Draw(img *image.Paletted, heatColourCount int) (err error) {
	for x := plot.Size.Min.X; x < plot.Size.Max.X; x++ {
		for y := plot.Size.Min.Y; y < plot.Size.Max.Y; y++ {
			c := MakeHeatColour(heatColourCount, plot.Get(x, y))
			if c != nil {
				img.Set(x, y, c)
			}
		}
	}
	return
}

func (plot *Plot) Set(x int, y int, w float64) {
	pos := plot.GetPos(x, y)
	if pos < 0 || pos > len(plot.Values) {
		return
	}
	plot.Values[pos] = w
	if w >= -1 && w <= 1 {
		plot.Sets++
	}
}

func (plot *Plot) Get(x int, y int) float64 {
	pos := plot.GetPos(x, y)
	if pos < 0 || pos > len(plot.Values) {
		return 0.0
	}
	return plot.Values[pos]
}

func (plot *Plot) GetPos(x int, y int) int {
	absX := x - plot.Size.Min.X
	absY := y - plot.Size.Min.Y
	return absY*plot.Size.Dx() + absX
}

func (plot *Plot) Equals(plot2 *Plot) bool {
	if plot == nil || plot2 == nil {
		return plot == plot2
	}
	if len(plot.Values) != len(plot2.Values) {
		return false
	}
	for i := range plot.Values {
		if plot.Values[i] != plot2.Values[i] {
			return false
		}
	}
	return true
}

func (function *Function) PlotForT(size image.Rectangle, t int, pointSize float64) (plot *Plot, TUsed bool, err error) {
	plot = &Plot{
		Size:   size,
		Values: make([]float64, size.Dy()*size.Dx(), size.Dy()*size.Dx()),
		T:      t,
	}
	for x := size.Min.X; x < size.Max.X; x++ {
		for y := size.Min.Y; y < size.Max.Y; y++ {
			var w float64
			w, TUsed, err = function.Evaluate(float64(x)*(pointSize), float64(y)*(pointSize), t)
			if err != nil {
				return nil, false, err
			}
			plot.Set(x, y, w)
		}
	}
	return
}

func (v Function) Depth() int {
	return v.Equals.Depth()
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
