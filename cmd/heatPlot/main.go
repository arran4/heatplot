package main

import (
	"flag"
	"heatPlot"
	"log"
	"os"
	"time"
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
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		log.Print("Please include the formula after the command you can use x y and t (t for time) in any way you wish")
		return
	}
	w, err := os.OpenFile(*outputFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer w.Close()
	heatPlot.RunFunction(flag.Arg(0), w, *size, *timeLowerBound, *timeUpperBound, *scale, *heatColourCount, *pixelSize, *speed)
	log.Printf("Done see %s", *outputFile)
}
