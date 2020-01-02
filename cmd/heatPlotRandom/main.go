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
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func main() {
	seed := time.Now().UnixNano()
	rand.Seed(seed)
	flag.Parse()
	w, err := os.OpenFile(*outputFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer w.Close()
	for {
		function := randomFunction()
		log.Printf("Got function: %s", function.String())
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

func randomFunction() *heatPlot.Function {
	return &heatPlot.Function{
		Equals: randomEquals(),
	}
}

func randomEquals() *heatPlot.Equals {
	return &heatPlot.Equals{
		RHS: randomExpr(),
		LHS: randomExpr(),
	}
}

func randomExpr() heatPlot.Expression {
	vs := []func() heatPlot.Expression{
		randomConstNumber,
		randomVar,
		//randomPlus,
		//randomSubtract,
		//randomMultiply,
		//randomDivide,
		//randomPower,
		//randomModulus,
		//randomNegate,
		//randomBrackets,
		randomActualFunction,
	}
	return vs[rand.Intn(len(vs))]()
}

func randomConstNumber() heatPlot.Expression {
	return &heatPlot.Const{
		Value: float64(rand.Intn(400)) / 4.0,
	}
}

func randomVar() heatPlot.Expression {
	vs := []string{"X", "Y", "T"}
	v := vs[rand.Intn(len(vs))]
	return &heatPlot.Var{
		Var: v,
	}
}

func randomPlus() heatPlot.Expression {
	return &heatPlot.Plus{
		RHS: randomExpr(),
		LHS: randomExpr(),
	}
}

func randomSubtract() heatPlot.Expression {
	return &heatPlot.Subtract{
		RHS: randomExpr(),
		LHS: randomExpr(),
	}
}

func randomMultiply() heatPlot.Expression {
	return &heatPlot.Multiply{
		RHS: randomExpr(),
		LHS: randomExpr(),
	}
}

func randomDivide() heatPlot.Expression {
	return &heatPlot.Divide{
		RHS: randomExpr(),
		LHS: randomExpr(),
	}
}

func randomPower() heatPlot.Expression {
	return &heatPlot.Power{
		RHS: randomExpr(),
		LHS: randomExpr(),
	}
}

func randomModulus() heatPlot.Expression {
	return &heatPlot.Modulus{
		RHS: randomExpr(),
		LHS: randomExpr(),
	}
}

func randomNegate() heatPlot.Expression {
	return &heatPlot.Negate{
		Expr: randomExpr(),
	}
}

func randomBrackets() heatPlot.Expression {
	return &heatPlot.Negate{
		Expr: randomExpr(),
	}
}

func randomActualFunction() heatPlot.Expression {
	functionName := heatPlot.FunctionNames[rand.Intn(len(heatPlot.FunctionNames))]
	if _, ok := heatPlot.SingleFunctions[functionName]; ok {
		return randomSingleFunction(functionName)
	}
	return randomDoubleFunction(functionName)
}

func randomSingleFunction(name string) heatPlot.Expression {
	return &heatPlot.SingleFunction{
		Name: name,
		Expr: randomExpr(),
	}
}

func randomDoubleFunction(name string) heatPlot.Expression {
	return &heatPlot.DoubleFunction{
		Expr1: randomExpr(),
		Expr2: randomExpr(),
		Infix: rand.Intn(2) == 0,
		Name:  name,
	}
}
