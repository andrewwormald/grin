package grin_test

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/andrewwormald/grin"
)

func TestNew(t *testing.T) {
	buf := grin.New[int](8)
	if buf == nil {
		t.Fatal("New() returned nil")
	}
}

func TestPushPop(t *testing.T) {
	buf := grin.New[int](8)

	if !buf.Push(1) {
		t.Fatal("Push(1) failed")
	}
	if !buf.Push(2) {
		t.Fatal("Push(2) failed")
	}
	if !buf.Push(3) {
		t.Fatal("Push(3) failed")
	}

	if got, ok := buf.Pop(); !ok || got != 1 {
		t.Errorf("Pop() = (%d, %v), want (1, true)", got, ok)
	}
	if got, ok := buf.Pop(); !ok || got != 2 {
		t.Errorf("Pop() = (%d, %v), want (2, true)", got, ok)
	}
	if got, ok := buf.Pop(); !ok || got != 3 {
		t.Errorf("Pop() = (%d, %v), want (3, true)", got, ok)
	}
}

func TestPopEmpty(t *testing.T) {
	buf := grin.New[int](8)

	got, ok := buf.Pop()
	if ok || got != 0 {
		t.Errorf("Pop() on empty buffer = (%d, %v), want (0, false)", got, ok)
	}
}

func TestMultiplePushPop(t *testing.T) {
	buf := grin.New[int](128)

	for i := 0; i < 100; i++ {
		if !buf.Push(i) {
			t.Fatalf("Push(%d) failed", i)
		}
	}

	for i := 0; i < 100; i++ {
		got, ok := buf.Pop()
		if !ok || got != i {
			t.Errorf("Pop() = (%d, %v), want (%d, true)", got, ok, i)
		}
	}
}

func TestStringType(t *testing.T) {
	buf := grin.New[string](8)

	if !buf.Push("hello") {
		t.Fatal("Push(\"hello\") failed")
	}
	if !buf.Push("world") {
		t.Fatal("Push(\"world\") failed")
	}

	if got, ok := buf.Pop(); !ok || got != "hello" {
		t.Errorf("Pop() = (%q, %v), want (\"hello\", true)", got, ok)
	}
	if got, ok := buf.Pop(); !ok || got != "world" {
		t.Errorf("Pop() = (%q, %v), want (\"world\", true)", got, ok)
	}
}

type testStruct struct {
	ID   int
	Name string
}

func TestStructType(t *testing.T) {
	buf := grin.New[testStruct](8)

	s1 := testStruct{ID: 1, Name: "first"}
	s2 := testStruct{ID: 2, Name: "second"}

	if !buf.Push(s1) {
		t.Fatal("Push(s1) failed")
	}
	if !buf.Push(s2) {
		t.Fatal("Push(s2) failed")
	}

	if got, ok := buf.Pop(); !ok || got != s1 {
		t.Errorf("Pop() = (%+v, %v), want (%+v, true)", got, ok, s1)
	}
	if got, ok := buf.Pop(); !ok || got != s2 {
		t.Errorf("Pop() = (%+v, %v), want (%+v, true)", got, ok, s2)
	}
}

func TestFIFOOrder(t *testing.T) {
	buf := grin.New[int](8)

	values := []int{10, 20, 30, 40, 50}
	for _, v := range values {
		if !buf.Push(v) {
			t.Fatalf("Push(%d) failed", v)
		}
	}

	for i, want := range values {
		got, ok := buf.Pop()
		if !ok || got != want {
			t.Errorf("Pop()[%d] = (%d, %v), want (%d, true)", i, got, ok, want)
		}
	}
}

func TestEmptyAfterPopping(t *testing.T) {
	buf := grin.New[string](8)

	if !buf.Push("test") {
		t.Fatal("Push(\"test\") failed")
	}
	if _, ok := buf.Pop(); !ok {
		t.Fatal("Pop() failed unexpectedly")
	}

	got, ok := buf.Pop()
	if ok || got != "" {
		t.Errorf("Pop() on empty buffer = (%q, %v), want (\"\", false)", got, ok)
	}
}

func TestBufferFull(t *testing.T) {
	buf := grin.New[int](4)

	for i := 0; i < 4; i++ {
		if !buf.Push(i) {
			t.Fatalf("Push(%d) failed, buffer should not be full", i)
		}
	}

	if buf.Push(999) {
		t.Error("Push(999) succeeded when buffer should be full")
	}

	if got, ok := buf.Pop(); !ok || got != 0 {
		t.Errorf("Pop() = (%d, %v), want (0, true)", got, ok)
	}

	if !buf.Push(999) {
		t.Error("Push(999) failed after popping one element")
	}
}

func TestRingWraparound(t *testing.T) {
	buf := grin.New[int](4)

	for round := 0; round < 3; round++ {
		for i := 0; i < 4; i++ {
			val := round*10 + i
			if !buf.Push(val) {
				t.Fatalf("Round %d: Push(%d) failed", round, val)
			}
		}

		for i := 0; i < 4; i++ {
			want := round*10 + i
			got, ok := buf.Pop()
			if !ok || got != want {
				t.Errorf("Round %d: Pop() = (%d, %v), want (%d, true)", round, got, ok, want)
			}
		}
	}
}

func TestPowerOfTwoSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("New(10) should panic for non-power-of-two size")
		}
	}()

	grin.New[int](10)
}

func TestObservabilityMethods(t *testing.T) {
	buf := grin.New[int](8)

	if buf.Cap() != 8 {
		t.Errorf("Cap() = %d, want 8", buf.Cap())
	}

	if buf.Len() != 0 {
		t.Errorf("Len() = %d, want 0", buf.Len())
	}

	if buf.Available() != 8 {
		t.Errorf("Available() = %d, want 8", buf.Available())
	}

	buf.Push(1)
	buf.Push(2)
	buf.Push(3)

	if buf.Len() != 3 {
		t.Errorf("After 3 pushes, Len() = %d, want 3", buf.Len())
	}

	if buf.Available() != 5 {
		t.Errorf("After 3 pushes, Available() = %d, want 5", buf.Available())
	}

	buf.Pop()

	if buf.Len() != 2 {
		t.Errorf("After 1 pop, Len() = %d, want 2", buf.Len())
	}

	if buf.Available() != 6 {
		t.Errorf("After 1 pop, Available() = %d, want 6", buf.Available())
	}
}

func TestConcurrentPushPop(t *testing.T) {
	buf := grin.New[int](1024)
	const numItems = 100000
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < numItems; i++ {
			for !buf.Push(i) {
				runtime.Gosched()
			}
		}
		done <- true
	}()

	go func() {
		received := make([]int, 0, numItems)
		for len(received) < numItems {
			if val, ok := buf.Pop(); ok {
				received = append(received, val)
			} else {
				runtime.Gosched()
			}
		}

		for i := 0; i < numItems; i++ {
			if received[i] != i {
				t.Errorf("FIFO violation: received[%d] = %d, want %d", i, received[i], i)
				break
			}
		}
		done <- true
	}()

	timeout := time.After(10 * time.Second)
	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("Test timed out - possible deadlock")
		}
	}
}

func TestConcurrentStress(t *testing.T) {
	buf := grin.New[uint64](256)
	const duration = 2 * time.Second
	var pushCount, popCount atomic.Uint64
	stop := make(chan bool)

	go func() {
		val := uint64(0)
		for {
			select {
			case <-stop:
				return
			default:
				if buf.Push(val) {
					val++
					pushCount.Add(1)
				} else {
					runtime.Gosched()
				}
			}
		}
	}()

	go func() {
		lastVal := uint64(0)
		for {
			select {
			case <-stop:
				return
			default:
				if val, ok := buf.Pop(); ok {
					if val != lastVal {
						t.Errorf("Order violation: got %d, expected %d", val, lastVal)
					}
					lastVal++
					popCount.Add(1)
				} else {
					runtime.Gosched()
				}
			}
		}
	}()

	time.Sleep(duration)
	close(stop)
	time.Sleep(100 * time.Millisecond)

	pushTotal := pushCount.Load()
	popTotal := popCount.Load()

	t.Logf("Stress test results: %d pushes, %d pops in %v", pushTotal, popTotal, duration)

	remaining := 0
	for {
		if _, ok := buf.Pop(); ok {
			remaining++
		} else {
			break
		}
	}

	if pushTotal != popTotal+uint64(remaining) {
		t.Errorf("Count mismatch: pushed %d, popped %d, remaining %d", pushTotal, popTotal, remaining)
	}
}

func TestConcurrentMultipleRounds(t *testing.T) {
	buf := grin.New[int](64)
	const rounds = 100
	const itemsPerRound = 50

	for round := 0; round < rounds; round++ {
		var wg sync.WaitGroup
		wg.Add(2)

		go func(r int) {
			defer wg.Done()
			for i := 0; i < itemsPerRound; i++ {
				val := r*1000 + i
				for !buf.Push(val) {
					runtime.Gosched()
				}
			}
		}(round)

		go func(r int) {
			defer wg.Done()
			for i := 0; i < itemsPerRound; i++ {
				expected := r*1000 + i
				for {
					if val, ok := buf.Pop(); ok {
						if val != expected {
							t.Errorf("Round %d: got %d, want %d", r, val, expected)
						}
						break
					}
					runtime.Gosched()
				}
			}
		}(round)

		wg.Wait()
	}
}

func TestConcurrentBackpressure(t *testing.T) {
	buf := grin.New[int](8)
	const numItems = 10000
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < numItems; i++ {
			for !buf.Push(i) {
				runtime.Gosched()
			}
		}
		done <- true
	}()

	go func() {
		for i := 0; i < numItems; i++ {
			for {
				if val, ok := buf.Pop(); ok {
					if val != i {
						t.Errorf("got %d, want %d", val, i)
					}
					break
				}
				runtime.Gosched()
			}
			if i%10 == 0 {
				time.Sleep(time.Microsecond)
			}
		}
		done <- true
	}()

	timeout := time.After(10 * time.Second)
	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("Test timed out")
		}
	}
}
