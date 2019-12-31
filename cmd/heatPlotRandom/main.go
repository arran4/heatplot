package main

import (
	"bitbucket.org/arran4/heatplot"
	"flag"
	"log"
	"os"
	"time"
	"math/rand"
)

var (
	heatColourCount = flag.Int("hcc", 126, "Heat colour count.. The number of distinct colours. Can't exceed 254 in total. This value is multiplied by 2. Shouldn't change this.")
	speed           = flag.Duration("speed", 100*time.Millisecond, "The number of microseconds to wait between each frame")
	pixelSize       = flag.Float64("pixelsize", .1, "How many x or y steps a pixel is. Ie .1 will mean that every 10 unscaled pixels is 1 normal step")
	scale           = flag.Int("scale", 2, "Magnification of the picture")
	timeLowerBound  = flag.Int("tlb", 0, "where to start T")
	timeUpperBound  = flag.Int("tub", 100, "Where to end t")
	size            = flag.Int("size", 100, "The size for each direction in the cartesian plane. Ie 100 would be -100 to 100 on the x and y axis")
	outputFile      = flag.String("outputFile", "./out.gif", "The output filename")
	footerText      = flag.String("footerText", "http://github.com/arran4/", "Text to put at the bottom of the picture")
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func main() {
	rand.Seed(time.Now().Unix())
	flag.Parse()
	w, err := os.OpenFile(*outputFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer w.Close()
	function := randomFunction()
	log.Printf("Creating function: %s", function.String())
	heatPlot.RunFunction(function, w, *size, *timeLowerBound, *timeUpperBound, *scale, *heatColourCount, *pixelSize, *speed, *footerText)
	log.Printf("Done see %s", *outputFile)
}

func randomFunction() *heatPlot.Function {
	return &heatPlot.Function {
		Equals: randomEquals(),
	}
}

func randomEquals() *heatPlot.Equals {
	return &heatPlot.Equals {
		RHS: randomExpr(),
		LHS: randomExpr(),
	}
}

func randomExpr() heatPlot.Expression {
	vs := []func () heatPlot.Expression {
		randomConstNumber,
		randomConstNumber,
		randomConstNumber,
		randomConstNumber,
		randomConstNumber,
		randomConstNumber,
		randomConstNumber,
		randomConstNumber,
		randomVar,
		randomVar,
		randomVar,
		randomVar,
		randomPlus,
		randomPlus,
		randomSubtract,
		randomSubtract,
		randomMultiply,
		randomMultiply,
		randomDivide,
		randomDivide,
		randomPower,
		randomPower,
		randomModulus,
		randomNegate,
		randomNegate,
		randomNegate,
		randomBrackets,

	}
	return vs[rand.Intn(len(vs)-1)]()
}

func randomConstNumber() heatPlot.Expression {
	return &heatPlot.Const{
		Value: rand.Float64() * (100),
	}
}

func randomVar() heatPlot.Expression {
	vs := []string{"X","Y","T"}
	v := vs[rand.Intn(len(vs)-1)]
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

func randomSingleFunction(name string) func () heatPlot.Expression {
	return func () heatPlot.Expression {
		return &heatPlot.SingleFunction{
			Name: name,
			Expr: randomExpr(),
		}
	}
}

func randomDoubleFunction(name string) func () heatPlot.Expression {
	return func () heatPlot.Expression {
		return &heatPlot.DoubleFunction{
			Expr1: randomExpr(),
			Expr2: randomExpr(),
			Infix: rand.Intn(2)==0,
			Name: name,
		}
	}
}
