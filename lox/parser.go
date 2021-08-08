package lox

import (
	"fmt"
	//"strconv"
	"strings"
)

type Parser struct {
	tokens []*Token
	current int
}

func NewParser(tokens []*Token) *Parser {
	return &Parser{tokens, 0}
}

func (p *Parser) comparison() Expr {
	expr := p.term()

	for match(Greater, GreaterEqual, Less, LessEqual) {
		operator := p.previous() 
		right := p.term()
		expr = &Binary{expr, operator, right}
	}

	return expr
}

func (p *Parser) expression() Expr {
	return p.equality()
}

func (p *Parser) equality() Expr {
	expr := p.comparison()

	for p.match(BangEqual, EqualEqual) {
		operator := p.previous()
		right := p.comparison()
		expr = &Binary{expr, operator, right}
	}
	return expr
}

func (p *Parser) factor() Expr {
	expr := p.unary()
	
	for p.match(Slash, Star) {
		operator := p.previous()
		right := p.unary()
		expr = &Binary{expr, operator, right}
	}

	return expr
}

func (p *Parser) primary() Expr {
	if match(False) { return &Literal{false} }
	if match(True) { return &Literal{true} }
	if match(Nil) { return &Literal{nil} }

	if match(Number, String) {
		return &Literal{ p.previous().Literal }
	}

	if match(LeftParen) {
		expr := p.expression()
		p.consume(RightParen, "Expect ')' after expression.")
		return &Grouping{expr}
	}
}

func (p *Parser) term() Expr {
	expr := p.factor()

	for p.match(Minus, Plus) {
		operator := p.previous()
		right := p.factor()
		expr = &Binary{expr, operator, right}
	}

	return expr
}

func (p *Parser) unary() Expr {
	if match(Bang, Minus) {
		operator := p.previous()
		right := p.unary()
		return &Unary{operator, right}
	}

	return p.primary()
}

// helper methods

func (p *Parser) check(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().Type == t
}

func (p *Parser) consume(t TokenType, msg string) *Token, err {
	if p.check(t) { return p.advance(), nil }

	return nil, 
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == Eof
}

func (p *Parser) match(types ...TokenType) bool {
	for t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}

	return false
}

func (p *Parser) peek() *Token {
	return tokens[current]
}

func (p *Parser) previous() *Token {
	return tokens[current-1]
}