package grin_test

import (
	"testing"

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

	// Fill the buffer completely
	for i := 0; i < 4; i++ {
		if !buf.Push(i) {
			t.Fatalf("Push(%d) failed, buffer should not be full", i)
		}
	}

	// Try to push one more, should fail
	if buf.Push(999) {
		t.Error("Push(999) succeeded when buffer should be full")
	}

	// Pop one element
	if got, ok := buf.Pop(); !ok || got != 0 {
		t.Errorf("Pop() = (%d, %v), want (0, true)", got, ok)
	}

	// Now we should be able to push again
	if !buf.Push(999) {
		t.Error("Push(999) failed after popping one element")
	}
}

func TestRingWraparound(t *testing.T) {
	buf := grin.New[int](4)

	// Fill and empty the buffer multiple times to test wraparound
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

	grin.New[int](10) // Should panic
}
