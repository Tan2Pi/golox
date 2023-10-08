package lox

import (
	"fmt"
	"os"
)

var HadError = false

var HadRuntimeError = false

type LoxError struct {
	line    int
	message string
}

type RuntimeError struct {
	Token Token
	Msg   string
}

func ReportRuntimeError(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	HadRuntimeError = true
}

func (e *RuntimeError) Error() string {
	//return reportString(e.Token.Line, "", e.Msg)
	return fmt.Sprintf("%v\n[line %+v]", e.Msg, e.Token.Line)
}

func NewRuntimeError(token Token, msg string) *RuntimeError {
	return &RuntimeError{Token: token, Msg: msg}
}

func LoxErrorHandler(token Token, message string) {
	if token.Type == Eof {
		report(token.Line, "at end", message)
	} else {
		report(token.Line, "at '"+token.Lexeme+"'", message)
	}
}

func NewLoxError(line int, message string) error {
	return &LoxError{line, message}
}

func (e *LoxError) Error() string {
	return fmt.Sprintf("[line %v] Error: %v", e.line, e.message)
}

func report(line int, where string, message string) {
	fmt.Fprintln(os.Stderr, reportString(line, where, message))
	HadError = true
}

func reportString(line int, where string, message string) string {
	return fmt.Sprintf("[line %v] Error %v: %v", line, where, message)
}

type ParseError struct {
	token Token
	msg   string
}

func (e ParseError) Error() string {
	return reportString(e.token.Line, "", e.msg)
}

type ReturnValue struct {
	Value any
}

func NewReturnVal(value any) ReturnValue {
	return ReturnValue{
		Value: value,
	}
}
