package resolver

import (
	"github.com/Tan2Pi/golox/lox"
	"github.com/Tan2Pi/golox/lox/expr"
	"github.com/Tan2Pi/golox/lox/interpreter"
	"github.com/Tan2Pi/golox/lox/stack"
	"github.com/Tan2Pi/golox/lox/stmt"
	"github.com/Tan2Pi/golox/lox/tokens"
)

type FunctionType int

const (
	None FunctionType = iota
	Function
	Method
	Initializer
)

type ClassType int

const (
	NoneClass ClassType = iota
	Class
	SubClass
)

type Resolver struct {
	ip        *interpreter.Interpreter
	scopes    *stack.Stack[map[string]bool]
	currFunc  FunctionType
	currClass ClassType
}

func New(i *interpreter.Interpreter) *Resolver {
	return &Resolver{
		ip:        i,
		scopes:    stack.New[map[string]bool](),
		currFunc:  None,
		currClass: NoneClass,
	}
}

func (r *Resolver) Resolve(statements []stmt.Stmt) {
	for _, s := range statements {
		r._resolveStmt(s)
	}
}

func (r *Resolver) beginScope() {
	r.scopes.Push(make(map[string]bool))
}

func (r *Resolver) endScope() {
	r.scopes.Pop()
}

func (r *Resolver) declare(name tokens.Token) {
	if r.scopes.IsEmpty() {
		return
	}

	scope := r.scopes.Peek()
	if _, exists := scope[name.Lexeme]; exists {
		lox.ErrorHandler(name, "Already a variable with this name in this scope.")
	}

	scope[name.Lexeme] = false
}

func (r *Resolver) define(name tokens.Token) {
	if r.scopes.IsEmpty() {
		return
	}

	r.scopes.Peek()[name.Lexeme] = true
}

func (r *Resolver) resolveFunction(function *stmt.Function, funcType FunctionType) {
	enclosingFn := r.currFunc
	r.currFunc = funcType
	r.beginScope()
	for _, param := range function.Params {
		r.declare(param)
		r.define(param)
	}
	r.Resolve(function.Body)
	r.endScope()
	r.currFunc = enclosingFn
}

func (r *Resolver) resolveLocal(e expr.Expr, name tokens.Token) {
	for i := r.scopes.Size() - 1; i >= 0; i-- {
		if _, ok := r.scopes.Get(i)[name.Lexeme]; ok {
			r.ip.Resolve(e, r.scopes.Size()-1-i)
			return
		}
	}
}

func (r *Resolver) _resolveStmt(stmt stmt.Stmt) {
	stmt.Accept(r)
}

func (r *Resolver) _resolveExpr(e expr.Expr) {
	e.Accept(r)
}

func (r *Resolver) VisitClassStmt(stmt *stmt.Class) any {
	enclosingClass := r.currClass
	r.currClass = Class
	r.declare(stmt.Name)
	r.define(stmt.Name)

	if stmt.Superclass != nil {
		r.currClass = SubClass
		if stmt.Superclass.Name.Lexeme == stmt.Name.Lexeme {
			lox.ErrorHandler(stmt.Superclass.Name, "A class can't inherit from itself.")
		}
		r._resolveExpr(stmt.Superclass)
		r.beginScope()
		r.scopes.Peek()["super"] = true
	}

	r.beginScope()
	r.scopes.Peek()["this"] = true

	for _, method := range stmt.Methods {
		declaration := Method
		if method.Name.Lexeme == "init" {
			declaration = Initializer
		}
		r.resolveFunction(method, declaration)
	}

	r.endScope()

	if stmt.Superclass != nil {
		r.endScope()
	}

	r.currClass = enclosingClass
	return nil
}

func (r *Resolver) VisitBlockStmt(s *stmt.Block) any {
	r.beginScope()
	r.Resolve(s.Statements)
	r.endScope()
	return nil
}

func (r *Resolver) VisitExpressionStmt(s *stmt.Expr) any {
	r._resolveExpr(s.Expression)
	return nil
}

func (r *Resolver) VisitPrintStmt(s *stmt.Print) any {
	r._resolveExpr(s.Expression)
	return nil
}

func (r *Resolver) VisitVariableStmt(s *stmt.Variable) any {
	r.declare(s.Name)
	if s.Initializer != nil {
		r._resolveExpr(s.Initializer)
	}
	r.define(s.Name)
	return nil
}

func (r *Resolver) VisitIfStmt(s *stmt.If) any {
	r._resolveExpr(s.Condition)
	r._resolveStmt(s.ThenBranch)
	if s.ElseBranch != nil {
		r._resolveStmt(s.ElseBranch)
	}
	return nil
}

func (r *Resolver) VisitWhileStmt(s *stmt.While) any {
	r._resolveExpr(s.Condition)
	r._resolveStmt(s.Body)
	return nil
}

func (r *Resolver) VisitFunctionStmt(s *stmt.Function) any {
	r.declare(s.Name)
	r.define(s.Name)

	r.resolveFunction(s, Function)
	return nil
}

func (r *Resolver) VisitReturnStmt(s *stmt.Return) any {
	if r.currFunc == None {
		lox.ErrorHandler(s.Keyword, "Can't return from top-level code.")
	}
	if s.Value != nil {
		if r.currFunc == Initializer {
			lox.ErrorHandler(s.Keyword, "Can't return a value from an initializer.")
		}
		r._resolveExpr(s.Value)
	}
	return nil
}

func (r *Resolver) VisitBinaryExpr(e *expr.Binary) any {
	r._resolveExpr(e.Left)
	r._resolveExpr(e.Right)
	return nil
}

func (r *Resolver) VisitGroupingExpr(e *expr.Grouping) any {
	r._resolveExpr(e.Expression)
	return nil
}

func (r *Resolver) VisitLiteralExpr(_ *expr.Literal) any {
	return nil
}

func (r *Resolver) VisitUnaryExpr(e *expr.Unary) any {
	r._resolveExpr(e.Right)
	return nil
}

func (r *Resolver) VisitVariableExpr(e *expr.Variable) any {
	if !r.scopes.IsEmpty() {
		if falsey, exists := r.scopes.Peek()[e.Name.Lexeme]; exists && !falsey {
			lox.ErrorHandler(e.Name, "Can't read local variable in its own initializer.")
		}
	}

	r.resolveLocal(e, e.Name)
	return nil
}

func (r *Resolver) VisitAssignExpr(e *expr.Assign) any {
	r._resolveExpr(e.Value)
	r.resolveLocal(e, e.Name)
	return nil
}

func (r *Resolver) VisitLogicalExpr(e *expr.Logical) any {
	r._resolveExpr(e.Left)
	r._resolveExpr(e.Right)
	return nil
}

func (r *Resolver) VisitSetExpr(e *expr.Set) any {
	r._resolveExpr(e.Value)
	r._resolveExpr(e.Object)
	return nil
}

func (r *Resolver) VisitThisExpr(e *expr.This) any {
	if r.currClass == NoneClass {
		lox.ErrorHandler(e.Keyword, "Can't use 'this' outside of a class.")
		return nil
	}
	r.resolveLocal(e, e.Keyword)
	return nil
}

func (r *Resolver) VisitCallExpr(e *expr.Call) any {
	r._resolveExpr(e.Callee)

	for _, arg := range e.Args {
		r._resolveExpr(arg)
	}

	return nil
}

func (r *Resolver) VisitGetExpr(e *expr.Get) any {
	r._resolveExpr(e.Object)
	return nil
}

func (r *Resolver) VisitSuperExpr(e *expr.Super) any {
	if r.currClass == NoneClass {
		lox.ErrorHandler(e.Keyword, "Can't use 'super' outside of a class.")
	} else if r.currClass != SubClass {
		lox.ErrorHandler(e.Keyword, "Can't use 'super' in a class with no superclass.")
	}
	r.resolveLocal(e, e.Keyword)
	return nil
}
