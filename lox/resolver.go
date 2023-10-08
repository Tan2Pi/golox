package lox

type FunctionType int

const (
	NONE FunctionType = iota
	FUNCTION
	METHOD
	INITALIZER
)

type ClassType int

const (
	NONE_CLASS ClassType = iota
	CLASS
)

type Resolver struct {
	ip        *Interpreter
	scopes    *Stack[map[string]bool]
	currFunc  FunctionType
	currClass ClassType
}

func NewResolver(i *Interpreter) *Resolver {
	return &Resolver{
		ip:        i,
		scopes:    NewStack[map[string]bool](),
		currFunc:  NONE,
		currClass: NONE_CLASS,
	}
}

func (r *Resolver) Resolve(statements []Stmt) {
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

func (r *Resolver) declare(name Token) {
	if r.scopes.IsEmpty() {
		return
	}

	scope := r.scopes.Peek()
	if _, exists := scope[name.Lexeme]; exists {
		LoxErrorHandler(name, "Already a variable with this name in this scope.")
	}

	scope[name.Lexeme] = false
}

func (r *Resolver) define(name Token) {
	if r.scopes.IsEmpty() {
		return
	}

	r.scopes.Peek()[name.Lexeme] = true
}

func (r *Resolver) resolveFunction(function *FunctionStmt, funcType FunctionType) {
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

func (r *Resolver) resolveLocal(expr Expr, name Token) {
	for i := r.scopes.Size() - 1; i >= 0; i-- {
		if _, ok := r.scopes.Get(i)[name.Lexeme]; ok {
			r.ip.Resolve(expr, r.scopes.Size()-1-i)
			return
		}
	}
}

func (r *Resolver) _resolveStmt(stmt Stmt) any {
	return stmt.Accept(r)
}

func (r *Resolver) _resolveExpr(expr Expr) {
	expr.Accept(r)
}

func (r *Resolver) VisitClassStmt(stmt *ClassStmt) any {
	enclosingClass := r.currClass
	r.currClass = CLASS
	r.declare(stmt.Name)
	r.define(stmt.Name)

	r.beginScope()
	r.scopes.Peek()["this"] = true

	for _, method := range stmt.Methods {
		declaration := METHOD
		if method.Name.Lexeme == "init" {
			declaration = INITALIZER
		}
		r.resolveFunction(method, declaration)
	}

	r.endScope()
	r.currClass = enclosingClass
	return nil
}

func (r *Resolver) VisitBlockStmt(stmt *BlockStmt) any {
	r.beginScope()
	r.Resolve(stmt.Statements)
	r.endScope()
	return nil
}

func (r *Resolver) VisitExpressionStmt(stmt *StmtExpr) any {
	r._resolveExpr(stmt.Expression)
	return nil
}

func (r *Resolver) VisitPrintStmt(stmt *StmtPrint) any {
	r._resolveExpr(stmt.Expression)
	return nil
}

func (r *Resolver) VisitVariableStmt(stmt *VariableStmt) any {
	r.declare(stmt.Name)
	if stmt.Initializer != nil {
		r._resolveExpr(stmt.Initializer)
	}
	r.define(stmt.Name)
	return nil
}

func (r *Resolver) VisitIfStmt(stmt *IfStmt) any {
	r._resolveExpr(stmt.Condition)
	r._resolveStmt(stmt.ThenBranch)
	if stmt.ElseBranch != nil {
		r._resolveStmt(stmt.ElseBranch)
	}
	return nil
}

func (r *Resolver) VisitWhileStmt(stmt *WhileStmt) any {
	r._resolveExpr(stmt.Condition)
	r._resolveStmt(stmt.Body)
	return nil
}

func (r *Resolver) VisitFunctionStmt(stmt *FunctionStmt) any {
	r.declare(stmt.Name)
	r.define(stmt.Name)

	r.resolveFunction(stmt, FUNCTION)
	return nil
}

func (r *Resolver) VisitReturnStmt(stmt *ReturnStmt) any {
	if r.currFunc == NONE {
		LoxErrorHandler(stmt.Keyword, "Can't return from top-level code.")
	}
	if stmt.Value != nil {
		if r.currFunc == INITALIZER {
			LoxErrorHandler(stmt.Keyword, "Can't return a value from an initializer.")
		}
		r._resolveExpr(stmt.Value)
	}
	return nil
}

func (r *Resolver) VisitBinaryExpr(expr *Binary) any {
	r._resolveExpr(expr.Left)
	r._resolveExpr(expr.Right)
	return nil
}

func (r *Resolver) VisitGroupingExpr(expr *Grouping) any {
	r._resolveExpr(expr.Expression)
	return nil
}

func (r *Resolver) VisitLiteralExpr(expr *Literal) any {
	return nil
}

func (r *Resolver) VisitUnaryExpr(expr *Unary) any {
	r._resolveExpr(expr.Right)
	return nil
}

func (r *Resolver) VisitVariableExpr(expr *Variable) any {
	if !r.scopes.IsEmpty() {
		if falsey, exists := r.scopes.Peek()[expr.Name.Lexeme]; exists && !falsey {
			LoxErrorHandler(expr.Name, "Can't read local variable in its own initializer.")
		}
	}

	r.resolveLocal(expr, expr.Name)
	return nil
}

func (r *Resolver) VisitAssignExpr(expr *Assign) any {
	r._resolveExpr(expr.Value)
	r.resolveLocal(expr, expr.Name)
	return nil
}

func (r *Resolver) VisitLogicalExpr(expr *Logical) any {
	r._resolveExpr(expr.Left)
	r._resolveExpr(expr.Right)
	return nil
}

func (r *Resolver) VisitSetExpr(expr *SetExpr) any {
	r._resolveExpr(expr.Value)
	r._resolveExpr(expr.Object)
	return nil
}

func (r *Resolver) VisitThisExpr(expr *ThisExpr) any {
	if r.currClass == NONE_CLASS {
		LoxErrorHandler(expr.Keyword, "Can't use 'this' outside of a class.")
		return nil
	}
	r.resolveLocal(expr, expr.Keyword)
	return nil
}

func (r *Resolver) VisitCallExpr(expr *Call) any {
	r._resolveExpr(expr.Callee)

	for _, arg := range expr.Args {
		r._resolveExpr(arg)
	}

	return nil
}

func (r *Resolver) VisitGetExpr(expr *GetExpr) any {
	r._resolveExpr(expr.Object)
	return nil
}
