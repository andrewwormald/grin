package grin_test

import (
	"runtime"
	"sync"
	"testing"

	"github.com/andrewwormald/grin"
)

func TestManyToOne_PushPop(t *testing.T) {
	buf := grin.NewManyToOne[int](8)

	for i := 0; i < 5; i++ {
		if !buf.Push(i) {
			t.Fatalf("Push(%d) failed", i)
		}
	}

	for i := 0; i < 5; i++ {
		val, ok := buf.Pop()
		if !ok || val != i {
			t.Fatalf("Pop() = (%d, %v), want (%d, true)", val, ok, i)
		}
	}
}

func TestManyToOne_ConcurrentProducers(t *testing.T) {
	const producers = 4
	const perProducer = 500
	buf := grin.NewManyToOne[int](1024)

	var wg sync.WaitGroup
	wg.Add(producers)

	for p := 0; p < producers; p++ {
		prefix := p * 10000
		go func(start int) {
			defer wg.Done()
			for i := 0; i < perProducer; i++ {
				id := start + i
				for !buf.Push(id) {
					runtime.Gosched()
				}
			}
		}(prefix)
	}

	results := make(map[int]bool, producers*perProducer)
	for len(results) < producers*perProducer {
		if val, ok := buf.Pop(); ok {
			if results[val] {
				t.Fatalf("duplicate value %d detected", val)
			}
			results[val] = true
		} else {
			runtime.Gosched()
		}
	}

	wg.Wait()

	if len(results) != producers*perProducer {
		t.Fatalf("got %d results, want %d", len(results), producers*perProducer)
	}
}
