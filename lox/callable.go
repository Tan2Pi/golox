package lox

import "fmt"

type Callable interface {
	Call(i *Interpreter, args []any) any
	Arity() int
}

type Function struct {
	Declaration *FunctionStmt
	Closure     *Environment
	isInit      bool
}

func NewFunction(declaration *FunctionStmt, closure *Environment, isInit bool) *Function {
	return &Function{
		Declaration: declaration,
		Closure:     closure,
		isInit:      isInit,
	}
}

func (f *Function) Bind(instance *LoxInstance) *Function {
	env := NewEnv()
	env.SetEnv(f.Closure)
	env.Define("this", instance)
	return NewFunction(f.Declaration, env, f.isInit)
}

func (f *Function) Call(i *Interpreter, args []any) (returnable any) {
	env := NewEnv()
	env.SetEnv(f.Closure)

	for i := range f.Declaration.Params {
		env.Define(f.Declaration.Params[i].Lexeme, args[i])
	}

	defer func() {
		if r := recover(); r != nil {
			switch ret := r.(type) {
			case ReturnValue:
				if f.isInit {
					returnable = f.Closure.GetAt(0, "this")
					return
				}
				returnable = ret.Value
				return
			default:
				panic(r)
			}
		}
	}()

	ret := i.executeBlock(f.Declaration.Body, env)
	if f.isInit {
		returnable = f.Closure.GetAt(0, "this")
		return
	}
	if ret != nil {
		//fmt.Printf("got ret: %v", ret)
		if ret, ok := ret.(ReturnValue); ok {
			returnable = ret.Value
			return
		}
	}
	return
}

func (f *Function) Arity() int {
	return len(f.Declaration.Params)
}

func (f *Function) String() string {
	return fmt.Sprintf("<fn %s>", f.Declaration.Name.Lexeme)
}

type LoxClass struct {
	Name    string
	methods map[string]*Function
}

func NewLoxClass(name string, methods map[string]*Function) *LoxClass {
	return &LoxClass{
		Name:    name,
		methods: methods,
	}
}

func (c *LoxClass) Call(i *Interpreter, args []any) any {
	instance := NewLoxInstance(c)
	init := c.FindMethod("init")
	if init != nil {
		init.Bind(instance).Call(i, args)
	}
	return instance
}

func (c *LoxClass) Arity() int {
	init := c.FindMethod("init")
	if init == nil {
		return 0
	}
	return init.Arity()
}

func (c *LoxClass) FindMethod(name string) *Function {
	if method, ok := c.methods[name]; ok {
		return method
	}
	return nil
}

func (c *LoxClass) String() string {
	return c.Name
}

type LoxInstance struct {
	Class  *LoxClass
	fields map[string]any
}

func NewLoxInstance(class *LoxClass) *LoxInstance {
	return &LoxInstance{
		Class:  class,
		fields: make(map[string]any),
	}
}

func (i *LoxInstance) Get(name Token) any {
	//fmt.Printf("get: token: %+v, instance: %+v\n", name, i.fields)
	if val, ok := i.fields[name.Lexeme]; ok {
		return val
	}

	if method := i.Class.FindMethod(name.Lexeme); method != nil {
		return method.Bind(i)
	}

	panic(NewRuntimeError(name, fmt.Sprintf("Undefined property '%s'.", name.Lexeme)))
}

func (i *LoxInstance) Set(name Token, value any) {
	i.fields[name.Lexeme] = value
	//fmt.Printf("set: token: %+v, instance: %+v\n", name, i.fields)
}

func (i *LoxInstance) String() string {
	return fmt.Sprintf("%s instance", i.Class.Name)
}
