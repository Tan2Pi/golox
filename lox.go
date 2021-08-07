package main

import (
	"fmt"
	//"io"
	"golox/lox"
	"io/ioutil"
	"os"
	"bufio"
	//"strings"
	//"path/filepath"
)

// func main() {
// 	if len(os.Args) > 1 {
// 		fmt.Println("usage: golox [script]")
// 		os.Exit(64)
// 	} else if len(os.Args) == 2 {
// 		runFile(os.Args[0])
// 	} else {
// 		runPrompt()
// 	}
// }

func main() {
	expression := lox.Binary{
		&lox.Unary{
			&lox.Token{lox.Minus, "-", nil, 1},
			&lox.Literal{123},
		},
		&lox.Token{lox.Star, "*", nil, 1},
		&lox.Grouping{ &lox.Literal{45.67} },
	}
	var p lox.AstPrinter
	fmt.Println(p.Print(&expression))
}

func runFile(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file: %v.\n Msg: %v", path, err.Error())
		os.Exit(1)
	}

	run(string(data))
	if lox.HadError {
		os.Exit(65)
	}
}

func runPrompt() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		//text = strings.TrimSuffix(text, "\n")
		if text == "quit" {
			break
		}
		fmt.Println(text)
		run(text)
		lox.HadError = false
	}
}

func run(source string) {
	scanner := lox.NewScanner(source)
	tokens := scanner.ScanTokens()

	for _, token := range tokens {
		fmt.Println(token)
	}
}