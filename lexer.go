package heatPlot

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
	calcLexerRegex, err = regexp.Compile("^(?:(\\s)|([+=,*^/()-])|(\\d+(?:\\.\\d+)?)|([XxYyTt])|(\\w+))")
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
		if len(lex.input) == 0 {
			return 0
		}
		r := lex.subLex(lval)
		if r == -1 {
			continue
		}
		return r
	}
}

func (lex *CalcLexer) subLex(lval *yySymType) int {
	rResult := calcLexerRegex.FindStringSubmatch(lex.input)
	defer func() {
		if rResult == nil || len(rResult) <= 1 || len(rResult[0]) == 0 {
			return
		}
		lex.input = lex.input[len(rResult[0]):]
	}()
	if rResult == nil || len(rResult) <= 1 || len(rResult[0]) == 0 {
		return 1
	}
	if len(rResult[1]) > 0 {
		return -1
	}
	if len(rResult[2]) > 0 {
		return int(rune(rResult[2][0]))
	}
	if len(rResult[3]) > 0 {
		var err error
		lval.float, err = strconv.ParseFloat(rResult[3], 64)
		if err != nil {
			lex.err = err
			return 1
		}
		return FLOAT
	}
	if len(rResult[4]) > 0 {
		lval.s = rResult[4]
		return VAR
	}
	if len(rResult[5]) > 0 {
		lval.s = rResult[5]
		return FUNCNAME
	}
	return 1
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
	return 1
}
