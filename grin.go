package grin

import (
	"sync/atomic"
)

type RingBuffer[T any] interface {
	Push(t T) bool
	Pop() (T, bool)
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

	_    [56]byte // Cache line padding for head - Do not remove
	head uint64   // Owned by the consumer, producer must use atomic operations to read

	_    [56]byte // Cache line padding for tail - Do not remove
	tail uint64   // Owned by the producer, consumer must use atomic operations to read
}

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
