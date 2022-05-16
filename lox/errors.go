package lox

import (
	"fmt"
)

var HadError = false

type LoxError struct {
	line    int
	message string
}

func LoxErrorHandler(token Token, message string) string {
	var err string
	if token.Type == Eof {
		err = report(token.Line, " at end", message)
	} else {
		err = report(token.Line, " at '"+token.Lexeme+"'", message)
	}
	return err
}

func NewLoxError(line int, message string) error {
	return &LoxError{line, message}
}

func (e *LoxError) Error() string {
	err := report(e.line, "", e.message)
	return err
}

func report(line int, where string, message string) string {
	return fmt.Sprintf("[line %v] Error %v: %v", line, where, message)
}

type ParseError struct {
	token Token
	msg   string
}

func (e *ParseError) Error() string {

}
