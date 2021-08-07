package lox

import (
	"fmt"
	//"strconv"
	"strings"
)

type AstPrinter struct {}

func (p *AstPrinter) Print(expr Expr) string {
	return expr.Accept(p).(string)
}

func (p *AstPrinter) VisitBinaryExpr(expr *Binary) interface{} {
	return p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (p *AstPrinter) VisitGroupingExpr(expr *Grouping) interface{} {
	return p.parenthesize("group", expr.Expression)
}

func (p *AstPrinter) VisitLiteralExpr(expr *Literal) interface{} {
	if expr.Value == nil {
		return "nil"
	}
	switch v := expr.Value.(type) {
	case string:
		return string(v)
	case int, float64:
		return fmt.Sprintf("%v", v)
	default:
		panic("unknown literal")	
	}
}

func (p *AstPrinter) VisitUnaryExpr(expr *Unary) interface{} {
	return p.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (p *AstPrinter) parenthesize(name string, exprs ...Expr) string {
	var builder strings.Builder
	builder.WriteString("(")
	builder.WriteString(name)
	for _, expr := range exprs {
		builder.WriteString(" ")
		builder.WriteString(expr.Accept(p).(string))
	}
	builder.WriteString(")")
	
	return builder.String()
}