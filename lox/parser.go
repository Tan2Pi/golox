package lox

import (
	"errors"
	"fmt"
)

type Parser struct {
	tokens   []*Token
	current  int
	hadError bool
}

func NewParser(tokens []*Token) *Parser {
	return &Parser{tokens, 0, false}
}

func (p *Parser) Parse() []Stmt {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Println(r)
	// 	}
	// }()

	statements := make([]Stmt, 0)
	for !p.isAtEnd() {
		statements = append(statements, p.declaration())
	}
	return statements
}

func (p *Parser) declaration() Stmt {
	defer func() {
		if r := recover(); r != nil {
			switch val := r.(type) {
			case error:
				if errors.As(val, &ParseError{}) {
					p.synchronize()
				}
			default:
				fmt.Printf("unknown error %+v", r)
			}
		}
	}()
	if p.match(Class) {
		return p.classDeclaration()
	}
	if p.match(Fun) {
		return p.function("function")
	}
	if p.match(Var) {
		return p.varDeclaration()
	}
	return p.statement()
}

func (p *Parser) classDeclaration() Stmt {
	name := p.consume(Identifier, "Expect class name.")
	p.consume(LeftBrace, "Expect '{' before class body.")
	methods := make([]*FunctionStmt, 0)
	for !p.check(RightBrace) && !p.isAtEnd() {
		methods = append(methods, p.function("method"))
	}

	p.consume(RightBrace, "Expect '}' after class body.")
	return &ClassStmt{
		Name:    *name,
		Methods: methods,
	}
}

func (p *Parser) varDeclaration() Stmt {
	name := p.consume(Identifier, "Expect variable name.")

	var initializer Expr
	if p.match(Equal) {
		initializer = p.expression()
	}

	p.consume(Semicolon, "Expect ';' after variable declaration.")
	return &VariableStmt{
		Name:        *name,
		Initializer: initializer,
	}
}

func (p *Parser) whileStatement() Stmt {
	p.consume(LeftParen, "Expect '(' after 'while'.")
	condition := p.expression()
	p.consume(RightParen, "Expect ')' after condition.")
	body := p.statement()

	return &WhileStmt{
		Condition: condition,
		Body:      body,
	}
}

func (p *Parser) comparison() Expr {
	expr := p.term()

	for p.match(Greater, GreaterEqual, Less, LessEqual) {
		operator := p.previous()
		right := p.term()
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) expression() Expr {
	return p.assignment()
}

func (p *Parser) statement() Stmt {
	if p.match(For) {
		return p.forStatement()
	}
	if p.match(If) {
		return p.ifStatement()
	}
	if p.match(Print) {
		return p.printStatement()
	}
	if p.match(Return) {
		return p.returnStatement()
	}
	if p.match(While) {
		return p.whileStatement()
	}

	if p.match(LeftBrace) {
		return &BlockStmt{
			Statements: p.block(),
		}
	}

	return p.expressionStatement()
}

func (p *Parser) forStatement() Stmt {
	p.consume(LeftParen, "Expect '(' after 'for'.")

	var initializer Stmt
	if p.match(Semicolon) {
		initializer = nil
	} else if p.match(Var) {
		initializer = p.varDeclaration()
	} else {
		initializer = p.expressionStatement()
	}

	var condition Expr
	if !p.check(Semicolon) {
		condition = p.expression()
	}
	p.consume(Semicolon, "Expect ';' after loop condition.")

	var increment Expr
	if !p.check(RightParen) {
		increment = p.expression()
	}
	p.consume(RightParen, "Expect ')' after for clauses.")

	body := p.statement()

	if increment != nil {
		body = &BlockStmt{
			Statements: []Stmt{
				body,
				&StmtExpr{
					Expression: increment,
				},
			},
		}
	}

	if condition == nil {
		condition = &Literal{
			Value: true,
		}
	}

	body = &WhileStmt{
		Condition: condition,
		Body:      body,
	}

	if initializer != nil {
		body = &BlockStmt{
			Statements: []Stmt{
				initializer,
				body,
			},
		}
	}

	return body
}

func (p *Parser) ifStatement() Stmt {
	p.consume(LeftParen, "Expect '(' after 'if'.")
	condition := p.expression()
	p.consume(RightParen, "Expect ')' after 'if'.")
	thenBranch := p.statement()
	var elseBranch Stmt
	if p.match(Else) {
		elseBranch = p.statement()
	}

	return &IfStmt{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}
}

func (p *Parser) printStatement() Stmt {
	value := p.expression()
	p.consume(Semicolon, "Expect ';' after value.")
	return &StmtPrint{Expression: value}
}

func (p *Parser) returnStatement() Stmt {
	keyword := p.previous()
	var value Expr
	if !p.check(Semicolon) {
		value = p.expression()
	}

	p.consume(Semicolon, "Expect ';' after return value.")
	return &ReturnStmt{
		Keyword: *keyword,
		Value:   value,
	}
}

func (p *Parser) expressionStatement() Stmt {
	expr := p.expression()
	p.consume(Semicolon, "Expect ';' after expression.")
	return &StmtExpr{Expression: expr}
}

func (p *Parser) function(kind string) *FunctionStmt {
	name := p.consume(Identifier, fmt.Sprintf("Expect %s name.", kind))
	p.consume(LeftParen, fmt.Sprintf("Expect '(' after %s name.", kind))
	params := make([]Token, 0)
	if !p.check(RightParen) {
		for {
			if len(params) >= 255 {
				p.error(*p.peek(), "Can't have more than 255 parameters.")
			}

			params = append(params, *p.consume(Identifier, "Expect parameter name."))

			if !p.match(Comma) {
				break
			}
		}
	}

	p.consume(RightParen, "Expect ')' after parameters.")
	p.consume(LeftBrace, fmt.Sprintf("Expect '{' before %s body.", kind))
	body := p.block()
	return &FunctionStmt{
		Name:   *name,
		Params: params,
		Body:   body,
	}
}

func (p *Parser) block() []Stmt {
	statements := make([]Stmt, 0)
	for !p.check(RightBrace) && !p.isAtEnd() {
		statements = append(statements, p.declaration())
	}

	p.consume(RightBrace, "Expect '}' after block.")
	return statements
}

func (p *Parser) assignment() Expr {
	expr := p.or()

	if p.match(Equal) {
		equal := p.previous()
		value := p.assignment()
		switch v := expr.(type) {
		case *Variable:
			return &Assign{
				Name:  v.Name,
				Value: value,
			}
		case *GetExpr:
			get := expr.(*GetExpr)
			return &SetExpr{
				Object: get.Object,
				Name:   get.Name,
				Value:  value,
			}
		default:
			p.error(*equal, "Invalid assignment target.")
		}
	}
	return expr
}

func (p *Parser) or() Expr {
	expr := p.and()

	for p.match(Or) {
		operator := p.previous()
		right := p.and()
		expr = &Logical{
			Left:     expr,
			Operator: *operator,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) and() Expr {
	expr := p.equality()
	for p.match(And) {
		operator := p.previous()
		right := p.equality()
		expr = &Logical{
			Left:     expr,
			Operator: *operator,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) equality() Expr {
	expr := p.comparison()

	for p.match(BangEqual, EqualEqual) {
		operator := p.previous()
		right := p.comparison()
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr
}

func (p *Parser) factor() Expr {
	expr := p.unary()

	for p.match(Slash, Star) {
		operator := p.previous()
		right := p.unary()
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) primary() Expr {
	if p.match(False) {
		return &Literal{Value: false}
	}
	if p.match(True) {
		return &Literal{Value: true}
	}
	if p.match(Nil) {
		return &Literal{Value: nil}
	}

	if p.match(Number, String) {
		return &Literal{Value: p.previous().Literal}
	}

	if p.match(This) {
		return &ThisExpr{Keyword: *p.previous()}
	}

	if p.match(Identifier) {
		return &Variable{
			Name: *p.previous(),
		}
	}

	if p.match(LeftParen) {
		expr := p.expression()
		p.consume(RightParen, "Expect ')' after expression.")
		return &Grouping{Expression: expr}
	}

	p.error(*p.peek(), "Expect expression.")
	return nil
}

func (p *Parser) term() Expr {
	expr := p.factor()

	for p.match(Minus, Plus) {
		operator := p.previous()
		right := p.factor()
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) unary() Expr {
	if p.match(Bang, Minus) {
		operator := p.previous()
		right := p.unary()
		return &Unary{Operator: operator, Right: right}
	}

	return p.call()
}

func (p *Parser) call() Expr {
	expr := p.primary()

	for {
		if p.match(LeftParen) {
			expr = p.finishCall(expr)
		} else if p.match(Dot) {
			name := p.consume(Identifier, "Expect property name after '.'.")
			expr = &GetExpr{
				Object: expr,
				Name:   *name,
			}
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) finishCall(callee Expr) Expr {
	args := make([]Expr, 0)

	if !p.check(RightParen) {
		for {
			if len(args) >= 255 {
				p.error(*p.peek(), "Can't have more than 255 arguments.")
			}
			args = append(args, p.expression())

			// run loop at least once
			if !p.match(Comma) {
				break
			}
		}
	}

	paren := p.consume(RightParen, "Expect ')' after arguments.")

	return &Call{
		Callee: callee,
		Paren:  *paren,
		Args:   args,
	}
}

// helper methods

func (p *Parser) check(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().Type == t
}

func (p *Parser) consume(t TokenType, msg string) *Token {
	if p.check(t) {
		return p.advance()
	}
	p.error(*p.peek(), msg)
	return nil
}

func (p *Parser) error(t Token, msg string) {
	LoxErrorHandler(t, msg)
	p.hadError = false
	panic(ParseError{t, msg})
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == Eof
}

func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}

	return false
}

func (p *Parser) peek() *Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() *Token {
	return p.tokens[p.current-1]
}

func (p *Parser) advance() *Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == Semicolon {
			return
		}
		switch p.peek().Type {
		case Class:
			fallthrough
		case Fun:
			fallthrough
		case Var:
			fallthrough
		case For:
			fallthrough
		case If:
			fallthrough
		case While:
			fallthrough
		case Print:
			fallthrough
		case Return:
			return
		// do nothing
		default:
		}

		p.advance()
	}
}
