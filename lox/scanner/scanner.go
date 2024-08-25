package scanner

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Tan2Pi/golox/lox"
	"github.com/Tan2Pi/golox/lox/tokens"
)

type Scanner struct {
	source  string
	runes   []rune
	tokens  []*tokens.Token
	start   int
	current int
	line    int
}

func New(source string) *Scanner {
	return &Scanner{
		source:  source,
		runes:   []rune(source),
		tokens:  make([]*tokens.Token, 0),
		start:   0,
		current: 0,
		line:    1}
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

	s.tokens = append(s.tokens, &tokens.Token{Type: tokens.EOF, Lexeme: "", Literal: nil, Line: s.line})
	return s.tokens
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.runes)
}

func (s *Scanner) scanToken() {
	c := s.advance()
	lox.Debug("c = %#v, start = %v, current = %v\n", string(c), s.start, s.current)
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
		s.commentsOrSlash()
	case '"':
		s.stringify()
	case ' ', '\r', '\t':
		// Ignore whitespace
		break
	case '\n':
		s.line++
	default:
		s.scanDefault(c)
	}
}

func (s *Scanner) scanDefault(c rune) {
	switch {
	case isDigit(c):
		s.number()
	case isAlpha(c):
		s.identifier()
	default:
		err := lox.NewError(s.line, "Unexpected character.")
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

func (s *Scanner) commentsOrSlash() {
	if s.match('/') {
		for s.peek() != '\n' && !s.isAtEnd() {
			s.advance()
		}
		lox.Debug("peek() = %#v, start = %v, current = %v\n", string(s.peek()), s.start, s.current)
	} else {
		s.addToken(tokens.Slash)
	}
}

func (s *Scanner) identifier() {
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}

	text := string(s.runes[s.start:s.current])
	lox.Debug("text = %#v, start = %v, current = %v\n", text, s.start, s.current)
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
	value := string(s.runes[s.start:s.current])
	num, err := strconv.ParseFloat(value, 32)
	if err != nil {
		fmt.Println(lox.NewError(s.line, "Invalid Number."))
	}
	s.createToken(tokens.Number, num)
}

func (s *Scanner) stringify() {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		panic(lox.NewError(s.line, "Unterminated string."))
	}

	s.advance()
	value := string(s.runes[s.start+1 : s.current-1])
	s.createToken(tokens.String, value)
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
	if s.current+1 >= len(s.runes) {
		return rune(0)
	}
	return s.runes[s.current+1]
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
	text := string(s.runes[s.start:s.current])
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
	return s.runes[s.current]
}

func mimicTernary(condition bool, valIfTrue tokens.TokenType, valIfFalse tokens.TokenType) tokens.TokenType {
	if condition {
		return valIfTrue
	}

	return valIfFalse
}

//nolint:gochecknoglobals // this is effectively a static variable
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
