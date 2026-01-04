// Package grin provides a Single Producer Single Consumer (SPSC) lock-free ring buffer.
//
// Memory Ordering Guarantees:
// This implementation relies on Go's atomic package which provides the necessary
// memory barriers across all supported architectures (x86, ARM, RISC-V, etc.).
//
// - atomic.LoadUint64 provides acquire semantics (reads happen-before subsequent operations)
// - atomic.StoreUint64 provides release semantics (prior writes happen-before the store)
package grin

import (
	"sync/atomic"
)

type RingBuffer[T any] interface {
	Push(t T) bool
	Pop() (T, bool)
	Cap() int
	Len() int
	Available() int
}

func New[T any](size int) RingBuffer[T] {
	if size&(size-1) != 0 {
		panic("size must be power of two")
	}

	return &ringBuffer[T]{
		store: make([]T, size),
		mask:  uint64(size) - 1,
	}
}

type ringBuffer[T any] struct {
	store []T
	mask  uint64
	_     [32]byte // Do not remove

	head uint64   // Owned by the consumer, producer must use atomic operations to read
	_    [56]byte // Do not remove

	tail uint64   // Owned by the producer, consumer must use atomic operations to read
	_    [56]byte // Do not remove
}

// Push adds an item to the ring buffer.
// Returns false if the buffer is full (non-blocking).
//
// Only safe to call from a single producer goroutine.
func (b *ringBuffer[T]) Push(t T) bool {
	tail := b.tail
	head := atomic.LoadUint64(&b.head)

	// Dont overwrite existing data, reject new data until consumed
	if tail-head == uint64(len(b.store)) {
		return false
	}

	b.store[tail&b.mask] = t
	atomic.StoreUint64(&b.tail, tail+1)
	return true
}

// Pop removes and returns an item from the ring buffer.
// Returns (zero value, false) if the buffer is empty (non-blocking).
//
// Only safe to call from a single consumer goroutine.
func (b *ringBuffer[T]) Pop() (T, bool) {
	tail := atomic.LoadUint64(&b.tail)
	head := b.head

	if tail == head {
		var zero T
		return zero, false
	}

	val := b.store[head&b.mask]
	atomic.StoreUint64(&b.head, head+1)
	return val, true
}

func (b *ringBuffer[T]) Cap() int {
	return len(b.store)
}

func (b *ringBuffer[T]) Len() int {
	tail := atomic.LoadUint64(&b.tail)
	head := atomic.LoadUint64(&b.head)
	return int(tail - head)
}

func (b *ringBuffer[T]) Available() int {
	return b.Cap() - b.Len()
}
