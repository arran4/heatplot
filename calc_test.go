package main

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
	} {
		t.Run(fmt.Sprintf("%d: %s", eachI, each), func(t *testing.T) {
			parser := yyNewParser()
			r := parser.Parse(NewCalcLexer(each))
			t.Logf("Result %d", r)
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
