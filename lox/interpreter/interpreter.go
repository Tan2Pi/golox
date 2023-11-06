package interpreter

import (
	"fmt"
	"golox/lox"
	"golox/lox/environment"
	"golox/lox/expr"
	"golox/lox/stmt"
	"golox/lox/tokens"
	"strings"
)

type Interpreter struct {
	env     *environment.Env
	globals *environment.Env
	locals  map[expr.Expr]int
}

func NewInterpreter() *Interpreter {
	i := &Interpreter{
		globals: environment.New(),
		locals:  make(map[expr.Expr]int),
	}
	i.globals.Define("clock", &Clock{})
	i.globals.Define("List", &LoxListClass{})
	i.globals.Define("Map", &LoxMapClass{})
	i.env = i.globals
	return i
}

func (i *Interpreter) Interpret(statements []stmt.Stmt) {
	defer func() {
		if r := recover(); r != nil {
			lox.ReportRuntimeError(r.(error))
		}
	}()

	for _, stmt := range statements {
		i.execute(stmt)
	}
}

func (i *Interpreter) VisitBinaryExpr(e *expr.Binary) any {
	left := i.evaluate(e.Left)
	right := i.evaluate(e.Right)

	switch e.Operator.Type {
	case tokens.Minus:
		checkNumberOperands(*e.Operator, left, right)
		return left.(float64) - right.(float64)
	case tokens.Slash:
		// TODO: Handle divide by zero
		checkNumberOperands(*e.Operator, left, right)
		return left.(float64) / right.(float64)
	case tokens.Star:
		checkNumberOperands(*e.Operator, left, right)
		return left.(float64) * right.(float64)
	case tokens.Plus:
		leftNum, leftOk := left.(float64)
		rightNum, rightOk := right.(float64)
		if leftOk && rightOk {
			return leftNum + rightNum
		}

		leftStr, leftOk := left.(string)
		rightStr, rightOk := right.(string)
		if leftOk && rightOk {
			return leftStr + rightStr
		}

		panic(lox.NewRuntimeError(*e.Operator, "Operands must be two numbers or two strings."))
	case tokens.Greater:
		checkNumberOperands(*e.Operator, left, right)
		return left.(float64) > right.(float64)
	case tokens.GreaterEqual:
		checkNumberOperands(*e.Operator, left, right)
		return left.(float64) >= right.(float64)
	case tokens.Less:
		checkNumberOperands(*e.Operator, left, right)
		return left.(float64) < right.(float64)
	case tokens.LessEqual:
		checkNumberOperands(*e.Operator, left, right)
		return left.(float64) <= right.(float64)
	case tokens.BangEqual:
		return left != right
	case tokens.EqualEqual:
		return left == right
	}

	return nil
}

func (i *Interpreter) VisitCallExpr(e *expr.Call) any {
	callee := i.evaluate(e.Callee)

	args := make([]any, 0)
	for _, arg := range e.Args {
		args = append(args, i.evaluate(arg))
	}

	if function, ok := callee.(Callable); ok {
		if len(args) != function.Arity() {
			msg := fmt.Sprintf("Expected %v arguments but got %v.", function.Arity(), len(args))
			panic(lox.NewRuntimeError(e.Paren, msg))
		}
		return function.Call(i, args)
	} else {
		panic(lox.NewRuntimeError(e.Paren, "Can only call functions and classes."))
	}
}

func (i *Interpreter) VisitGetExpr(e *expr.Get) any {
	obj := i.evaluate(e.Object)
	if instance, ok := obj.(Instance); ok {
		return instance.Get(e.Name)
	}

	panic(lox.NewRuntimeError(e.Name, "Only instances have properties."))
}

func (i *Interpreter) VisitGroupingExpr(e *expr.Grouping) any {
	return i.evaluate(e.Expression)
}

func (i *Interpreter) VisitLiteralExpr(e *expr.Literal) any {
	return e.Value
}

func (i *Interpreter) VisitLogicalExpr(e *expr.Logical) any {
	left := i.evaluate(e.Left)
	if e.Operator.Type == tokens.Or {
		if isTruthy(left) {
			return left
		}
	} else {
		if !isTruthy(left) {
			return left
		}
	}

	return i.evaluate(e.Right)
}

func (i *Interpreter) VisitSetExpr(e *expr.Set) any {
	obj := i.evaluate(e.Object)

	if instance, ok := obj.(*LoxInstance); ok {
		value := i.evaluate(e.Value)
		instance.Set(e.Name, value)
		return value
	}

	panic(lox.NewRuntimeError(e.Name, "Only instances have fields."))
}

func (i *Interpreter) VisitSuperExpr(e *expr.Super) any {
	distance := i.locals[e]
	superclass, ok := i.env.GetAt(distance, "super").(*Class)
	if !ok {
		panic(lox.NewRuntimeError(e.Keyword, "Undefined variable 'super'."))
	}
	obj, ok := i.env.GetAt(distance-1, "this").(*LoxInstance)
	if !ok {
		panic(lox.NewRuntimeError(e.Keyword, "Undefined variable 'this'."))
	}
	method := superclass.FindMethod(e.Method.Lexeme)
	if method == nil {
		panic(lox.NewRuntimeError(e.Method, fmt.Sprintf("Undefined property '%s'.", e.Method.Lexeme)))
	}
	return method.Bind(obj)
}

func (i *Interpreter) VisitThisExpr(e *expr.This) any {
	return i.lookupVariable(e.Keyword, e)
}

func (i *Interpreter) VisitUnaryExpr(e *expr.Unary) any {
	right := i.evaluate(e.Right)

	switch e.Operator.Type {
	case tokens.Minus:
		checkNumberOperand(*e.Operator, right)
		return -right.(float64)
	case tokens.Bang:
		return !isTruthy(right)
	}

	return nil
}

func (i *Interpreter) VisitVariableExpr(e *expr.Variable) any {
	return i.lookupVariable(e.Name, e)
}

func (i *Interpreter) lookupVariable(name tokens.Token, e expr.Expr) any {
	if dist, ok := i.locals[e]; ok {
		return i.env.GetAt(dist, name.Lexeme)
	} else {
		return i.globals.Get(name)
	}
}

func (i *Interpreter) VisitAssignExpr(e *expr.Assign) any {
	value := i.evaluate(e.Value)

	if dist, ok := i.locals[e]; ok {
		i.env.AssignAt(dist, e.Name, value)
	} else {
		i.globals.Assign(e.Name, value)
	}

	return value
}

func (i *Interpreter) evaluate(e expr.Expr) any {
	return e.Accept(i)
}

func (i *Interpreter) VisitExpressionStmt(s *stmt.Expr) any {
	return i.evaluate(s.Expression)
}

func (i *Interpreter) VisitFunctionStmt(s *stmt.Function) any {
	function := NewFunction(s, i.env, false)
	i.env.Define(s.Name.Lexeme, function)
	return nil
}

func (i *Interpreter) VisitIfStmt(s *stmt.If) any {
	if isTruthy(i.evaluate(s.Condition)) {
		return i.execute(s.ThenBranch)
	} else if s.ElseBranch != nil {
		return i.execute(s.ElseBranch)
	}
	return nil
}

func (i *Interpreter) VisitPrintStmt(s *stmt.Print) any {
	value := i.evaluate(s.Expression)
	fmt.Println(stringify(value))
	return nil
}

func (i *Interpreter) VisitReturnStmt(stmt *stmt.Return) any {
	var value any
	if stmt.Value != nil {
		value = i.evaluate(stmt.Value)
	}
	panic(lox.NewReturnVal(value))
	//return NewReturnVal(value)
}

func (i *Interpreter) VisitVariableStmt(s *stmt.Variable) any {
	var value any
	if s.Initializer != nil {
		value = i.evaluate(s.Initializer)
	}
	i.env.Define(s.Name.Lexeme, value)
	return nil
}

func (i *Interpreter) VisitWhileStmt(s *stmt.While) any {
	for isTruthy(i.evaluate(s.Condition)) {
		i.execute(s.Body)
		// if res != nil {
		// 	return res
		// }
	}
	return nil
}

func (i *Interpreter) VisitBlockStmt(s *stmt.Block) any {
	e := environment.New()
	e.SetEnv(i.env)
	return i.executeBlock(s.Statements, e)
}

func (i *Interpreter) VisitClassStmt(s *stmt.Class) any {
	var superclass *Class
	if s.Superclass != nil {
		var ok bool
		superclass, ok = i.evaluate(s.Superclass).(*Class)
		if !ok {
			panic(lox.NewRuntimeError(s.Superclass.Name, "Superclass must be a class."))
		}
	}
	i.env.Define(s.Name.Lexeme, nil)

	if superclass != nil {
		var e *environment.Env
		e, i.env = i.env, environment.New()
		i.env.SetEnv(e)
		i.env.Define("super", superclass)
	}

	methods := make(map[string]*Function)
	for _, method := range s.Methods {
		fn := NewFunction(method, i.env, method.Name.Lexeme == "init")
		methods[method.Name.Lexeme] = fn
	}
	klass := NewClass(s.Name.Lexeme, superclass, methods)
	if superclass != nil {
		i.env = i.env.Enclosing()
	}
	i.env.Assign(s.Name, klass)
	return nil
}

func (i *Interpreter) execute(stmt stmt.Stmt) any {
	stmt.Accept(i)
	return nil
}

func (i *Interpreter) Resolve(e expr.Expr, depth int) {
	i.locals[e] = depth
}

func (i *Interpreter) executeBlock(stmts []stmt.Stmt, env *environment.Env) any {
	prev := i.env
	defer func() {
		i.env = prev
	}()

	i.env = env
	for _, stmt := range stmts {
		i.execute(stmt)

	}
	return nil
}

// helper methods

func isTruthy(obj any) bool {
	if obj == nil {
		return false
	}

	if b, ok := obj.(bool); ok {
		return b
	}

	return true
}

func stringify(object any) string {
	if object == nil {
		return "nil"
	}

	if num, ok := object.(float64); ok {
		text := fmt.Sprintf("%.3f", num)
		return strings.TrimSuffix(text, ".000")
	}

	return fmt.Sprintf("%v", object)
}

func checkNumberOperands(operator tokens.Token, left any, right any) {
	_, leftOk := left.(float64)
	_, rightOk := right.(float64)
	if leftOk && rightOk {
		return
	}

	panic(lox.NewRuntimeError(operator, "Operands must be numbers."))
}

func checkNumberOperand(operator tokens.Token, operand any) {
	_, ok := operand.(float64)
	if !ok {
		panic(lox.NewRuntimeError(operator, "Operand must be a number."))
	}
}
