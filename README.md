# grin

A Single Producer Single Consumer (SPSC) lock-free ring buffer for Go. Zero-allocation, zero-mutex, low-latency implementation for passing data between goroutines.

## Features

- **Lock-free**: Uses atomic operations instead of mutexes for maximum throughput
- **Zero allocation**: No heap allocations during Push/Pop operations
- **Cache-line optimized**: Prevents false sharing between producer and consumer
- **Type-safe**: Generic implementation using Go generics
- **High performance**: Up to 6x faster than channels for single-producer/single-consumer operations

## Benchmark Results

Benchmarks comparing grin vs Go channels vs `container/ring` (Apple M1 Pro, Go 1.23):

```
BenchmarkGrin_Push-8             	97138131	   11.96 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_Push-8          	137294083	    8.800 ns/op	       8 B/op	       0 allocs/op
BenchmarkChannel_Push-8          	16363477	   71.60 ns/op	       0 B/op	       0 allocs/op

BenchmarkGrin_PushPop-8          	100000000	   10.58 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_PushPop-8       	132342357	    9.282 ns/op	       8 B/op	       0 allocs/op
BenchmarkChannel_PushPop-8       	52933585	   22.76 ns/op	       0 B/op	       0 allocs/op

BenchmarkGrin_Sequential-8       	  659934	    1820 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_Sequential-8    	 2572219	     465.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkChannel_Sequential-8    	  407391	    2957 ns/op	       0 B/op	       0 allocs/op

BenchmarkGrin_FillDrain-8        	  164268	    7300 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_FillDrain-8     	  345164	    3455 ns/op	    2048 B/op	     256 allocs/op
BenchmarkChannel_FillDrain-8     	  101649	   11808 ns/op	       0 B/op	       0 allocs/op
```

**Key Takeaways:**
- **grin vs Channels**: 6x faster for Push, 2x faster for PushPop, 1.6x faster for FillDrain
- **grin vs container/ring**: Slower for sequential bulk operations (4x), but grin is concurrent-safe for SPSC and tracks buffer fullness. Different use cases—container/ring has no atomics overhead but isn't thread-safe.
- **Zero allocations**: grin allocates nothing during operation, container/ring allocates on every value assignment

## Usage

```go
package main

import (
    "fmt"
    "runtime"
    "time"

    "github.com/andrewwormald/grin"
)

func main() {
    // Create a ring buffer with capacity of 1024 (must be power of 2)
    buf := grin.New[int](1024)

    // Producer goroutine with backpressure handling
    go func() {
        for i := 0; i < 100; i++ {
            for !buf.Push(i) {
                // Buffer full - yield to scheduler instead of busy-wait
                runtime.Gosched()
            }
        }
    }()

    // Consumer goroutine
    go func() {
        for {
            if val, ok := buf.Pop(); ok {
                fmt.Println(val)
            } else {
                // Buffer empty - yield to scheduler
                runtime.Gosched()
            }
        }
    }()
}
```

### Backpressure Strategies

When the buffer is full, avoid busy-waiting which wastes CPU cycles. Choose a strategy based on your latency requirements:

```go
// Strategy 1: Yield to scheduler (low CPU, microsecond latency)
for !buf.Push(item) {
    runtime.Gosched()
}

// Strategy 2: Exponential backoff (balanced approach)
backoff := time.Nanosecond
for !buf.Push(item) {
    time.Sleep(backoff)
    backoff = min(backoff*2, time.Millisecond)
}

// Strategy 3: Hybrid (spin briefly, then yield)
attempts := 0
for !buf.Push(item) {
    if attempts < 100 {
        // Spin for lowest latency
        attempts++
    } else {
        // Yield after threshold
        runtime.Gosched()
    }
}

// Strategy 4: Drop or handle differently (for real-time systems)
if !buf.Push(item) {
    // Log drop, sample, or handle overflow
    handleBackpressure(item)
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

    // Cap returns the total capacity of the ring buffer.
    Cap() int

    // Len returns the current number of elements in the buffer.
    Len() int

    // Available returns the number of free slots in the buffer.
    Available() int
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
