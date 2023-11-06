package expr

import "golox/lox/tokens"

type Visitor interface {
	VisitBinaryExpr(expr *Binary) any
	VisitGroupingExpr(expr *Grouping) any
	VisitLiteralExpr(expr *Literal) any
	VisitUnaryExpr(expr *Unary) any
	VisitVariableExpr(expr *Variable) any
	VisitAssignExpr(expr *Assign) any
	VisitLogicalExpr(expr *Logical) any
	VisitCallExpr(expr *Call) any
	VisitGetExpr(expr *Get) any
	VisitSetExpr(expr *Set) any
	VisitThisExpr(expr *This) any
	VisitSuperExpr(expr *Super) any
}

type expr struct{}

func (e *expr) Express() Expr {
	return e
}

func (e *expr) Accept(visitor Visitor) any {
	return nil
}

type Expr interface {
	Accept(visitor Visitor) any
	Express() Expr
}

type Binary struct {
	expr
	Left     Expr
	Operator *tokens.Token
	Right    Expr
}

func (expr *Binary) Accept(visitor Visitor) any {
	return visitor.VisitBinaryExpr(expr)
}

type Grouping struct {
	expr
	Expression Expr
}

func (expr *Grouping) Accept(visitor Visitor) any {
	return visitor.VisitGroupingExpr(expr)
}

type Literal struct {
	expr
	Value any
}

func (expr *Literal) Accept(visitor Visitor) any {
	return visitor.VisitLiteralExpr(expr)
}

type Unary struct {
	expr
	Operator *tokens.Token
	Right    Expr
}

func (expr *Unary) Accept(visitor Visitor) any {
	return visitor.VisitUnaryExpr(expr)
}

type Variable struct {
	expr
	Name tokens.Token
}

func (expr *Variable) Accept(visitor Visitor) any {
	return visitor.VisitVariableExpr(expr)
}

type Assign struct {
	expr
	Name  tokens.Token
	Value Expr
}

func (expr *Assign) Accept(visitor Visitor) any {
	return visitor.VisitAssignExpr(expr)
}

type Logical struct {
	expr
	Left     Expr
	Operator tokens.Token
	Right    Expr
}

func (expr *Logical) Accept(v Visitor) any {
	return v.VisitLogicalExpr(expr)
}

type Call struct {
	expr
	Callee Expr
	Paren  tokens.Token
	Args   []Expr
}

func (expr *Call) Accept(v Visitor) any {
	return v.VisitCallExpr(expr)
}

type Get struct {
	expr
	Object Expr
	Name   tokens.Token
}

func (expr *Get) Accept(v Visitor) any {
	return v.VisitGetExpr(expr)
}

type Set struct {
	expr
	Object Expr
	Name   tokens.Token
	Value  Expr
}

func (expr *Set) Accept(v Visitor) any {
	return v.VisitSetExpr(expr)
}

type This struct {
	expr
	Keyword tokens.Token
}

func (expr *This) Accept(v Visitor) any {
	return v.VisitThisExpr(expr)
}

type Super struct {
	expr
	Keyword tokens.Token
	Method  tokens.Token
}

func (expr *Super) Accept(v Visitor) any {
	return v.VisitSuperExpr(expr)
}
