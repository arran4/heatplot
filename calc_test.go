package heatPlot

import (
	"fmt"
	"testing"
)

func init() {
	//yyDebug = 3
}

func TestEndToEndParser(t *testing.T) {
	for eachI, each := range []string{
		"y / 4 = x + 2",
		"y / 4 = x * x + 2",
		"y / 4 = x + x * 2",
		"y / 4 = x * (x + 2)",
		"y / 4 = (x + x) * 2",
		"y / 4 = x mod 2 * 2",
		"y / 4 = mod(x, 3) * 2",
		"y / 4 = abs(x) + 2 * 2",
	} {
		t.Run(fmt.Sprintf("%d: %s", eachI, each), func(t *testing.T) {
			parser := yyNewParser()
			yyResult = nil
			r := parser.Parse(NewCalcLexer(each))
			t.Logf("Result %d for %#v", r, each)
			if yyResult == nil {
				t.Logf("Error; no result returned %#v", parser)
				t.Fail()
			} else if yyResult.String() != each {
				t.Logf("Failed to match %v with %v", yyResult.String(), each)
				t.Fail()
			}
		})
	}
}
