package grin_test

import (
	"container/ring"
	"testing"

	"github.com/andrewwormald/grin"
)

func BenchmarkCustomRingBuffer_Push(b *testing.B) {
	buf := grin.New[int](1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Push(i)
	}
}

func BenchmarkStdRing_Push(b *testing.B) {
	r := ring.New(1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Value = i
		r = r.Next()
	}
}

func BenchmarkCustomRingBuffer_PushPop(b *testing.B) {
	buf := grin.New[int](1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Push(i)
		buf.Pop()
	}
}

func BenchmarkStdRing_PushPop(b *testing.B) {
	r := ring.New(1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Value = i
		r = r.Next()
		_ = r.Value
		r = r.Prev()
	}
}

func BenchmarkCustomRingBuffer_Sequential(b *testing.B) {
	buf := grin.New[int](256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Fill half the buffer
		for j := 0; j < 128; j++ {
			buf.Push(j)
		}
		// Drain half the buffer
		for j := 0; j < 128; j++ {
			buf.Pop()
		}
	}
}

func BenchmarkStdRing_Sequential(b *testing.B) {
	r := ring.New(256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Fill half the buffer
		for j := 0; j < 128; j++ {
			r.Value = j
			r = r.Next()
		}
		// Drain half the buffer
		for j := 0; j < 128; j++ {
			_ = r.Value
			r = r.Next()
		}
	}
}

func BenchmarkCustomRingBuffer_Wraparound(b *testing.B) {
	buf := grin.New[int](64)
	// Pre-fill to force wraparound
	for i := 0; i < 32; i++ {
		buf.Push(i)
	}
	for i := 0; i < 32; i++ {
		buf.Pop()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Push(i)
		buf.Pop()
	}
}

func BenchmarkStdRing_Wraparound(b *testing.B) {
	r := ring.New(64)
	// Pre-advance to simulate wraparound
	for i := 0; i < 32; i++ {
		r.Value = i
		r = r.Next()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Value = i
		r = r.Next()
		_ = r.Value
	}
}

func BenchmarkCustomRingBuffer_FillDrain(b *testing.B) {
	buf := grin.New[int](512)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Fill completely
		for j := 0; j < 512; j++ {
			buf.Push(j)
		}
		// Drain completely
		for j := 0; j < 512; j++ {
			buf.Pop()
		}
	}
}

func BenchmarkStdRing_FillDrain(b *testing.B) {
	r := ring.New(512)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := r
		// Fill completely
		for j := 0; j < 512; j++ {
			r.Value = j
			r = r.Next()
		}
		// Drain completely
		r = start
		for j := 0; j < 512; j++ {
			_ = r.Value
			r = r.Next()
		}
	}
}

func BenchmarkCustomRingBuffer_LargeBuffer(b *testing.B) {
	buf := grin.New[int](4096)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Push(i)
		if i%2 == 0 {
			buf.Pop()
		}
	}
}

func BenchmarkStdRing_LargeBuffer(b *testing.B) {
	r := ring.New(4096)
	readPtr := r
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Value = i
		r = r.Next()
		if i%2 == 0 {
			_ = readPtr.Value
			readPtr = readPtr.Next()
		}
	}
}

func BenchmarkChannel_Push(b *testing.B) {
	ch := make(chan int, 1024)
	// Drain in background to prevent blocking
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-ch:
			case <-done:
				return
			}
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- i
	}
	b.StopTimer()
	close(done)
}

func BenchmarkChannel_PushPop(b *testing.B) {
	ch := make(chan int, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- i
		<-ch
	}
}

func BenchmarkChannel_Sequential(b *testing.B) {
	ch := make(chan int, 256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Fill half the buffer
		for j := 0; j < 128; j++ {
			ch <- j
		}
		// Drain half the buffer
		for j := 0; j < 128; j++ {
			<-ch
		}
	}
}

func BenchmarkChannel_Wraparound(b *testing.B) {
	ch := make(chan int, 64)
	// Pre-fill to force wraparound
	for i := 0; i < 32; i++ {
		ch <- i
	}
	for i := 0; i < 32; i++ {
		<-ch
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- i
		<-ch
	}
}

func BenchmarkChannel_FillDrain(b *testing.B) {
	ch := make(chan int, 512)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Fill completely
		for j := 0; j < 512; j++ {
			ch <- j
		}
		// Drain completely
		for j := 0; j < 512; j++ {
			<-ch
		}
	}
}

func BenchmarkChannel_LargeBuffer(b *testing.B) {
	ch := make(chan int, 4096)
	b.ResetTimer()
	sent := 0
	received := 0
	for i := 0; i < b.N; i++ {
		select {
		case ch <- i:
			sent++
		default:
			// Channel full, skip
		}
		if i%2 == 0 && received < sent {
			select {
			case <-ch:
				received++
			default:
			}
		}
	}
}
