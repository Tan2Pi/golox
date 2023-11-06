package interpreter

import (
	"fmt"
	"golox/lox"
	"golox/lox/tokens"
	"time"
)

type Instance interface {
	Get(name tokens.Token) any
}

type Clock struct{}

func (c *Clock) Arity() int {
	return 0
}

func (c *Clock) Call(i *Interpreter, args []any) any {
	return float64(time.Now().Unix())
}

func (c *Clock) String() string {
	return "<native fn>"
}

type LoxListClass struct {
	Name string
}

type LoxListInstance struct {
	Values []any
}

func (l *LoxListClass) Call(i *Interpreter, args []any) any {
	return &LoxListInstance{
		Values: []any{},
	}
}

func (l *LoxListClass) Arity() int {
	return 0
}

func (l *LoxListClass) String() string {
	return "<native class>"
}

func (l *LoxListInstance) Call(i *Interpreter, args []any) any {
	//fmt.Printf("args: %v", args)
	return nil
}

func (l *LoxListInstance) Get(name tokens.Token) any {

	if name.Lexeme == "append" {
		appender := new(LoxListAppend)
		appender.LoxListInstance = l
		return appender
	} else if name.Lexeme == "getAt" {
		getter := new(LoxListGetAt)
		getter.LoxListInstance = l
		return getter
	}

	panic(lox.NewRuntimeError(name, fmt.Sprintf("Undefined property '%s'.", name.Lexeme)))
}

func (l *LoxListInstance) Arity() int {
	return 0
}

func (l *LoxListInstance) append(item any) {
	l.Values = append(l.Values, item)
}

func (l *LoxListInstance) getAt(index int) any {
	return l.Values[index]
}

func (l *LoxListInstance) String() string {
	return fmt.Sprintf("%v", l.Values)
}

type LoxListGetAt struct {
	*LoxListInstance
}

func (l *LoxListGetAt) Call(i *Interpreter, args []any) any {
	if index, ok := args[0].(float64); ok {
		return l.getAt(int(index))
	}
	return nil
}

func (l *LoxListGetAt) Arity() int {
	return 1
}

type LoxListAppend struct {
	*LoxListInstance
}

func (l *LoxListAppend) Call(i *Interpreter, args []any) any {
	l.append(args[0])
	return nil
}

func (l *LoxListAppend) Arity() int {
	return 1
}

type LoxMapClass struct {
	Name string
}

func (l *LoxMapClass) Call(i *Interpreter, args []any) any {
	return &LoxMapInstance{
		Values: make(map[any]any),
	}
}

func (l *LoxMapClass) Arity() int {
	return 0
}

type LoxMapInstance struct {
	Values map[any]any
}

func (l *LoxMapInstance) Call(i *Interpreter, args []any) any {
	return nil
}

func (l *LoxMapInstance) Get(name tokens.Token) any {
	if name.Lexeme == "put" {
		putter := new(LoxMapPut)
		putter.LoxMapInstance = l
		return putter
	} else if name.Lexeme == "get" {
		getter := new(LoxMapGet)
		getter.LoxMapInstance = l
		return getter
	}

	panic(lox.NewRuntimeError(name, fmt.Sprintf("Undefined property '%s'.", name.Lexeme)))
}

func (l *LoxMapInstance) Arity() int {
	return 0
}

func (l *LoxMapInstance) put(key, value any) {
	l.Values[key] = value
}

func (l *LoxMapInstance) get(key any) any {
	return l.Values[key]
}

func (l *LoxMapInstance) String() string {
	return fmt.Sprintf("%v", l.Values)
}

type LoxMapPut struct {
	*LoxMapInstance
}

func (l *LoxMapPut) Call(i *Interpreter, args []any) any {
	l.put(args[0], args[1])
	return nil
}

func (l *LoxMapPut) Arity() int {
	return 2
}

type LoxMapGet struct {
	*LoxMapInstance
}

func (l *LoxMapGet) Call(i *Interpreter, args []any) any {
	return l.get(args[0])
}

func (l *LoxMapGet) Arity() int {
	return 1
}
