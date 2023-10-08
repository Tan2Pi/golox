package lox

type Stack[T any] struct {
	items []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		items: make([]T, 0),
	}
}

func (s *Stack[T]) Push(data T) {
	s.items = append(s.items, data)
}

func (s *Stack[T]) Pop() {
	s.items = s.items[:len(s.items)-1]
}

func (s *Stack[T]) Peek() T {
	return s.items[len(s.items)-1]
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

func (s *Stack[T]) Size() int {
	return len(s.items)
}

func (s *Stack[T]) Get(i int) T {
	return s.items[i]
}
