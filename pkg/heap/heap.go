package heap

import (
	"sync"

	"golang.org/x/exp/constraints"
)

// Heap represents a heap based priority queue
type Heap[T constraints.Ordered] struct {
	mu    sync.Mutex
	items []T
}

// Len returns the length of the queue.
func (h *Heap[T]) Len() int {
	return len(h.items)
}

// Empty returns true if the queue is empty.
func (h *Heap[T]) Empty() bool {
	return len(h.items) == 0
}

// Reset clears the queue
func (h *Heap[T]) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.items = []T{}
}

// Peek returns the topmost item from the queue but does not remove it from the queue.
func (h *Heap[T]) Peek() T {
	return h.items[0]
}

// Push pushes a new item to the queue.
func (h *Heap[T]) Push(item T) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.items = append(h.items, item)
	h.heapifyUp()
}

// Pop pops the topmost item off the queue
func (h *Heap[T]) Pop() T {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.swap(0, len(h.items)-1)
	var it T
	it, h.items = h.items[len(h.items)-1], h.items[:len(h.items)-1]
	h.heapifyDown()

	return it
}

func (h *Heap[T]) leftChildIdx(parent int) int {
	return (parent * 2) + 1
}

func (h *Heap[T]) rightChildIdx(parent int) int {
	return (parent * 2) + 2
}

func (h *Heap[T]) parentIdx(child int) int {
	return (child - 1) / 2
}

func (h *Heap[T]) leftChild(parent int) T {
	return h.items[h.leftChildIdx(parent)]
}

func (h *Heap[T]) rightChild(parent int) T {
	return h.items[h.rightChildIdx(parent)]
}

func (h *Heap[T]) parent(child int) T {
	return h.items[h.parentIdx(child)]
}

func (h *Heap[T]) hasParent(child int) bool {
	return child > 0 && h.parentIdx(child) >= 0
}

func (h *Heap[T]) swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
}

func (h *Heap[T]) heapifyUp() {
	idx := len(h.items) - 1

	for h.hasParent(idx) && h.parent(idx) < h.items[idx] {
		h.swap(h.parentIdx(idx), idx)
		idx = h.parentIdx(idx)
	}
}

func (h *Heap[T]) heapifyDown() {
	idx, sz := 0, len(h.items)

	for idx < sz {
		if h.leftChildIdx(idx) >= sz {
			break
		}

		var biggerIdx int
		if h.rightChildIdx(idx) < len(h.items) {
			if h.leftChild(idx) > h.rightChild(idx) {
				biggerIdx = h.leftChildIdx(idx)
			} else {
				biggerIdx = h.rightChildIdx(idx)
			}

		} else {
			biggerIdx = h.leftChildIdx(idx)
		}

		if h.items[idx] > h.items[biggerIdx] {
			break
		}

		h.swap(idx, biggerIdx)
		idx = biggerIdx
	}
}
