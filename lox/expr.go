package lox

type ExprVisitor interface {
	VisitBinaryExpr(expr *Binary) any
	VisitGroupingExpr(expr *Grouping) any
	VisitLiteralExpr(expr *Literal) any
	VisitUnaryExpr(expr *Unary) any
	VisitVariableExpr(expr *Variable) any
	VisitAssignExpr(expr *Assign) any
	VisitLogicalExpr(expr *Logical) any
	VisitCallExpr(expr *Call) any
	VisitGetExpr(expr *GetExpr) any
	VisitSetExpr(expr *SetExpr) any
	VisitThisExpr(expr *ThisExpr) any
}

type expr struct{}

func (e *expr) Express() Expr {
	return e
}

func (e *expr) Accept(visitor ExprVisitor) any {
	return nil
}

type Expr interface {
	Accept(visitor ExprVisitor) any
	Express() Expr
}

type Binary struct {
	expr
	Left     Expr
	Operator *Token
	Right    Expr
}

func (expr *Binary) Accept(visitor ExprVisitor) any {
	return visitor.VisitBinaryExpr(expr)
}

type Grouping struct {
	expr
	Expression Expr
}

func (expr *Grouping) Accept(visitor ExprVisitor) any {
	return visitor.VisitGroupingExpr(expr)
}

type Literal struct {
	expr
	Value any
}

func (expr *Literal) Accept(visitor ExprVisitor) any {
	return visitor.VisitLiteralExpr(expr)
}

type Unary struct {
	expr
	Operator *Token
	Right    Expr
}

func (expr *Unary) Accept(visitor ExprVisitor) any {
	return visitor.VisitUnaryExpr(expr)
}

type Variable struct {
	expr
	Name Token
}

func (expr *Variable) Accept(visitor ExprVisitor) any {
	return visitor.VisitVariableExpr(expr)
}

type Assign struct {
	expr
	Name  Token
	Value Expr
}

func (expr *Assign) Accept(visitor ExprVisitor) any {
	return visitor.VisitAssignExpr(expr)
}

type Logical struct {
	expr
	Left     Expr
	Operator Token
	Right    Expr
}

func (expr *Logical) Accept(v ExprVisitor) any {
	return v.VisitLogicalExpr(expr)
}

type Call struct {
	expr
	Callee Expr
	Paren  Token
	Args   []Expr
}

func (expr *Call) Accept(v ExprVisitor) any {
	return v.VisitCallExpr(expr)
}

type GetExpr struct {
	expr
	Object Expr
	Name   Token
}

func (expr *GetExpr) Accept(v ExprVisitor) any {
	return v.VisitGetExpr(expr)
}

type SetExpr struct {
	expr
	Object Expr
	Name   Token
	Value  Expr
}

func (expr *SetExpr) Accept(v ExprVisitor) any {
	return v.VisitSetExpr(expr)
}

type ThisExpr struct {
	expr
	Keyword Token
}

func (expr *ThisExpr) Accept(v ExprVisitor) any {
	return v.VisitThisExpr(expr)
}
