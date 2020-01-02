package main

import (
	heatPlot "bitbucket.org/arran4/heatplot"
	"log"
)

func main() {
	log.Printf("Function Names: ")
	for _, fstr := range heatPlot.FunctionNames {
		log.Printf("%s", fstr)
	}
	log.Printf("Single Functions: ")
	for fstr := range heatPlot.SingleFunctions {
		log.Printf("%s", fstr)
	}
	log.Printf("Double Functions: ")
	for fstr := range heatPlot.DoubleFunctions {
		log.Printf("%s", fstr)
	}
}
