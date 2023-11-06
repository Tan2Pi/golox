package scanner

import (
	"fmt"
	"golox/lox"
	"golox/lox/tokens"
	"os"
	"strconv"
)

type Scanner struct {
	source  string
	tokens  []*tokens.Token
	start   int
	current int
	line    int
}

func New(source string) *Scanner {
	return &Scanner{source: source, tokens: make([]*tokens.Token, 0), start: 0, current: 0, line: 1}
}

func (s *Scanner) ScanTokens() []*tokens.Token {
	defer func() {
		if err := recover(); err != nil {
			lox.HadError = true
			fmt.Fprintln(os.Stderr, err.(error).Error())
		}
	}()
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, &tokens.Token{Type: tokens.Eof, Lexeme: "", Literal: nil, Line: s.line})
	return s.tokens
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) scanToken() {
	c := s.advance()
	switch c {
	case '(':
		s.addToken(tokens.LeftParen)
	case ')':
		s.addToken(tokens.RightParen)
	case '{':
		s.addToken(tokens.LeftBrace)
	case '}':
		s.addToken(tokens.RightBrace)
	case ',':
		s.addToken(tokens.Comma)
	case '.':
		s.addToken(tokens.Dot)
	case '-':
		s.addToken(tokens.Minus)
	case '+':
		s.addToken(tokens.Plus)
	case ';':
		s.addToken(tokens.Semicolon)
	case '*':
		s.addToken(tokens.Star)
	case '!':
		t := mimicTernary(s.match('='), tokens.BangEqual, tokens.Bang)
		s.addToken(t)
	case '=':
		t := mimicTernary(s.match('='), tokens.EqualEqual, tokens.Equal)
		s.addToken(t)
	case '<':
		t := mimicTernary(s.match('='), tokens.LessEqual, tokens.Less)
		s.addToken(t)
	case '>':
		t := mimicTernary(s.match('='), tokens.GreaterEqual, tokens.Greater)
		s.addToken(t)
	case '/':
		if s.match('/') {
			for !s.isAtEnd() && s.peek() != '\n' {
				s.advance()
			}
		} else {
			s.addToken(tokens.Slash)
		}
	case '"':

		s.stringify()
	case ' ', '\r', '\t':
		// Ignore whitespace
	case '\n':
		s.line++
	default:
		if isDigit(c) {
			s.number()
		} else if isAlpha(c) {
			s.identifier()
		} else {
			err := lox.NewLoxError(s.line, "Unexpected character.")
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
}

func (s *Scanner) identifier() {
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}

	text := string([]rune(s.source)[s.start:s.current])
	t, exists := keywords[text]
	if !exists {
		t = tokens.Identifier
	}
	s.addToken(t)
}

func (s *Scanner) number() {
	for isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance()
		for isDigit(s.peek()) {
			s.advance()
		}
	}
	value := string([]rune(s.source)[s.start:s.current])
	num, err := strconv.ParseFloat(value, 32)
	if err != nil {
		err := lox.NewLoxError(s.line, "Invalid Number.")
		fmt.Println(err)
	}
	s.createToken(tokens.Number, num)
}

func (s *Scanner) stringify() {
	for s.peek() != '"' && !s.isAtEnd() {
		//fmt.Printf("%s\n", string(s.peek()))
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}
	//fmt.Printf("%s\n", string(s.peek()))

	if s.isAtEnd() {
		panic(lox.NewLoxError(s.line, "Unterminated string."))
	}

	s.advance()
	// runes := []rune(s.source)
	// value := string(runes[s.start+1 : s.current-1])
	v2 := s.source[s.start+1 : s.current-1]
	//fmt.Printf("v2: %s\n", v2)
	s.createToken(tokens.String, v2)
}

func (s *Scanner) match(expected rune) bool {
	if s.isAtEnd() {
		return false
	}
	if s.currentChar() != expected {
		return false
	}

	s.current++
	return true
}

func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return rune(0)
	}
	return s.currentChar()
}

func (s *Scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return rune(0)
	}
	return []rune(s.source)[s.current+1]
}

func (s *Scanner) advance() rune {
	currentChar := s.currentChar()
	s.current++
	return currentChar
}

func (s *Scanner) addToken(kind tokens.TokenType) {
	s.createToken(kind, nil)
}

func (s *Scanner) createToken(kind tokens.TokenType, literal any) {
	runes := []rune(s.source)
	text := string(runes[s.start:s.current])
	s.tokens = append(s.tokens, &tokens.Token{
		Type:    kind,
		Lexeme:  text,
		Literal: literal,
		Line:    s.line})
}

func isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isAlphaNumeric(c rune) bool {
	return isAlpha(c) || isDigit(c)
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func (s *Scanner) currentChar() rune {
	return rune(s.source[s.current])
}

func mimicTernary(condition bool, valIfTrue tokens.TokenType, valIfFalse tokens.TokenType) tokens.TokenType {
	if condition {
		return valIfTrue
	} else {
		return valIfFalse
	}
}

var keywords = map[string]tokens.TokenType{
	"and":    tokens.And,
	"class":  tokens.Class,
	"else":   tokens.Else,
	"false":  tokens.False,
	"for":    tokens.For,
	"fun":    tokens.Fun,
	"if":     tokens.If,
	"nil":    tokens.Nil,
	"or":     tokens.Or,
	"print":  tokens.Print,
	"return": tokens.Return,
	"super":  tokens.Super,
	"this":   tokens.This,
	"true":   tokens.True,
	"var":    tokens.Var,
	"while":  tokens.While,
}
