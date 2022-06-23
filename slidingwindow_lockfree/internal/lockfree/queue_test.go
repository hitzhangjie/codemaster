package lockfree_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hitzhangjie/codemaster/slidingwindow_lockfree/internal/lockfree"
)

func TestQueueDequeueEmpty(t *testing.T) {
	q := lockfree.NewQueue()
	if q.Dequeue() != nil {
		t.Fatalf("dequeue empty queue returns non-nil")
	}
}

func TestQueue_Length(t *testing.T) {
	q := lockfree.NewQueue()
	if q.Length() != 0 {
		t.Fatalf("empty queue has non-zero length")
	}

	q.Enqueue(1)
	if q.Length() != 1 {
		t.Fatalf("count of enqueue wrong, want %d, got %d.", 1, q.Length())
	}

	q.Dequeue()
	if q.Length() != 0 {
		t.Fatalf("count of dequeue wrong, want %d, got %d", 0, q.Length())
	}
}

func ExampleQueue() {
	q := lockfree.NewQueue()

	q.Enqueue("1st item")
	q.Enqueue("2nd item")
	q.Enqueue("3rd item")

	fmt.Println(q.Dequeue())
	fmt.Println(q.Dequeue())
	fmt.Println(q.Dequeue())

	// Output:
	// 1st item
	// 2nd item
	// 3rd item
}

type queueInterface interface {
	Enqueue(interface{})
	Dequeue() interface{}
}

type mutexQueue struct {
	v  []interface{}
	mu sync.Mutex
}

func newMutexQueue() *mutexQueue {
	return &mutexQueue{v: make([]interface{}, 0)}
}

func (q *mutexQueue) Enqueue(v interface{}) {
	q.mu.Lock()
	q.v = append(q.v, v)
	q.mu.Unlock()
}

func (q *mutexQueue) Dequeue() interface{} {
	q.mu.Lock()
	if len(q.v) == 0 {
		q.mu.Unlock()
		return nil
	}
	v := q.v[0]
	q.v = q.v[1:]
	q.mu.Unlock()
	return v
}

type chanQueue struct {
	ch chan interface{}
}

func newChanQueue() *chanQueue {
	return &chanQueue{
		ch: make(chan interface{}, 1024),
	}
}

func (c *chanQueue) Enqueue(i interface{}) {
	c.ch <- i
}

func (c *chanQueue) Dequeue() interface{} {
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*100)
	select {
	case v := <-c.ch:
		return v
	case <-ctx.Done():
		return nil
	default:
		return nil
	}
}

//go test -bench=BenchmarkLockFreeQueue -count=5
//goos: darwin
//goarch: amd64
//cpu: Intel(R) Core(TM) i7-8569U CPU @ 2.80GHz
//BenchmarkLockFreeQueue-8        64692733                18.32 ns/op
//BenchmarkLockFreeQueue-8        64811620                18.19 ns/op
//BenchmarkLockFreeQueue-8        64799133                18.11 ns/op
//BenchmarkLockFreeQueue-8        62302936                18.33 ns/op
//BenchmarkLockFreeQueue-8        63553734                18.20 ns/op
func BenchmarkLockFreeQueue(b *testing.B) {
	length := 1 << 12
	inputs := make([]int, length)
	for i := 0; i < length; i++ {
		inputs = append(inputs, rand.Int()%2)
	}
	q := lockfree.NewQueue()

	var c int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := int(atomic.AddInt64(&c, 1)-1) % length
			v := inputs[i]
			if v == 1 {
				q.Enqueue(v)
			} else {
				q.Dequeue()
			}
		}
	})
}

//go test -bench=BenchmarkMutexSliceQueue -count=5
//goos: darwin
//goarch: amd64
//cpu: Intel(R) Core(TM) i7-8569U CPU @ 2.80GHz
//BenchmarkMutexSliceQueue-8      14319187                84.40 ns/op
//BenchmarkMutexSliceQueue-8      14222097                86.08 ns/op
//BenchmarkMutexSliceQueue-8      14041225                84.13 ns/op
//BenchmarkMutexSliceQueue-8      14324254                86.18 ns/op
//BenchmarkMutexSliceQueue-8      13946214                86.98 ns/op
func BenchmarkMutexSliceQueue(b *testing.B) {
	length := 1 << 12
	inputs := make([]int, length)
	for i := 0; i < length; i++ {
		inputs = append(inputs, rand.Int()%2)
	}
	q := newMutexQueue()

	var c int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := int(atomic.AddInt64(&c, 1)-1) % length
			v := inputs[i]
			if v == 1 {
				q.Enqueue(v)
			} else {
				q.Dequeue()
			}
		}
	})
}

//go test -bench=BenchmarkChanQueue -count=5
//goos: darwin
//goarch: amd64
//cpu: Intel(R) Core(TM) i7-8569U CPU @ 2.80GHz
//BenchmarkChanQueue-8     3246128               373.2 ns/op
//BenchmarkChanQueue-8     3177400               462.6 ns/op
//BenchmarkChanQueue-8     2407434               448.4 ns/op
//BenchmarkChanQueue-8     2662528               456.8 ns/op
//BenchmarkChanQueue-8     2361782               451.4 ns/op
func BenchmarkChanQueue(b *testing.B) {
	length := 1 << 12
	inputs := make([]int, length)
	for i := 0; i < length; i++ {
		inputs = append(inputs, rand.Int()%2)
	}
	q := newChanQueue()

	var c int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := int(atomic.AddInt64(&c, 1)-1) % length
			v := inputs[i]
			if v == 1 {
				q.Enqueue(v)
			} else {
				q.Dequeue()
			}
		}
	})
}
