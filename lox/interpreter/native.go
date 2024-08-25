package interpreter

import (
	"fmt"

	"github.com/Tan2Pi/golox/lox"
	"github.com/Tan2Pi/golox/lox/tokens"
)

type Instance interface {
	Get(name tokens.Token) any
}

type NativeError struct {
	Msg string
}

func (e NativeError) Error() string {
	return e.Msg
}

func NativePanic(format string, args ...any) {
	panic(NativeError{
		Msg: fmt.Sprintf(format, args...),
	})
}

type NativeInstance struct {
	GetFunc  func(name tokens.Token) any
	Methods  map[string]*NativeMethod
	Stringer func() string
}

func (n *NativeInstance) Get(name tokens.Token) any {
	if method, ok := n.Methods[name.Lexeme]; ok {
		return method
	}

	panic(lox.NewRuntimeError(name, fmt.Sprintf("Undefined property '%s'.", name.Lexeme)))
}

func (n *NativeInstance) String() string {
	return n.Stringer()
}

type NativeMethod struct {
	CallFunc  func(i *Interpreter, args []any) any
	ArityFunc func() int
}

func (n *NativeMethod) Call(i *Interpreter, args []any) any {
	return n.CallFunc(i, args)
}

func (n *NativeMethod) Arity() int {
	return n.ArityFunc()
}

var _ Callable = (*LoxNativeConstructor)(nil)

type LoxNativeConstructor struct {
	Name     string
	CallFunc func(i *Interpreter, args any) any
	ArityVal int
}

func (c *LoxNativeConstructor) Call(i *Interpreter, args []any) any {
	return c.CallFunc(i, args)
}

func (c *LoxNativeConstructor) Arity() int {
	return c.ArityVal
}
