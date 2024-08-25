package main

import (
	"fmt"
	"strings"

	"bufio"
	"os"

	"github.com/Tan2Pi/golox/lox"
	"github.com/Tan2Pi/golox/lox/interpreter"
	"github.com/Tan2Pi/golox/lox/parser"
	"github.com/Tan2Pi/golox/lox/resolver"
	"github.com/Tan2Pi/golox/lox/scanner"
)

const (
	RuntimeErrorCode = 70
	DefaultErrorCode = 65
	UsageErrorCode   = 64

	MaxArgs = 2
)

//nolint:gochecknoglobals // need global interpreter for prompt
var interp = interpreter.NewInterpreter()

func main() {
	switch numArgs := len(os.Args); {
	case numArgs > MaxArgs:
		fmt.Println("usage: golox [script]")
		os.Exit(UsageErrorCode)
	case numArgs == MaxArgs:
		stop := lox.CollectProfile()
		defer stop()
		runFile(os.Args[1])
	default:
		runPrompt()
	}
}

func runFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file: %v.\n Msg: %v", path, err.Error())
		os.Exit(1)
	}

	run(string(data))

	if lox.HadRuntimeError {
		os.Exit(RuntimeErrorCode)
	}

	if lox.HadError {
		os.Exit(DefaultErrorCode)
	}
}

func runPrompt() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "quit" {
			break
		}
		run(text)
		lox.HadError = false
	}
}

func run(source string) {
	scanner := scanner.New(source)
	tokens := scanner.ScanTokens()
	if lox.HadError {
		return
	}

	parser := parser.New(tokens)
	statements := parser.Parse()
	if lox.HadError {
		return
	}

	resolver := resolver.New(interp)
	resolver.Resolve(statements)

	if lox.HadError {
		return
	}

	interp.Interpret(statements)
}
