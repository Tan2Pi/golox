package stmt

import (
	"github.com/Tan2Pi/golox/lox/expr"
	"github.com/Tan2Pi/golox/lox/tokens"
)

type Visitor interface {
	VisitExpressionStmt(stmt *Expr) any
	VisitPrintStmt(stmt *Print) any
	VisitVariableStmt(stmt *Variable) any
	VisitBlockStmt(stmt *Block) any
	VisitIfStmt(stmt *If) any
	VisitWhileStmt(stmt *While) any
	VisitFunctionStmt(stmt *Function) any
	VisitReturnStmt(stmt *Return) any
	VisitClassStmt(stmt *Class) any
}

type Stmt interface {
	Statement() Stmt
	Accept(v Visitor) any
}

type stmt struct{}

func (s *stmt) Statement() Stmt {
	return s
}

func (s *stmt) Accept(_ Visitor) any {
	return nil
}

type Expr struct {
	stmt
	Expression expr.Expr
}

func (s *Expr) Accept(v Visitor) any {
	return v.VisitExpressionStmt(s)
}

type Print struct {
	stmt
	Expression expr.Expr
}

func (s *Print) Accept(v Visitor) any {
	return v.VisitPrintStmt(s)
}

type Variable struct {
	stmt
	Name        tokens.Token
	Initializer expr.Expr
}

func (s *Variable) Accept(v Visitor) any {
	return v.VisitVariableStmt(s)
}

type Block struct {
	stmt
	Statements []Stmt
}

func (s *Block) Accept(v Visitor) any {
	return v.VisitBlockStmt(s)
}

type If struct {
	stmt
	Condition  expr.Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (s *If) Accept(v Visitor) any {
	return v.VisitIfStmt(s)
}

type While struct {
	stmt
	Condition expr.Expr
	Body      Stmt
}

func (s *While) Accept(v Visitor) any {
	return v.VisitWhileStmt(s)
}

type Function struct {
	stmt
	Name   tokens.Token
	Params []tokens.Token
	Body   []Stmt
}

func (s *Function) Accept(v Visitor) any {
	return v.VisitFunctionStmt(s)
}

type Return struct {
	stmt
	Keyword tokens.Token
	Value   expr.Expr
}

func (s *Return) Accept(v Visitor) any {
	return v.VisitReturnStmt(s)
}

type Class struct {
	stmt
	Name       tokens.Token
	Superclass *expr.Variable
	Methods    []*Function
}

func (s *Class) Accept(v Visitor) any {
	return v.VisitClassStmt(s)
}
