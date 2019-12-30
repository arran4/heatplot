yacc:
	go get golang.org/x/tools/cmd/goyacc
	goyacc -o calc.go -v calc.output calc.y
