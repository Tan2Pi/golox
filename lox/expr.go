package lox

import (
	//"fmt"
	//"strconv"
)

type ExprVisitor interface {
	VisitBinaryExpr(expr *Binary) interface{}
	VisitGroupingExpr(expr *Grouping) interface{}
	VisitLiteralExpr(expr *Literal) interface{}
	VisitUnaryExpr(expr *Unary) interface{}
}

type Expr interface {
	Accept(visitor ExprVisitor) interface{}
}

type Binary struct {
	Left Expr
	Operator *Token
	Right Expr
}

func (expr *Binary) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitBinaryExpr(expr)
}

type Grouping struct {
	Expression Expr
}

func (expr *Grouping) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitGroupingExpr(expr)
}

type Literal struct {
	Value interface{}
}

func (expr *Literal) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitLiteralExpr(expr)
}

type Unary struct {
	Operator *Token
	Right Expr
}

func (expr *Unary) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitUnaryExpr(expr)
}