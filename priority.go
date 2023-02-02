package queue

import (
	"container/heap"
	"sort"
	"sync"
)

// Lesser is the interface that wraps basic Less method.
type Lesser interface {
	// Less compares the caller to the other
	Less(other any) bool
}

// Ensure Priority implements the heap.Interface.
var _ heap.Interface = (*priorityHeap[noopLesser])(nil)

// priorityHeap implements the heap.Interface, thus enabling this struct
// to be accepted as a parameter for the methods available in the heap package.
type priorityHeap[T any] struct {
	elems    []T
	lessFunc func(elem, otherElem T) bool
}

// Len is the number of elements in the collection.
func (h *priorityHeap[T]) Len() int {
	return len(h.elems)
}

// Less reports whether the element with index i
// must sort before the element with index j.
func (h *priorityHeap[T]) Less(i, j int) bool {
	return h.lessFunc(h.elems[i], h.elems[j])
}

// Swap swaps the elements with indexes i and j.
func (h *priorityHeap[T]) Swap(i, j int) {
	h.elems[i], h.elems[j] = h.elems[j], h.elems[i]
}

// Push inserts elem into the heap.
func (h *priorityHeap[T]) Push(elem any) {
	//nolint: forcetypeassert // since priorityHeap is unexported, this
	// method cannot be directly called by a library client, it is only called
	// by the heap package functions. Thus, it is safe to expect that the
	// input parameter `elem` type is always T.
	h.elems = append(h.elems, elem.(T))
}

// Pop removes and returns the highest priority element.
func (h *priorityHeap[T]) Pop() any {
	n := len(h.elems)

	elem := (h.elems)[n-1]

	h.elems = (h.elems)[0 : n-1]

	return elem
}

// fifoLessFunc will keep the elements in a fifo order.
// This is the default less function used if no less function is provided
// for the priority queue.
func fifoLessFunc[T any](T, T) bool { return false }

// Ensure Priority implements the Queue interface.
var _ Queue[noopLesser] = (*Priority[noopLesser])(nil)

// Priority is a Queue implementation.
// ! The elements must implement the Lesser interface.
//
// The ordering is given by the Lesser.Less method implementation.
// The head of the queue is always the highest priority element.
//
// ! If capacity is provided and is less than the number of elements provided,
// the elements slice is sorted and trimmed to fit the capacity.
//
// For ordered types (types that support the operators < <= >= >), the order
// can be defined by using the following operators:
// > - for ascending order
// < - for descending order.
type Priority[T any] struct {
	initialElements []T
	elements        *priorityHeap[T]

	capacity *int

	// synchronization
	lock     sync.RWMutex
	initOnce sync.Once
}

// NewPriority returns a new Priority Queue containing the given elements.
func NewPriority[T any](
	elems []T,
	lessFunc func(elem, elemAfter T) bool,
	opts ...Option,
) *Priority[T] {
	// default options
	options := options{
		capacity: nil,
	}

	for _, o := range opts {
		o.apply(&options)
	}

	if elems == nil {
		elems = []T{}
	}

	if lessFunc == nil {
		lessFunc = fifoLessFunc[T]
	}

	elementsHeap := &priorityHeap[T]{
		elems:    elems,
		lessFunc: lessFunc,
	}

	// if capacity is provided and is less than the number of elements
	// provided, the elements are sorted and trimmed to fit the capacity.
	if options.capacity != nil && *options.capacity < elementsHeap.Len() {
		sort.Slice(elementsHeap.elems, func(i, j int) bool {
			return lessFunc((elementsHeap.elems)[i], (elementsHeap.elems)[j])
		})

		elementsHeap.elems = (elementsHeap.elems)[:*options.capacity]
	}

	heap.Init(elementsHeap)

	initialElems := make([]T, elementsHeap.Len())

	copy(initialElems, elementsHeap.elems)

	pq := &Priority[T]{
		initialElements: initialElems,
		elements:        elementsHeap,
		capacity:        options.capacity,
	}

	pq.init()

	return pq
}

// ==================================Insertion=================================

// Offer inserts the element into the queue.
// If the queue is full it returns the ErrQueueIsFull error.
func (pq *Priority[T]) Offer(elem T) error {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	pq.init()

	if pq.capacity != nil && pq.elements.Len() >= *pq.capacity {
		return ErrQueueIsFull
	}

	heap.Push(pq.elements, elem)

	return nil
}

// Reset sets the queue to its initial stat, by replacing the current
// elements with the elements provided at creation.
func (pq *Priority[T]) Reset() {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	pq.init()

	if pq.elements.Len() > len(pq.initialElements) {
		pq.elements.elems = (pq.elements.elems)[:len(pq.initialElements)]
	}

	if pq.elements.Len() < len(pq.initialElements) {
		pq.elements.elems = make([]T, len(pq.initialElements))
	}

	copy(pq.elements.elems, pq.initialElements)
}

// ===================================Removal==================================

// Get removes and returns the head of the queue.
// If no element is available it returns an ErrNoElementsAvailable error.
func (pq *Priority[T]) Get() (elem T, _ error) {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	pq.init()

	if pq.elements.Len() == 0 {
		return elem, ErrNoElementsAvailable
	}

	//nolint: forcetypeassert // since the heap package does not yet support
	// generic types it has to use the `any` type. In this case, by design,
	// type of the items available in the pq.elements collection is always T.
	return heap.Pop(pq.elements).(T), nil
}

// =================================Examination================================

// Peek retrieves but does not return the head of the queue.
func (pq *Priority[T]) Peek() (elem T, _ error) {
	pq.lock.RLock()
	defer pq.lock.RUnlock()

	pq.init()

	if pq.elements.Len() == 0 {
		return elem, ErrNoElementsAvailable
	}

	return pq.elements.elems[0], nil
}

// Size returns the number of elements in the queue.
func (pq *Priority[T]) Size() int {
	pq.lock.RLock()
	defer pq.lock.RUnlock()

	pq.init()

	return pq.elements.Len()
}

func (pq *Priority[T]) init() {
	pq.initOnce.Do(func() {
		if pq.elements == nil {
			pq.elements = &priorityHeap[T]{
				lessFunc: fifoLessFunc[T],
			}
		}
	})
}
