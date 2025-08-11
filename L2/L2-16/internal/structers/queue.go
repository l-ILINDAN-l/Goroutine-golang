package structers

import (
	"container/list"
)

type Queue[T any] struct {
	elem *list.List
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		list.New(),
	}
}

func (q *Queue[T]) Enqueue(value T) {
	q.elem.PushBack(value)
}

func (q *Queue[T]) Dequeue() (T, bool) {
	if q.IsEmpty() {
		var zero T
		return zero, false
	}

	element := q.elem.Front()
	q.elem.Remove(element)

	return element.Value.(T), true
}

// IsEmpty проверяет, пуста ли очередь.
func (q *Queue[T]) IsEmpty() bool {
	return q.elem.Len() == 0
}

// Len возвращает количество элементов в очереди.
func (q *Queue[T]) Len() int {
	return q.elem.Len()
}
