package interpreter

import (
	"fmt"
)

func LoxMapConstructor() any {
	return &LoxNativeConstructor{
		Name: "Map",
		CallFunc: func(_ *Interpreter, _ any) any {
			return NewLoxMap()
		},
		ArityVal: 0,
	}
}

type LoxMap struct {
	*NativeInstance
	data map[any]any
}

func NewLoxMap() *LoxMap {
	lm := &LoxMap{
		NativeInstance: &NativeInstance{},
		data:           make(map[any]any),
	}

	lm.Methods = map[string]*NativeMethod{
		"put": {
			CallFunc: func(_ *Interpreter, args []any) any {
				lm.data[args[0]] = args[1]
				return nil
			},
			ArityFunc: func() int {
				return 2 //nolint:mnd // fairly self-explanatory
			},
		},
		"get": {
			CallFunc: func(_ *Interpreter, args []any) any {
				if val, ok := lm.data[args[0]]; ok {
					return val
				}

				return nil
			},
			ArityFunc: func() int {
				return 1
			},
		},
		"size": {
			CallFunc: func(_ *Interpreter, _ []any) any {
				return len(lm.data)
			},
			ArityFunc: func() int {
				return 0
			},
		},
		"contains": {
			CallFunc: func(_ *Interpreter, args []any) any {
				_, ok := lm.data[args[0]]
				return ok
			},
			ArityFunc: func() int {
				return 1
			},
		},
		"remove": {
			CallFunc: func(_ *Interpreter, args []any) any {
				delete(lm.data, args[0])
				return nil
			},
			ArityFunc: func() int {
				return 1
			},
		},
	}

	lm.Stringer = func() string {
		return fmt.Sprintf("%v", lm.data)
	}

	return lm
}
