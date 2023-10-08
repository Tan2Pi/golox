package main

import (
	"fmt"
	_ "net/http/pprof"
	"strings"

	"bufio"
	"golox/lox"
	"os"
)

var interpreter *lox.Interpreter = lox.NewInterpreter()

func main() {
	if len(os.Args) > 2 {
		fmt.Println("usage: golox [script]")
		os.Exit(64)
	} else if len(os.Args) == 2 {
		runFile(os.Args[1])
	} else {
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
		os.Exit(70)
	}

	if lox.HadError {
		os.Exit(65)
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
	scanner := lox.NewScanner(source)
	tokens := scanner.ScanTokens()
	if lox.HadError {
		return
	}

	parser := lox.NewParser(tokens)
	statements := parser.Parse()
	if lox.HadError {
		return
	}

	resolver := lox.NewResolver(interpreter)
	resolver.Resolve(statements)

	if lox.HadError {
		return
	}

	interpreter.Interpret(statements)
}
