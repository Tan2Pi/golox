package lox

import (
//"fmt"
)

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal interface{}
	Line    int
}

// func (t *Token) String() string {
// 	return fmt.Sprintf("Type: '%s', Lexeme: '%s', Literal: '%s', Line: %d", t.Type, t.Lexeme, t.Literal, t.Line)
// }

// go:generate stringer -type=TokenType
type TokenType int

const (
	LeftParen TokenType = iota
	RightParen
	LeftBrace
	RightBrace
	Comma
	Dot
	Minus
	Plus
	Semicolon
	Slash
	Star

	// 1-2 character tokens
	Bang
	BangEqual
	Equal
	EqualEqual
	Greater
	GreaterEqual
	Less
	LessEqual

	// Literals
	Identifier
	String
	Number

	// Keywords
	And
	Class
	Else
	False
	Fun
	For
	If
	Nil
	Or
	Print
	Return
	Super
	This
	True
	Var
	While

	Eof
)
