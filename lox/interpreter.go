package lox

import (
	"fmt"
	"strings"
)

type Interpreter struct {
	env     *Environment
	globals *Environment
	locals  map[Expr]int
}

func NewInterpreter() *Interpreter {
	i := &Interpreter{
		globals: NewEnv(),
		locals:  make(map[Expr]int),
	}
	i.globals.Define("clock", &Clock{})
	// i.globals.Define("List", &LoxListClass{})
	// i.globals.Define("Map", &LoxMapClass{})
	i.env = i.globals
	return i
}

func (i *Interpreter) Interpret(statements []Stmt) {
	defer func() {
		if r := recover(); r != nil {
			//err = r.(error)
			ReportRuntimeError(r.(error))
		}
	}()

	for _, stmt := range statements {
		i.execute(stmt)
	}
}

func (i *Interpreter) VisitBinaryExpr(expr *Binary) any {
	left := i.evaluate(expr.Left)
	right := i.evaluate(expr.Right)

	switch expr.Operator.Type {
	case Minus:
		checkNumberOperands(*expr.Operator, left, right)
		return left.(float64) - right.(float64)
	case Slash:
		// TODO: Handle divide by zero
		checkNumberOperands(*expr.Operator, left, right)
		return left.(float64) / right.(float64)
	case Star:
		checkNumberOperands(*expr.Operator, left, right)
		return left.(float64) * right.(float64)
	case Plus:
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

		panic(NewRuntimeError(*expr.Operator, "Operands must be two numbers or two strings."))
	case Greater:
		checkNumberOperands(*expr.Operator, left, right)
		return left.(float64) > right.(float64)
	case GreaterEqual:
		checkNumberOperands(*expr.Operator, left, right)
		return left.(float64) >= right.(float64)
	case Less:
		checkNumberOperands(*expr.Operator, left, right)
		return left.(float64) < right.(float64)
	case LessEqual:
		checkNumberOperands(*expr.Operator, left, right)
		return left.(float64) <= right.(float64)
	case BangEqual:
		return left != right
	case EqualEqual:
		return left == right
	}

	return nil
}

func (i *Interpreter) VisitCallExpr(expr *Call) any {
	callee := i.evaluate(expr.Callee)

	args := make([]any, 0)
	for _, arg := range expr.Args {
		args = append(args, i.evaluate(arg))
	}

	if function, ok := callee.(Callable); ok {
		if len(args) != function.Arity() {
			msg := fmt.Sprintf("Expected %v arguments but got %v.", function.Arity(), len(args))
			panic(NewRuntimeError(expr.Paren, msg))
		}
		return function.Call(i, args)
	} else {
		panic(NewRuntimeError(expr.Paren, "Can only call functions and classes."))
	}
}

func (i *Interpreter) VisitGetExpr(expr *GetExpr) any {
	obj := i.evaluate(expr.Object)
	if instance, ok := obj.(Instance); ok {
		return instance.Get(expr.Name)
	}

	panic(NewRuntimeError(expr.Name, "Only instances have properties."))
}

func (i *Interpreter) VisitGroupingExpr(expr *Grouping) any {
	return i.evaluate(expr.Expression)
}

func (i *Interpreter) VisitLiteralExpr(expr *Literal) any {
	return expr.Value
}

func (i *Interpreter) VisitLogicalExpr(expr *Logical) any {
	left := i.evaluate(expr.Left)
	if expr.Operator.Type == Or {
		if isTruthy(left) {
			return left
		}
	} else {
		if !isTruthy(left) {
			return left
		}
	}

	return i.evaluate(expr.Right)
}

func (i *Interpreter) VisitSetExpr(expr *SetExpr) any {
	obj := i.evaluate(expr.Object)

	if instance, ok := obj.(*LoxInstance); ok {
		value := i.evaluate(expr.Value)
		instance.Set(expr.Name, value)
		return value
	}

	panic(NewRuntimeError(expr.Name, "Only instances have fields."))
}

func (i *Interpreter) VisitThisExpr(expr *ThisExpr) any {
	return i.lookupVariable(expr.Keyword, expr)
}

func (i *Interpreter) VisitUnaryExpr(expr *Unary) any {
	right := i.evaluate(expr.Right)

	// switch t := right.(type) {
	// case *Unary:

	// }

	switch expr.Operator.Type {
	case Minus:
		checkNumberOperand(*expr.Operator, right)
		return -right.(float64)
	case Bang:
		return !isTruthy(right)
	}

	return nil
}

func (i *Interpreter) VisitVariableExpr(expr *Variable) any {
	return i.lookupVariable(expr.Name, expr)
}

func (i *Interpreter) lookupVariable(name Token, expr Expr) any {
	if dist, ok := i.locals[expr]; ok {
		return i.env.GetAt(dist, name.Lexeme)
	} else {
		return i.globals.Get(name)
	}
}

func (i *Interpreter) VisitAssignExpr(expr *Assign) any {
	value := i.evaluate(expr.Value)

	if dist, ok := i.locals[expr]; ok {
		i.env.AssignAt(dist, expr.Name, value)
	} else {
		i.globals.Assign(expr.Name, value)
	}

	return value
}

func (i *Interpreter) evaluate(expr Expr) any {
	return expr.Accept(i)
}

func (i *Interpreter) VisitExpressionStmt(stmt *StmtExpr) any {
	return i.evaluate(stmt.Expression)
}

func (i *Interpreter) VisitFunctionStmt(stmt *FunctionStmt) any {
	function := NewFunction(stmt, i.env, false)
	i.env.Define(stmt.Name.Lexeme, function)
	return nil
}

func (i *Interpreter) VisitIfStmt(stmt *IfStmt) any {
	if isTruthy(i.evaluate(stmt.Condition)) {
		return i.execute(stmt.ThenBranch)
	} else if stmt.ElseBranch != nil {
		return i.execute(stmt.ElseBranch)
	}
	return nil
}

func (i *Interpreter) VisitPrintStmt(stmt *StmtPrint) any {
	value := i.evaluate(stmt.Expression)
	fmt.Println(stringify(value))
	return nil
}

func (i *Interpreter) VisitReturnStmt(stmt *ReturnStmt) any {
	var value any
	if stmt.Value != nil {
		value = i.evaluate(stmt.Value)
	}
	panic(NewReturnVal(value))
	//return NewReturnVal(value)
}

func (i *Interpreter) VisitVariableStmt(stmt *VariableStmt) any {
	var value any
	if stmt.Initializer != nil {
		value = i.evaluate(stmt.Initializer)
	}
	i.env.Define(stmt.Name.Lexeme, value)
	return nil
}

func (i *Interpreter) VisitWhileStmt(stmt *WhileStmt) any {
	for isTruthy(i.evaluate(stmt.Condition)) {
		i.execute(stmt.Body)
		// if res != nil {
		// 	return res
		// }
	}
	return nil
}

func (i *Interpreter) VisitBlockStmt(stmt *BlockStmt) any {
	e := NewEnv()
	e.SetEnv(i.env)
	return i.executeBlock(stmt.Statements, e)
}

func (i *Interpreter) VisitClassStmt(stmt *ClassStmt) any {
	i.env.Define(stmt.Name.Lexeme, nil)

	methods := make(map[string]*Function)
	for _, method := range stmt.Methods {
		fn := NewFunction(method, i.env, method.Name.Lexeme == "init")
		methods[method.Name.Lexeme] = fn
	}
	klass := NewLoxClass(stmt.Name.Lexeme, methods)
	i.env.Assign(stmt.Name, klass)
	return nil
}

func (i *Interpreter) execute(stmt Stmt) any {
	//	begin := time.Now().UnixNano()
	stmt.Accept(i)
	//fmt.Printf("exec res: %v\n", res)
	return nil
	// diff := time.Now().UnixNano() - begin
	// fmt.Printf("Executed stmt in %d ns", diff)
}

func (i *Interpreter) Resolve(expr Expr, depth int) {
	i.locals[expr] = depth
}

func (i *Interpreter) executeBlock(stmts []Stmt, env *Environment) any {
	prev := i.env
	defer func() {
		i.env = prev
	}()

	i.env = env
	for _, stmt := range stmts {
		i.execute(stmt)
		//fmt.Printf("res: %v", res)

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

// func isEqual(a any, b any) bool {
// 	if a == nil && b == nil {
// 		return true
// 	}

// 	if a == nil {
// 		return false
// 	}

// 	return a == b
// }

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

func checkNumberOperands(operator Token, left any, right any) {
	_, leftOk := left.(float64)
	_, rightOk := right.(float64)
	if leftOk && rightOk {
		return
	}

	panic(NewRuntimeError(operator, "Operands must be numbers."))
}

func checkNumberOperand(operator Token, operand any) {
	_, ok := operand.(float64)
	if !ok {
		panic(NewRuntimeError(operator, "Operand must be a number."))
	}
}
