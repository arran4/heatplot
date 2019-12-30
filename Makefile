setup:
	go get golang.org/x/tools/cmd/goyacc

yacc:
	goyacc -o calc.go -v calc.output calc.y
