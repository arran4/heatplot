setup:
	go install golang.org/x/tools/cmd/goyacc@latest

yacc:
	goyacc -o calc.go -v calc.output calc.y
