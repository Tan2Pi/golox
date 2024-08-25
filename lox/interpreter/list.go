package interpreter

import (
	"fmt"
	"reflect"
)

func LoxListConstructor() any {
	return &LoxNativeConstructor{
		Name: "List",
		CallFunc: func(_ *Interpreter, _ any) any {
			return NewLoxList()
		},
		ArityVal: 0,
	}
}

type LoxList struct {
	*NativeInstance
	Values []any
}

func NewLoxList() *LoxList {
	ll := &LoxList{
		NativeInstance: &NativeInstance{},
		Values:         []any{},
	}
	ll.Methods = map[string]*NativeMethod{
		"get": {
			CallFunc: func(_ *Interpreter, args []any) any {
				if index, ok := args[0].(float64); ok {
					idx := int(index)
					if idx >= len(ll.Values) {
						NativePanic("Index out of range [%v] with length %v.",
							idx, len(ll.Values))
					}
					return ll.Values[int(index)]
				} else if !ok {
					NativePanic(
						"Cannot index into List with parameter of type '%v'.",
						reflect.TypeOf(args[0]))
				}
				return nil
			},
			ArityFunc: func() int {
				return 1
			},
		},
		"append": {
			CallFunc: func(_ *Interpreter, args []any) any {
				ll.Values = append(ll.Values, args[0])
				return nil
			},
			ArityFunc: func() int {
				return 1
			},
		},
		"size": {
			CallFunc: func(_ *Interpreter, _ []any) any {
				return len(ll.Values)
			},
			ArityFunc: func() int {
				return 0
			},
		},
	}

	ll.Stringer = func() string {
		return fmt.Sprintf("%v", ll.Values)
	}

	return ll
}
