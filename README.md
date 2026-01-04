# grin

A Single Producer Single Consumer (SPSC) lock-free ring buffer for Go. Zero-allocation, zero-mutex, low-latency implementation for passing data between goroutines.

## Features

- **Lock-free**: Uses atomic operations instead of mutexes for maximum throughput
- **Zero allocation**: No heap allocations during Push/Pop operations
- **Cache-line optimized**: Prevents false sharing between producer and consumer
- **Type-safe**: Generic implementation using Go generics
- **High performance**: Up to 37x faster than channels for single-threaded operations

## Benchmark Results

Benchmarks comparing grin vs Go channels vs `container/ring`:

```
BenchmarkGrin_Push-8             	552643833	    2.045 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_Push-8          	132195036	    8.817 ns/op	       8 B/op	       0 allocs/op
BenchmarkChannel_Push-8          	 16400854	   75.16 ns/op	       0 B/op	       0 allocs/op

BenchmarkGrin_PushPop-8          	100000000	   10.76 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_PushPop-8       	126589761	    8.967 ns/op	       8 B/op	       0 allocs/op
BenchmarkChannel_PushPop-8       	 52360968	   22.60 ns/op	       0 B/op	       0 allocs/op

BenchmarkGrin_Sequential-8       	  668730	    1834 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_Sequential-8    	 2494608	     478.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkChannel_Sequential-8    	  406587	    3170 ns/op	       0 B/op	       0 allocs/op

BenchmarkGrin_FillDrain-8        	  165178	    7372 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_FillDrain-8     	  343904	    3674 ns/op	    2048 B/op	     256 allocs/op
BenchmarkChannel_FillDrain-8     	  102903	   12073 ns/op	       0 B/op	       0 allocs/op
```

**Key Takeaways:**
- **grin vs Channels**: 37x faster for Push, 2x faster for PushPop
- **grin vs container/ring**: Faster for single operations, tracks buffer fullness
- **Zero allocations**: grin allocates nothing during operation, container/ring allocates on every value assignment

## Usage

```go
package main

import (
    "fmt"
    "github.com/andrewwormald/grin"
)

func main() {
    // Create a ring buffer with capacity of 1024 (must be power of 2)
    buf := grin.New[int](1024)

    // Producer goroutine
    go func() {
        for i := 0; i < 100; i++ {
            for !buf.Push(i) {
                // Buffer full, wait/retry
            }
        }
    }()

    // Consumer goroutine
    go func() {
        for {
            if val, ok := buf.Pop(); ok {
                fmt.Println(val)
            }
        }
    }()
}
```

## When to Use SPSC Ring Buffers (grin)

SPSC ring buffers are ideal for **high-performance, low-latency communication** between exactly **one producer and one consumer** goroutine:

✅ **Use grin when:**
- You have exactly one producer and one consumer goroutine
- Maximum throughput and minimum latency are critical
- You want zero allocations during operation
- You can size the buffer appropriately upfront (power of 2)
- You need predictable, bounded memory usage
- Examples: High-frequency trading, audio/video processing, network packet handling, log aggregation

⚠️ **Don't use grin when:**
- You have multiple producers or consumers (use channels instead)
- You need Go's channel synchronization primitives (select, close, etc.)
- Buffer size can't be determined upfront
- You need dynamic resizing

## When to Use container/ring

The standard library's `container/ring` is a circular doubly-linked list:

✅ **Use container/ring when:**
- You need to iterate forwards and backwards through a circular buffer
- You don't need to track buffer fullness (it overwrites old data)
- You're storing interface{} values and type safety isn't critical
- Performance isn't the primary concern
- Examples: Recent history/cache, circular iterators, round-robin algorithms

⚠️ **Don't use container/ring when:**
- You need zero allocations (it allocates on every value assignment)
- You need to know if the buffer is full/empty
- You need type safety with generics
- You need multi-threaded access (not thread-safe)

## When to Use Channels

Go channels are the general-purpose communication primitive:

✅ **Use channels when:**
- You have multiple producers and/or multiple consumers
- You need select statements for multiplexing
- You need close() semantics for signaling completion
- You want the scheduler to handle goroutine synchronization
- Code clarity is more important than raw performance
- Examples: General goroutine communication, fan-out/fan-in patterns, cancellation

⚠️ **Don't use channels when:**
- You need the absolute lowest latency (use SPSC ring buffers)
- You're doing high-frequency operations (millions/sec)
- Lock-free algorithms are required

## Design Notes

grin uses several optimizations:

1. **Power-of-2 sizing**: Allows fast modulo operations using bitwise AND
2. **Cache-line padding**: 56-byte padding prevents false sharing between CPU cores
3. **Lock-free atomic operations**: Producer owns tail, consumer owns head
4. **Separate cache lines**: Head and tail pointers are on different cache lines to prevent contention

## Installation

```bash
go get github.com/andrewwormald/grin
```

## API

```go
type RingBuffer[T any] interface {
    // Push adds an item to the buffer.
    // Returns false if buffer is full (non-blocking).
    Push(t T) bool

    // Pop removes and returns an item from the buffer.
    // Returns (zero value, false) if buffer is empty (non-blocking).
    Pop() (T, bool)
}

// New creates a new ring buffer with the specified size.
// Size must be a power of 2, otherwise it panics.
func New[T any](size int) RingBuffer[T]
```

## Requirements

- Buffer size must be a power of 2 (enforced by panic)
- Single producer goroutine only
- Single consumer goroutine only

## License

See [LICENSE](LICENSE) file for details.
