package main

import (
	"errors"
	"log"
	"regexp"
	"strconv"
)

var (
	calcLexerRegex *regexp.Regexp
)

func init() {
	var err error
	calcLexerRegex, err = regexp.Compile("^(?:(\\s)|([+=*^/()-])|(\\d+(?:\\.\\d+)?)|([XxYyTt]))")
	if err != nil {
		log.Panic("Regex compile issue", err)
	}
}

type CalcLexer struct {
	input string
	err   error
}

func NewCalcLexer(input string) yyLexer {
	return &CalcLexer{
		input: input,
	}
}

func (lex *CalcLexer) Lex(lval *yySymType) int {
	for {
		if r := lex.subLex(lval); r == -1 {
			continue
		} else {
			return r
		}
	}
}

func (lex *CalcLexer) subLex(lval *yySymType) int {
	rResult := calcLexerRegex.FindStringSubmatch(lex.input)
	log.Printf("%v %v", lex.input, rResult)
	defer func() {
		if rResult == nil || len(rResult) <= 1 || len(rResult[0]) == 0 {
			return
		}
		lex.input = lex.input[len(rResult[0]):]
	}()
	if rResult == nil || len(rResult) <= 1 || len(rResult[0]) == 0 {
		return 0
	}
	if len(rResult[1]) > 0 {
		return -1
	}
	if len(rResult[2]) > 0 {
		return yyToknameByString("'" + rResult[2] + "'")
	}
	if len(rResult[3]) > 0 {
		var err error
		lval.float, err = strconv.ParseFloat(rResult[3], 64)
		if err != nil {
			lex.err = err
			return 0
		}
		return FLOAT
	}
	if len(rResult[4]) > 0 {
		lval.s = rResult[4]
		return VAR
	}
	return 0
}

func (lex *CalcLexer) Error(s string) {
	lex.err = errors.New(s)
}

func yyToknameByString(s string) int {
	for i, e := range yyToknames {
		if e == s {
			return i
		}
	}
	return -1
}
