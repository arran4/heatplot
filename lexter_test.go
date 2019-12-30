package main

import (
	"fmt"
	"testing"
)

func TestLexer(t *testing.T) {
	for eachI, each := range []struct {
		Input  string
		Output []int
		Name   string
	}{
		{
			Name:   "Run 1",
			Input:  "1 + 2 * (3 / 4)",
			Output: []int{FLOAT, yyToknameByString("'+'"), FLOAT, yyToknameByString("'*'"), yyToknameByString("'('"), FLOAT, yyToknameByString("'/'"), FLOAT, yyToknameByString("')'"), yyToknameByString("$end")},
		},
	} {
		t.Run(fmt.Sprintf("Test %d", eachI), func(t *testing.T) {
			yyLexer := NewCalcLexer(each.Input)
			for i, e := range each.Output {
				if a := yyLexer.Lex(&yySymType{}); a != e {
					t.Logf("Token %d failed, got %d instead expected %d", i, a, e)
					t.Fail()
				} else {
					t.Logf("Token %d, got %d", i, a)
				}
			}
		})
	}
}
