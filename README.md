<p align="center">
  <img src="logo.png" alt="grin" width="120"/>
</p>



<p align="center">
Lock-free ring buffers for Go: a Single Producer Single Consumer (SPSC) queue (`grin.New`) and a Multi Producer Single Consumer (MPSC) queue (`grin.NewManyToOne`). Zero-allocation, zero-mutex, low-latency communication between goroutines.
</p>

## Features

- **Lock-free**: Uses atomic operations instead of mutexes for maximum throughput
- **Zero allocation**: No heap allocations during Push/Pop operations
- **Cache-line optimized**: Prevents false sharing between producer and consumer
- **Type-safe**: Generic implementation using Go generics
- **High performance**: Up to 6x faster than channels for single-producer/single-consumer operations
- **Two flavors**: SPSC ring buffer (`New`) and MPSC ring buffer (`NewManyToOne`)

## Ring Buffer Options

| Constructor | Pattern | Description |
| --- | --- | --- |
| `grin.New[T](size)` | SPSC | Highest throughput when exactly one producer and one consumer. |
| `grin.NewManyToOne[T](size)` | MPSC | Many producers with a single consumer, similar to Agrona's `ManyToOneConcurrentArrayQueue`. |

## Benchmark Results

Benchmarks comparing grin (SPSC + MPSC) vs Go channels vs `container/ring` (AMD EPYC 7763, Go 1.25.5):

```
BenchmarkGrin_Push-4                	166426032	         7.249 ns/op	       0 B/op	       0 allocs/op
BenchmarkManyToOne_PushParallel-4   	51127701	        24.43 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_Push-4             	85801380	        13.22 ns/op	       8 B/op	       0 allocs/op

BenchmarkGrin_PushPop-4             	235495621	         5.139 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_PushPop-4          	76851436	        13.69 ns/op	       8 B/op	       0 allocs/op

BenchmarkGrin_Sequential-4          	 1614048	       744.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_Sequential-4       	 2350016	       510.4 ns/op	       0 B/op	       0 allocs/op

BenchmarkGrin_Wraparound-4          	233518587	         5.144 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_Wraparound-4       	79932855	        13.12 ns/op	       0 B/op	       0 allocs/op

BenchmarkGrin_FillDrain-4           	  404118	      3014 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_FillDrain-4        	  228778	      5048 ns/op	    2048 B/op	     256 allocs/op

BenchmarkGrin_LargeBuffer-4         	206543743	         5.820 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdRing_LargeBuffer-4      	76360908	        14.29 ns/op	       8 B/op	       0 allocs/op

BenchmarkChannel_Push-4             	14023837	        85.89 ns/op	       0 B/op	       0 allocs/op
BenchmarkChannel_PushPop-4          	38714640	        30.88 ns/op	       0 B/op	       0 allocs/op
BenchmarkChannel_Sequential-4       	  291084	      3974 ns/op	       0 B/op	       0 allocs/op
BenchmarkChannel_Wraparound-4       	39052872	        30.36 ns/op	       0 B/op	       0 allocs/op
BenchmarkChannel_FillDrain-4        	   76225	     15661 ns/op	       0 B/op	       0 allocs/op
BenchmarkChannel_LargeBuffer-4      	61520910	        19.32 ns/op	       0 B/op	       0 allocs/op
```

**Key Takeaways:**
- **grin (SPSC) vs Channels**: ~12x faster for Push, ~6x faster for PushPop
- **grin (MPSC)**: Provides lock-free many-producer support with zero allocations while remaining substantially faster than channels for contended writes
- **grin vs container/ring**: grin stays allocation-free and thread-safe; `container/ring` is not concurrent-safe and allocates on writes

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

`go get github.com/andrewwormald/grin`

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

// NewManyToOne creates a multi-producer, single-consumer ring buffer.
// Size must be a power of 2, otherwise it panics.
func NewManyToOne[T any](size int) *ManyToOne[T]
```

## Requirements

- Buffer size must be a power of 2 (enforced by panic)
- `New`: exactly one producer goroutine and one consumer goroutine (not safe for multiple producers)
- `NewManyToOne`: multiple producers, one consumer goroutine

## License

See [LICENSE](LICENSE) file for details.
