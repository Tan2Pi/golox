package lox

import "fmt"

type Environment struct {
	enclosing *Environment
	values    map[string]any
}

func NewEnv() *Environment {
	return &Environment{
		enclosing: nil,
		values:    make(map[string]any),
	}
}

func (e *Environment) SetEnv(enclosing *Environment) {
	e.enclosing = enclosing
}

func (e *Environment) Define(name string, value any) {
	e.values[name] = value
	//fmt.Printf("env: %+v", e.values)
}

func (e *Environment) Assign(name Token, value any) {
	if _, ok := e.values[name.Lexeme]; ok {
		e.values[name.Lexeme] = value
		return
	}

	if e.enclosing != nil {
		e.enclosing.Assign(name, value)
		return
	}

	panic(NewRuntimeError(name,
		fmt.Sprintf("Undefined variable '%s'.", name.Lexeme),
	))
}

func (e *Environment) Get(name Token) any {
	//fmt.Printf("env: %+v", e.values)
	if val, ok := e.values[name.Lexeme]; ok {
		return val
	}

	if e.enclosing != nil {
		return e.enclosing.Get(name)
	}

	panic(NewRuntimeError(name,
		fmt.Sprintf("Undefined variable '%s'.", name.Lexeme),
	))

}

func (e *Environment) GetAt(dist int, name string) any {
	return e.ancestor(dist).values[name]
}

func (e *Environment) AssignAt(dist int, name Token, value any) {
	e.ancestor(dist).values[name.Lexeme] = value
}

func (e *Environment) ancestor(dist int) *Environment {
	ancestor := e
	for i := 0; i < dist; i++ {
		ancestor = ancestor.enclosing
	}
	return ancestor
}
