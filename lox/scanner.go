package lox

import (
	"fmt"
	"os"
	"strconv"
)

type Scanner struct {
	source  string
	tokens  []*Token
	start   int
	current int
	line    int
}

func NewScanner(source string) *Scanner {
	return &Scanner{source: source, tokens: make([]*Token, 0), start: 0, current: 0, line: 1}
}

func (s *Scanner) ScanTokens() []*Token {
	defer func() {
		if err := recover(); err != nil {
			HadError = true
			fmt.Fprintln(os.Stderr, err.(error).Error())
		}
	}()
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, &Token{Eof, "", nil, s.line})
	return s.tokens
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) scanToken() {
	c := s.advance()
	switch c {
	case '(':
		s.addToken(LeftParen)
	case ')':
		s.addToken(RightParen)
	case '{':
		s.addToken(LeftBrace)
	case '}':
		s.addToken(RightBrace)
	case ',':
		s.addToken(Comma)
	case '.':
		s.addToken(Dot)
	case '-':
		s.addToken(Minus)
	case '+':
		s.addToken(Plus)
	case ';':
		s.addToken(Semicolon)
	case '*':
		s.addToken(Star)
	case '!':
		t := mimicTernary(s.match('='), BangEqual, Bang)
		s.addToken(t)
	case '=':
		t := mimicTernary(s.match('='), EqualEqual, Equal)
		s.addToken(t)
	case '<':
		t := mimicTernary(s.match('='), LessEqual, Less)
		s.addToken(t)
	case '>':
		t := mimicTernary(s.match('='), GreaterEqual, Greater)
		s.addToken(t)
	case '/':
		if s.match('/') {
			for !s.isAtEnd() && s.peek() != '\n' {
				s.advance()
			}
		} else {
			s.addToken(Slash)
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
			err := NewLoxError(s.line, "Unexpected character.")
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
		t = Identifier
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
		err := NewLoxError(s.line, "Invalid Number.")
		fmt.Println(err)
	}
	s.createToken(Number, num)
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
		panic(NewLoxError(s.line, "Unterminated string."))
	}

	s.advance()
	// runes := []rune(s.source)
	// value := string(runes[s.start+1 : s.current-1])
	v2 := s.source[s.start+1 : s.current-1]
	//fmt.Printf("v2: %s\n", v2)
	s.createToken(String, v2)
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

func (s *Scanner) addToken(kind TokenType) {
	s.createToken(kind, nil)
}

func (s *Scanner) createToken(kind TokenType, literal any) {
	runes := []rune(s.source)
	text := string(runes[s.start:s.current])
	s.tokens = append(s.tokens, &Token{kind, text, literal, s.line})
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

func mimicTernary(condition bool, valIfTrue TokenType, valIfFalse TokenType) TokenType {
	if condition {
		return valIfTrue
	} else {
		return valIfFalse
	}
}

var keywords = map[string]TokenType{
	"and":    And,
	"class":  Class,
	"else":   Else,
	"false":  False,
	"for":    For,
	"fun":    Fun,
	"if":     If,
	"nil":    Nil,
	"or":     Or,
	"print":  Print,
	"return": Return,
	"super":  Super,
	"this":   This,
	"true":   True,
	"var":    Var,
	"while":  While,
}
