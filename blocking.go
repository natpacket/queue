package queue

import (
	"sync"
)

// Blocking provides a read-only queue for a list of T.
//
// It supports operations for retrieving and adding elements to a FIFO queue.
// If there are no elements available the retrieve operations wait until
// elements are added to the queue.
type Blocking[T any] struct {
	// elements queue
	elements      []T
	elementsIndex int

	lock sync.Mutex
	cond *sync.Cond
}

// NewBlocking returns a new Blocking Queue containing the given elements..
func NewBlocking[T any](elements []T) *Blocking[T] {
	b := &Blocking[T]{
		elements:      elements,
		elementsIndex: 0,
		lock:          sync.Mutex{},
	}

	b.cond = sync.NewCond(&b.lock)

	return b
}

// Take removes and returns the head of the elements queue.
// If no element is available it waits until the queue
//
// It does not actually remove elements from the elements slice, but
// it's incrementing the underlying index.
func (q *Blocking[T]) Take() (v T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	idx := q.getNextIndexOrWait()

	elem := q.elements[idx]

	q.elementsIndex++

	return elem
}

func (q *Blocking[T]) getNextIndexOrWait() int {
	if q.elementsIndex < len(q.elements) {
		return q.elementsIndex
	}

	q.cond.Wait()

	return q.getNextIndexOrWait()
}

// Push inserts the element into the queue,
// while also increasing the queue size.
func (q *Blocking[T]) Push(elem T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.elements = append(q.elements, elem)

	q.cond.Signal()
}

// Peek retrieves but does not return the head of the queue.
func (q *Blocking[T]) Peek() T {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.elementsIndex == len(q.elements) {
		q.cond.Wait()
	}

	elem := q.elements[q.elementsIndex]

	return elem
}

// Reset sets the queue elements index to 0. The queue will be in its initial
// state.
func (q *Blocking[T]) Reset() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.elementsIndex = 0

	q.cond.Broadcast()
}
