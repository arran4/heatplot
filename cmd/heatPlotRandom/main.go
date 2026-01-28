package main

import (
	"bitbucket.org/arran4/heatplot"
	"flag"
	"fmt"
	"image"
	"log"
	"math/rand"
	"os"
	"time"
)

var (
	heatColourCount = flag.Int("hcc", 126, "Heat colour count.. The number of distinct colours. Can't exceed 254 in total. This value is multiplied by 2. Shouldn't change this.")
	speed           = flag.Duration("speed", 100*time.Millisecond, "The number of microseconds to wait between each frame")
	pointSize       = flag.Float64("pointSize", .1, "How many x or y steps a pixel is. Ie .1 will mean that every 10 unscaled pixels is 1 normal step")
	scale           = flag.Int("scale", 2, "Magnification of the picture")
	timeLowerBound  = flag.Int("tlb", 0, "where to start T")
	timeUpperBound  = flag.Int("tub", 25, "Where to end t")
	size            = flag.Int("size", 100, "The size for each direction in the cartesian plane. Ie 100 would be -100 to 100 on the x and y axis")
	outputFile      = flag.String("outputFile", "./out.gif", "The output filename")
	footerText      = flag.String("footerText", "RND", "Text to put at the bottom of the picture")
	rng             *rand.Rand
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func main() {
	seed := time.Now().UnixNano()
	rng = rand.New(rand.NewSource(seed))
	flag.Parse()
	w, err := os.OpenFile(*outputFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer w.Close()
	for {
		function := randomFunction(10)
		if function == nil {
			log.Printf("Nil function? Retry")
			continue
		}
		fstr := function.String()
		function = function.Simplify()
		fstrSimplified := function.String()
		log.Printf("Got function: %s", fstr)
		if fstrSimplified == fstr {
		} else {
			log.Printf("Got simplified function: %s", fstrSimplified)
		}
		depth := function.Depth()
		if depth <= 3 {
			log.Printf("Not deep enough")
			continue
		}
		if depth > 10 {
			log.Printf("Too deep")
			continue
		}
		plotSize := image.Rect(-*size, -*size, *size, *size)
		tUsed, plots := function.Plot(*timeLowerBound, *timeUpperBound, plotSize, *pointSize)
		setCount, usedFrames, frameChanges := 0, 0, 0
		for plotI, plot := range plots {
			setCount += plot.Sets
			if plot.Sets > 0 {
				usedFrames++
			}
			if plotI > 1 && !plot.Equals(plots[plotI-1]) {
				frameChanges++
			}
		}
		if frameChanges <= 1 && len(plots) > 3 {
			log.Printf("Too few frames are different")
			continue
		}
		if usedFrames < len(plots)/2 {
			log.Printf("Too few frames used, less than 50%%")
			continue
		}
		if float64(setCount) < float64(len(plots)*plotSize.Dx()*plotSize.Dy())*0.01 {
			log.Printf("Less than 1%% of all frames used trying again.")
			continue
		}
		if float64(setCount) > float64(len(plots)*plotSize.Dx()*plotSize.Dy())*0.90 {
			log.Printf("More than 90%% of all frames used trying again.")
			continue
		}
		log.Printf("looks good making image")
		heatPlot.RenderPlots(*heatColourCount, plots, plotSize, *scale, function, *timeUpperBound, tUsed, *footerText, *speed, w)
		function.PlotAndDraw(w, *size, *timeLowerBound, *timeUpperBound, *scale, *heatColourCount, *pointSize, *speed, fmt.Sprintf("%s seed: %d", *footerText, seed))
		log.Printf("Done see %s", *outputFile)
		break
	}
}

func randomFunction(d int) *heatPlot.Function {
	return &heatPlot.Function{
		Equals: randomEquals(d),
	}
}

func randomEquals(d int) *heatPlot.Equals {
	return &heatPlot.Equals{
		RHS: randomExpr(d),
		LHS: randomExpr(d),
	}
}

func randomExpr(d int) heatPlot.Expression {
	if d <= 0 {
		return randomVar(0)
	}
	vs := []func(d int) heatPlot.Expression{
		randomConstNumber,
		randomVar,
		randomPlus,
		randomSubtract,
		randomMultiply,
		randomDivide,
		randomPower,
		randomModulus,
		randomNegate,
		randomBrackets,
		randomActualFunction,
	}
	return vs[rng.Intn(len(vs))](d - 1)
}

func randomConstNumber(d int) heatPlot.Expression {
	return &heatPlot.Const{
		Value: float64(rng.Intn(400)) / 4.0,
	}
}

func randomVar(d int) heatPlot.Expression {
	vs := []string{"X", "Y", "T"}
	v := vs[rng.Intn(len(vs))]
	return &heatPlot.Var{
		Var: v,
	}
}

func randomPlus(d int) heatPlot.Expression {
	return &heatPlot.Plus{
		RHS: randomExpr(d),
		LHS: randomExpr(d),
	}
}

func randomSubtract(d int) heatPlot.Expression {
	return &heatPlot.Subtract{
		RHS: randomExpr(d),
		LHS: randomExpr(d),
	}
}

func randomMultiply(d int) heatPlot.Expression {
	return &heatPlot.Multiply{
		RHS: randomExpr(d),
		LHS: randomExpr(d),
	}
}

func randomDivide(d int) heatPlot.Expression {
	return &heatPlot.Divide{
		RHS: randomExpr(d),
		LHS: randomExpr(d),
	}
}

func randomPower(d int) heatPlot.Expression {
	return &heatPlot.Power{
		RHS: randomExpr(d),
		LHS: randomExpr(d),
	}
}

func randomModulus(d int) heatPlot.Expression {
	return &heatPlot.Modulus{
		RHS: randomExpr(d),
		LHS: randomExpr(d),
	}
}

func randomNegate(d int) heatPlot.Expression {
	return &heatPlot.Negate{
		Expr: &heatPlot.Brackets{
			Expr: randomExpr(d),
		},
	}
}

func randomBrackets(d int) heatPlot.Expression {
	return &heatPlot.Brackets{
		Expr: randomExpr(d),
	}
}

func randomActualFunction(d int) heatPlot.Expression {
	functionName := heatPlot.FunctionNames[rng.Intn(len(heatPlot.FunctionNames))]
	if _, ok := heatPlot.SingleFunctions[functionName]; ok {
		return randomSingleFunction(functionName, d)
	}
	return randomDoubleFunction(functionName, d)
}

func randomSingleFunction(name string, d int) heatPlot.Expression {
	return &heatPlot.SingleFunction{
		Name: name,
		Expr: randomExpr(d),
	}
}

func randomDoubleFunction(name string, d int) heatPlot.Expression {
	return &heatPlot.DoubleFunction{
		Expr1: randomExpr(d),
		Expr2: randomExpr(d),
		Infix: rng.Intn(2) == 0,
		Name:  name,
	}
}
