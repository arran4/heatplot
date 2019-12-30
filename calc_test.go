package main

import (
	"fmt"
	"testing"
)

func TestEndToEndParser(t *testing.T) {
	for eachI, each := range []string{
		"y = x + 2",
	} {
		t.Run(fmt.Sprintf("%d: %s", eachI, each), func(t *testing.T) {
			r := yyParse(NewCalcLexer(each))
			t.Logf("Result %d", r)
			if yyResult == nil {
				t.Logf("Error; no result returned")
				t.Fail()
			} else if yyResult.String() != each {
				t.Logf("Failed to match %v with %v", yyResult.String(), each)
				t.Fail()
			}
		})
	}
}
