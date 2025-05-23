package queue

// A Queue is an ordered sequence of items, the order is usually first in first out.
// New items are added to the back of the queue and
// existing items are removed from the front of the queue.
//
// This interface provides basic methods for adding and extracting elements
// from the queue.
// Items are extracted from the head of the queue and added to the tail
// of the queue.
type Queue[T comparable] interface {
	// Get retrieves and removes the head of the queue.
	Get() (T, error)

	// Offer inserts the element to the tail of the queue.
	Offer(elem T) error

	// Reset sets the queue to its initial state.
	Reset()

	// Contains returns true if the queue contains the element.
	Contains(elem T) bool

	// Peek retrieves but does not remove the head of the queue.
	Peek() (T, error)

	// Size returns the number of elements in the queue.
	Size() int

	// IsEmpty returns true if the queue is empty.
	IsEmpty() bool

	// Iterator returns a channel that will be filled with the elements
	Iterator() <-chan T

	// Clear removes all elements from the queue.
	Clear() []T
}
