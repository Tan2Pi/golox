package environment

import (
	"fmt"

	"github.com/Tan2Pi/golox/lox"
	"github.com/Tan2Pi/golox/lox/tokens"
)

type Env struct {
	enclosing *Env
	values    map[string]any
}

func New() *Env {
	return &Env{
		enclosing: nil,
		values:    make(map[string]any),
	}
}

func (e *Env) SetEnv(enclosing *Env) {
	e.enclosing = enclosing
}

func (e *Env) Enclosing() *Env {
	return e.enclosing
}

func (e *Env) Define(name string, value any) {
	e.values[name] = value
}

func (e *Env) Assign(name tokens.Token, value any) {
	if _, ok := e.values[name.Lexeme]; ok {
		e.values[name.Lexeme] = value
		return
	}

	if e.enclosing != nil {
		e.enclosing.Assign(name, value)
		return
	}

	panic(lox.NewRuntimeError(name,
		fmt.Sprintf("Undefined variable '%s'.", name.Lexeme),
	))
}

func (e *Env) Get(name tokens.Token) any {
	if val, ok := e.values[name.Lexeme]; ok {
		return val
	}

	if e.enclosing != nil {
		return e.enclosing.Get(name)
	}

	panic(lox.NewRuntimeError(name,
		fmt.Sprintf("Undefined variable '%s'.", name.Lexeme),
	))
}

func (e *Env) GetAt(dist int, name string) any {
	return e.ancestor(dist).values[name]
}

func (e *Env) AssignAt(dist int, name tokens.Token, value any) {
	e.ancestor(dist).values[name.Lexeme] = value
}

func (e *Env) ancestor(dist int) *Env {
	ancestor := e
	for range dist {
		ancestor = ancestor.enclosing
	}
	return ancestor
}
