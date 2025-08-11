package structers

type Set[T comparable] struct {
	elem map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{make(map[T]struct{})}
}

func (s Set[T]) Add(values ...T) {
	for _, v := range values {
		s.elem[v] = struct{}{}
	}
}

func (s Set[T]) Remove(values ...T) {
	for _, v := range values {
		delete(s.elem, v)
	}
}

func (s Set[T]) Len() int {
	return len(s.elem)
}

func (s Set[T]) GetAll() []T {
	elements := make([]T, 0, len(s.elem))
	for k := range s.elem {
		elements = append(elements, k)
	}
	return elements
}

func (s Set[T]) Contains(value T) bool {
	if _, ok := s.elem[value]; ok {
		return true
	}
	return false
}
