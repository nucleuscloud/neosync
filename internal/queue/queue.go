package queue

import (
	"errors"
)

var (
	ErrQueueEmpty = errors.New("queue is empty")
)

type Queue[T any] struct {
	items []T
}

// Create a new Queue
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		items: make([]T, 0),
	}
}

// Add an item to the end of the Queue
func (q *Queue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

// Remove and return the first item in the Queue
func (q *Queue[T]) Dequeue() (T, error) {
	var zero T
	if len(q.items) == 0 {
		return zero, ErrQueueEmpty
	}

	item := q.items[0]
	q.items = q.items[1:]
	return item, nil
}

// Peek at the first item without removing it
func (q *Queue[T]) Peek() (T, error) {
	if len(q.items) == 0 {
		var zero T
		return zero, ErrQueueEmpty
	}
	return q.items[0], nil
}

// Get the current size of the Queue
func (q *Queue[T]) Size() int {
	return len(q.items)
}

// Check if the Queue is empty
func (q *Queue[T]) IsEmpty() bool {
	return len(q.items) == 0
}
