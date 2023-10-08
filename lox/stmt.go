package lox

type StmtVisitor interface {
	VisitExpressionStmt(stmt *StmtExpr) any
	VisitPrintStmt(stmt *StmtPrint) any
	VisitVariableStmt(stmt *VariableStmt) any
	VisitBlockStmt(stmt *BlockStmt) any
	VisitIfStmt(stmt *IfStmt) any
	VisitWhileStmt(stmt *WhileStmt) any
	VisitFunctionStmt(stmt *FunctionStmt) any
	VisitReturnStmt(stmt *ReturnStmt) any
	VisitClassStmt(stmt *ClassStmt) any
}

type Stmt interface {
	Statement() Stmt
	Accept(v StmtVisitor) any
}

type stmt struct{}

func (s *stmt) Statement() Stmt {
	return s
}

func (s *stmt) Accept(v StmtVisitor) any {
	return nil
}

type StmtExpr struct {
	stmt
	Expression Expr
}

func (s *StmtExpr) Accept(v StmtVisitor) any {
	return v.VisitExpressionStmt(s)
}

type StmtPrint struct {
	stmt
	Expression Expr
}

func (s *StmtPrint) Accept(v StmtVisitor) any {
	return v.VisitPrintStmt(s)
}

type VariableStmt struct {
	stmt
	Name        Token
	Initializer Expr
}

func (s *VariableStmt) Accept(v StmtVisitor) any {
	return v.VisitVariableStmt(s)
}

type BlockStmt struct {
	stmt
	Statements []Stmt
}

func (s *BlockStmt) Accept(v StmtVisitor) any {
	return v.VisitBlockStmt(s)
}

type IfStmt struct {
	stmt
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (s *IfStmt) Accept(v StmtVisitor) any {
	return v.VisitIfStmt(s)
}

type WhileStmt struct {
	stmt
	Condition Expr
	Body      Stmt
}

func (s *WhileStmt) Accept(v StmtVisitor) any {
	return v.VisitWhileStmt(s)
}

type FunctionStmt struct {
	stmt
	Name   Token
	Params []Token
	Body   []Stmt
}

func (s *FunctionStmt) Accept(v StmtVisitor) any {
	return v.VisitFunctionStmt(s)
}

type ReturnStmt struct {
	stmt
	Keyword Token
	Value   Expr
}

func (s *ReturnStmt) Accept(v StmtVisitor) any {
	return v.VisitReturnStmt(s)
}

type ClassStmt struct {
	stmt
	Name       Token
	Superclass *Variable
	Methods    []*FunctionStmt
}

func (s *ClassStmt) Accept(v StmtVisitor) any {
	return v.VisitClassStmt(s)
}
