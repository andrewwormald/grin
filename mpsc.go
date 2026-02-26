package grin

import (
	"runtime"
	"sync/atomic"
)

// ManyToOne is a lock-free, multi-producer single-consumer (MPSC) ring buffer.
// It mirrors the semantics of Agrona's ManyToOneConcurrentArrayQueue: many
// goroutines can Push concurrently while a single goroutine Pop's in order.
// The queue is bounded, non-blocking, and uses per-slot sequence numbers to
// allow producers to safely claim slots without locks.
type ManyToOne[T any] struct {
	slots []mpscSlot[T]
	mask  uint64
	_     [32]byte // Padding to keep slots metadata off the head cache line

	head uint64   // Owned by the consumer
	_    [56]byte // Padding to prevent false sharing with producers

	tail uint64   // Shared by producers via atomic operations
	_    [56]byte // Padding to keep tail on its own cache line
}

type mpscSlot[T any] struct {
	seq uint64
	val T
}

// NewManyToOne creates a bounded MPSC ring buffer.
// Size must be a power of 2; otherwise it panics.
func NewManyToOne[T any](size int) *ManyToOne[T] {
	if size&(size-1) != 0 {
		panic("size must be power of two")
	}

	slots := make([]mpscSlot[T], size)
	for i := range slots {
		slots[i].seq = uint64(i)
	}

	return &ManyToOne[T]{
		slots: slots,
		mask:  uint64(size - 1),
	}
}

// Push adds an item to the buffer in a wait-free manner for the consumer.
// Producers contend via CAS on tail to claim a unique sequence number. A slot
// is only published after the value is written and the slot sequence is
// advanced, preventing the consumer from observing uninitialized data.
func (b *ManyToOne[T]) Push(v T) bool {
	for {
		tail := atomic.LoadUint64(&b.tail)
		slot := &b.slots[tail&b.mask]

		seq := atomic.LoadUint64(&slot.seq)
		diff := int64(seq) - int64(tail)

		// diff == 0 -> slot free; diff < 0 -> buffer full
		if diff < 0 {
			return false
		}

		if diff == 0 && atomic.CompareAndSwapUint64(&b.tail, tail, tail+1) {
			slot.val = v
			atomic.StoreUint64(&slot.seq, tail+1)
			return true
		}

		// Another producer won the CAS; yield to reduce contention.
		runtime.Gosched()
	}
}

// Pop removes and returns the next item.
// Returns (zero, false) if the buffer is empty.
func (b *ManyToOne[T]) Pop() (T, bool) {
	// head is owned by the consumer; no atomic load required.
	head := b.head
	slot := &b.slots[head&b.mask]

	seq := atomic.LoadUint64(&slot.seq)
	if int64(seq)-int64(head+1) != 0 {
		var zero T
		return zero, false
	}

	val := slot.val
	atomic.StoreUint64(&slot.seq, head+uint64(len(b.slots)))
	b.head = head + 1
	return val, true
}

func (b *ManyToOne[T]) Cap() int {
	return len(b.slots)
}

func (b *ManyToOne[T]) Len() int {
	tail := atomic.LoadUint64(&b.tail)
	head := atomic.LoadUint64(&b.head)
	return int(tail - head)
}

func (b *ManyToOne[T]) Available() int {
	return b.Cap() - b.Len()
}
