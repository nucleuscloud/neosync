package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	t.Run("new queue should be empty", func(t *testing.T) {
		q := NewQueue[int]()
		assert.True(t, q.IsEmpty())
		assert.Equal(t, 0, q.Size())
	})

	t.Run("enqueue adds items to queue", func(t *testing.T) {
		q := NewQueue[string]()
		q.Enqueue("first")
		q.Enqueue("second")

		assert.Equal(t, 2, q.Size())
		assert.False(t, q.IsEmpty())
	})

	t.Run("dequeue removes and returns first item", func(t *testing.T) {
		q := NewQueue[int]()
		q.Enqueue(1)
		q.Enqueue(2)

		first, err := q.Dequeue()
		assert.NoError(t, err)
		assert.Equal(t, 1, first)
		assert.Equal(t, 1, q.Size())

		second, err := q.Dequeue()
		assert.NoError(t, err)
		assert.Equal(t, 2, second)
		assert.Equal(t, 0, q.Size())
	})

	t.Run("dequeue on empty queue returns error", func(t *testing.T) {
		q := NewQueue[int]()
		item, err := q.Dequeue()
		assert.Error(t, err)
		assert.Equal(t, ErrQueueEmpty, err)
		assert.Empty(t, item)
	})

	t.Run("peek returns first item without removing", func(t *testing.T) {
		q := NewQueue[string]()
		q.Enqueue("first")
		q.Enqueue("second")

		item, err := q.Peek()
		assert.NoError(t, err)
		assert.Equal(t, "first", item)
		assert.Equal(t, 2, q.Size())
	})

	t.Run("peek on empty queue returns error", func(t *testing.T) {
		q := NewQueue[string]()
		item, err := q.Peek()
		assert.Error(t, err)
		assert.Equal(t, ErrQueueEmpty, err)
		assert.Empty(t, item)
	})

	t.Run("handles complex struct types", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		q := NewQueue[*Person]()

		alice := &Person{Name: "Alice", Age: 30}
		bob := &Person{Name: "Bob", Age: 25}

		q.Enqueue(alice)
		q.Enqueue(bob)

		assert.Equal(t, 2, q.Size())

		first, err := q.Peek()
		assert.NoError(t, err)
		assert.Equal(t, alice, first)

		dequeued, err := q.Dequeue()
		assert.NoError(t, err)
		assert.Equal(t, alice, dequeued)
		assert.Equal(t, 1, q.Size())
	})

	t.Run("handles pointer types", func(t *testing.T) {
		type ComplexStruct struct {
			ID       int
			Data     map[string]interface{}
			Children []*ComplexStruct
		}

		q := NewQueue[*ComplexStruct]()

		parent := &ComplexStruct{
			ID: 1,
			Data: map[string]interface{}{
				"key": "value",
			},
			Children: make([]*ComplexStruct, 0),
		}

		child := &ComplexStruct{
			ID:       2,
			Data:     map[string]interface{}{},
			Children: nil,
		}

		parent.Children = append(parent.Children, child)

		q.Enqueue(parent)
		q.Enqueue(child)

		assert.Equal(t, 2, q.Size())

		firstItem, err := q.Dequeue()
		assert.NoError(t, err)
		assert.Equal(t, parent, firstItem)
		assert.Equal(t, 1, firstItem.ID)
		assert.Equal(t, "value", firstItem.Data["key"])
		assert.Equal(t, 1, len(firstItem.Children))

		secondItem, err := q.Dequeue()
		assert.NoError(t, err)
		assert.Equal(t, child, secondItem)
		assert.Equal(t, 2, secondItem.ID)
	})
}
