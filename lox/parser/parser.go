package parser

import (
	"errors"
	"fmt"
	"golox/lox"
	"golox/lox/expr"
	"golox/lox/stmt"
	"golox/lox/tokens"
)

type Parser struct {
	tokens   []*tokens.Token
	current  int
	hadError bool
}

func New(tokens []*tokens.Token) *Parser {
	return &Parser{tokens, 0, false}
}

func (p *Parser) Parse() []stmt.Stmt {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Println(r)
	// 	}
	// }()

	statements := make([]stmt.Stmt, 0)
	for !p.isAtEnd() {
		statements = append(statements, p.declaration())
	}
	return statements
}

func (p *Parser) declaration() stmt.Stmt {
	defer func() {
		if r := recover(); r != nil {
			switch val := r.(type) {
			case error:
				if errors.As(val, &lox.ParseError{}) {
					p.synchronize()
				}
			default:
				fmt.Printf("unknown error %+v", r)
			}
		}
	}()
	if p.match(tokens.Class) {
		return p.classDeclaration()
	}
	if p.match(tokens.Fun) {
		return p.function("function")
	}
	if p.match(tokens.Var) {
		return p.varDeclaration()
	}
	return p.statement()
}

func (p *Parser) classDeclaration() stmt.Stmt {
	name := p.consume(tokens.Identifier, "Expect class name.")
	var superClass *expr.Variable
	if p.match(tokens.Less) {
		p.consume(tokens.Identifier, "Expect superclass name.")
		superClass = &expr.Variable{
			Name: *p.previous(),
		}
	}
	p.consume(tokens.LeftBrace, "Expect '{' before class body.")
	methods := make([]*stmt.Function, 0)
	for !p.check(tokens.RightBrace) && !p.isAtEnd() {
		methods = append(methods, p.function("method"))
	}

	p.consume(tokens.RightBrace, "Expect '}' after class body.")
	return &stmt.Class{
		Name:       *name,
		Superclass: superClass,
		Methods:    methods,
	}
}

func (p *Parser) varDeclaration() stmt.Stmt {
	name := p.consume(tokens.Identifier, "Expect variable name.")

	var initializer expr.Expr
	if p.match(tokens.Equal) {
		initializer = p.expression()
	}

	p.consume(tokens.Semicolon, "Expect ';' after variable declaration.")
	return &stmt.Variable{
		Name:        *name,
		Initializer: initializer,
	}
}

func (p *Parser) whileStatement() stmt.Stmt {
	p.consume(tokens.LeftParen, "Expect '(' after 'while'.")
	condition := p.expression()
	p.consume(tokens.RightParen, "Expect ')' after condition.")
	body := p.statement()

	return &stmt.While{
		Condition: condition,
		Body:      body,
	}
}

func (p *Parser) comparison() expr.Expr {
	e := p.term()

	for p.match(tokens.Greater, tokens.GreaterEqual, tokens.Less, tokens.LessEqual) {
		operator := p.previous()
		right := p.term()
		e = &expr.Binary{Left: e, Operator: operator, Right: right}
	}

	return e
}

func (p *Parser) expression() expr.Expr {
	return p.assignment()
}

func (p *Parser) statement() stmt.Stmt {
	if p.match(tokens.For) {
		return p.forStatement()
	}
	if p.match(tokens.If) {
		return p.ifStatement()
	}
	if p.match(tokens.Print) {
		return p.printStatement()
	}
	if p.match(tokens.Return) {
		return p.returnStatement()
	}
	if p.match(tokens.While) {
		return p.whileStatement()
	}

	if p.match(tokens.LeftBrace) {
		return &stmt.Block{
			Statements: p.block(),
		}
	}

	return p.expressionStatement()
}

func (p *Parser) forStatement() stmt.Stmt {
	p.consume(tokens.LeftParen, "Expect '(' after 'for'.")

	var initializer stmt.Stmt
	if p.match(tokens.Semicolon) {
		initializer = nil
	} else if p.match(tokens.Var) {
		initializer = p.varDeclaration()
	} else {
		initializer = p.expressionStatement()
	}

	var condition expr.Expr
	if !p.check(tokens.Semicolon) {
		condition = p.expression()
	}
	p.consume(tokens.Semicolon, "Expect ';' after loop condition.")

	var increment expr.Expr
	if !p.check(tokens.RightParen) {
		increment = p.expression()
	}
	p.consume(tokens.RightParen, "Expect ')' after for clauses.")

	body := p.statement()

	if increment != nil {
		body = &stmt.Block{
			Statements: []stmt.Stmt{
				body,
				&stmt.Expr{
					Expression: increment,
				},
			},
		}
	}

	if condition == nil {
		condition = &expr.Literal{
			Value: true,
		}
	}

	body = &stmt.While{
		Condition: condition,
		Body:      body,
	}

	if initializer != nil {
		body = &stmt.Block{
			Statements: []stmt.Stmt{
				initializer,
				body,
			},
		}
	}

	return body
}

func (p *Parser) ifStatement() stmt.Stmt {
	p.consume(tokens.LeftParen, "Expect '(' after 'if'.")
	condition := p.expression()
	p.consume(tokens.RightParen, "Expect ')' after 'if'.")
	thenBranch := p.statement()
	var elseBranch stmt.Stmt
	if p.match(tokens.Else) {
		elseBranch = p.statement()
	}

	return &stmt.If{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}
}

func (p *Parser) printStatement() stmt.Stmt {
	value := p.expression()
	p.consume(tokens.Semicolon, "Expect ';' after value.")
	return &stmt.Print{Expression: value}
}

func (p *Parser) returnStatement() stmt.Stmt {
	keyword := p.previous()
	var value expr.Expr
	if !p.check(tokens.Semicolon) {
		value = p.expression()
	}

	p.consume(tokens.Semicolon, "Expect ';' after return value.")
	return &stmt.Return{
		Keyword: *keyword,
		Value:   value,
	}
}

func (p *Parser) expressionStatement() stmt.Stmt {
	expr := p.expression()
	p.consume(tokens.Semicolon, "Expect ';' after expression.")
	return &stmt.Expr{Expression: expr}
}

func (p *Parser) function(kind string) *stmt.Function {
	name := p.consume(tokens.Identifier, fmt.Sprintf("Expect %s name.", kind))
	p.consume(tokens.LeftParen, fmt.Sprintf("Expect '(' after %s name.", kind))
	params := make([]tokens.Token, 0)
	if !p.check(tokens.RightParen) {
		for {
			if len(params) >= 255 {
				p.error(*p.peek(), "Can't have more than 255 parameters.")
			}

			params = append(params, *p.consume(tokens.Identifier, "Expect parameter name."))

			if !p.match(tokens.Comma) {
				break
			}
		}
	}

	p.consume(tokens.RightParen, "Expect ')' after parameters.")
	p.consume(tokens.LeftBrace, fmt.Sprintf("Expect '{' before %s body.", kind))
	body := p.block()
	return &stmt.Function{
		Name:   *name,
		Params: params,
		Body:   body,
	}
}

func (p *Parser) block() []stmt.Stmt {
	statements := make([]stmt.Stmt, 0)
	for !p.check(tokens.RightBrace) && !p.isAtEnd() {
		statements = append(statements, p.declaration())
	}

	p.consume(tokens.RightBrace, "Expect '}' after block.")
	return statements
}

func (p *Parser) assignment() expr.Expr {
	e := p.or()

	if p.match(tokens.Equal) {
		equal := p.previous()
		value := p.assignment()
		switch v := e.(type) {
		case *expr.Variable:
			return &expr.Assign{
				Name:  v.Name,
				Value: value,
			}
		case *expr.Get:
			get := e.(*expr.Get)
			return &expr.Set{
				Object: get.Object,
				Name:   get.Name,
				Value:  value,
			}
		default:
			p.error(*equal, "Invalid assignment target.")
		}
	}
	return e
}

func (p *Parser) or() expr.Expr {
	e := p.and()

	for p.match(tokens.Or) {
		operator := p.previous()
		right := p.and()
		e = &expr.Logical{
			Left:     e,
			Operator: *operator,
			Right:    right,
		}
	}

	return e
}

func (p *Parser) and() expr.Expr {
	e := p.equality()
	for p.match(tokens.And) {
		operator := p.previous()
		right := p.equality()
		e = &expr.Logical{
			Left:     e,
			Operator: *operator,
			Right:    right,
		}
	}

	return e
}

func (p *Parser) equality() expr.Expr {
	e := p.comparison()

	for p.match(tokens.BangEqual, tokens.EqualEqual) {
		operator := p.previous()
		right := p.comparison()
		e = &expr.Binary{Left: e, Operator: operator, Right: right}
	}
	return e
}

func (p *Parser) factor() expr.Expr {
	e := p.unary()

	for p.match(tokens.Slash, tokens.Star) {
		operator := p.previous()
		right := p.unary()
		e = &expr.Binary{Left: e, Operator: operator, Right: right}
	}

	return e
}

func (p *Parser) primary() expr.Expr {
	if p.match(tokens.False) {
		return &expr.Literal{Value: false}
	}
	if p.match(tokens.True) {
		return &expr.Literal{Value: true}
	}
	if p.match(tokens.Nil) {
		return &expr.Literal{Value: nil}
	}

	if p.match(tokens.Number, tokens.String) {
		return &expr.Literal{Value: p.previous().Literal}
	}

	if p.match(tokens.Super) {
		keyword := p.previous()
		p.consume(tokens.Dot, "Expect '.' after 'super'.")
		method := p.consume(tokens.Identifier, "Expect superclass method name.")
		return &expr.Super{
			Keyword: *keyword,
			Method:  *method,
		}
	}

	if p.match(tokens.This) {
		return &expr.This{Keyword: *p.previous()}
	}

	if p.match(tokens.Identifier) {
		return &expr.Variable{
			Name: *p.previous(),
		}
	}

	if p.match(tokens.LeftParen) {
		e := p.expression()
		p.consume(tokens.RightParen, "Expect ')' after expression.")
		return &expr.Grouping{Expression: e}
	}

	p.error(*p.peek(), "Expect expression.")
	return nil
}

func (p *Parser) term() expr.Expr {
	e := p.factor()

	for p.match(tokens.Minus, tokens.Plus) {
		operator := p.previous()
		right := p.factor()
		e = &expr.Binary{Left: e, Operator: operator, Right: right}
	}

	return e
}

func (p *Parser) unary() expr.Expr {
	if p.match(tokens.Bang, tokens.Minus) {
		operator := p.previous()
		right := p.unary()
		return &expr.Unary{Operator: operator, Right: right}
	}

	return p.call()
}

func (p *Parser) call() expr.Expr {
	e := p.primary()

	for {
		if p.match(tokens.LeftParen) {
			e = p.finishCall(e)
		} else if p.match(tokens.Dot) {
			name := p.consume(tokens.Identifier, "Expect property name after '.'.")
			e = &expr.Get{
				Object: e,
				Name:   *name,
			}
		} else {
			break
		}
	}

	return e
}

func (p *Parser) finishCall(callee expr.Expr) expr.Expr {
	args := make([]expr.Expr, 0)

	if !p.check(tokens.RightParen) {
		for {
			if len(args) >= 255 {
				p.error(*p.peek(), "Can't have more than 255 arguments.")
			}
			args = append(args, p.expression())

			// run loop at least once
			if !p.match(tokens.Comma) {
				break
			}
		}
	}

	paren := p.consume(tokens.RightParen, "Expect ')' after arguments.")

	return &expr.Call{
		Callee: callee,
		Paren:  *paren,
		Args:   args,
	}
}

// helper methods

func (p *Parser) check(t tokens.TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().Type == t
}

func (p *Parser) consume(t tokens.TokenType, msg string) *tokens.Token {
	if p.check(t) {
		return p.advance()
	}
	p.error(*p.peek(), msg)
	return nil
}

func (p *Parser) error(t tokens.Token, msg string) {
	lox.LoxErrorHandler(t, msg)
	p.hadError = false
	panic(lox.ParseError{Token: t, Msg: msg})
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == tokens.Eof
}

func (p *Parser) match(types ...tokens.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}

	return false
}

func (p *Parser) peek() *tokens.Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() *tokens.Token {
	return p.tokens[p.current-1]
}

func (p *Parser) advance() *tokens.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == tokens.Semicolon {
			return
		}
		switch p.peek().Type {
		case tokens.Class:
			fallthrough
		case tokens.Fun:
			fallthrough
		case tokens.Var:
			fallthrough
		case tokens.For:
			fallthrough
		case tokens.If:
			fallthrough
		case tokens.While:
			fallthrough
		case tokens.Print:
			fallthrough
		case tokens.Return:
			return
		// do nothing
		default:
		}

		p.advance()
	}
}
