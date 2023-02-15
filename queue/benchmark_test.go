package queue_test

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hitzhangjie/codemaster/queue"
)

///////////////////////////////////////////////////////////////////////////////
// benchmark lockfree / mutex+slice / chan queue
// 压测配置统一：
// - 协程数固定GOMAXPROCS
// - queue操作逻辑，测试时先写后读

// go test -bench=BenchmarkLockFreeQueue -count=5
// goos: darwin
// goarch: amd64
// cpu: Intel(R) Core(TM) i7-8569U CPU @ 2.80GHz
// BenchmarkLockFreeQueue-8        64692733                18.32 ns/op
// BenchmarkLockFreeQueue-8        64811620                18.19 ns/op
// BenchmarkLockFreeQueue-8        64799133                18.11 ns/op
// BenchmarkLockFreeQueue-8        62302936                18.33 ns/op
// BenchmarkLockFreeQueue-8        63553734                18.20 ns/op

// go test -bench=BenchmarkLockFreeQueue -count=5
// goos: darwin
// goarch: arm64
// BenchmarkLockFreeQueue-10               16280550                68.99 ns/op
// BenchmarkLockFreeQueue-10               18681440                72.15 ns/op
// BenchmarkLockFreeQueue-10               17397050                71.82 ns/op
// BenchmarkLockFreeQueue-10               23805744                73.15 ns/op
// BenchmarkLockFreeQueue-10               16105462                76.17 ns/op
func BenchmarkLockFreeQueue(b *testing.B) {
	length := 1 << 12
	inputs := make([]int, length)
	for i := 0; i < length; i++ {
		inputs = append(inputs, rand.Int()%2)
	}
	q := queue.NewLockfreeQueue()

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

// go test -bench=BenchmarkMutexSliceQueue -count=5
// goos: darwin
// goarch: amd64
// cpu: Intel(R) Core(TM) i7-8569U CPU @ 2.80GHz
// BenchmarkMutexSliceQueue-8      14319187                84.40 ns/op
// BenchmarkMutexSliceQueue-8      14222097                86.08 ns/op
// BenchmarkMutexSliceQueue-8      14041225                84.13 ns/op
// BenchmarkMutexSliceQueue-8      14324254                86.18 ns/op
// BenchmarkMutexSliceQueue-8      13946214                86.98 ns/op

// go test -bench=BenchmarkMutexSliceQueue -count=5
// goos: darwin
// goarch: arm64
// BenchmarkMutexSliceQueue-10                      8839629               137.5 ns/op
// BenchmarkMutexSliceQueue-10                      8811577               123.6 ns/op
// BenchmarkMutexSliceQueue-10                     10479646               126.7 ns/op
// BenchmarkMutexSliceQueue-10                     10064587               119.0 ns/op
// BenchmarkMutexSliceQueue-10                      9314046               117.4 ns/op
func BenchmarkMutexSliceQueue(b *testing.B) {
	length := 1 << 12
	inputs := make([]int, length)
	for i := 0; i < length; i++ {
		inputs = append(inputs, rand.Int()%2)
	}
	q := queue.NewMutexSliceQueue()

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

// go test -bench=BenchmarkChanQueue -count=5
// goos: darwin
// goarch: amd64
// cpu: Intel(R) Core(TM) i7-8569U CPU @ 2.80GHz
// BenchmarkChanQueue-8     3246128               373.2 ns/op
// BenchmarkChanQueue-8     3177400               462.6 ns/op
// BenchmarkChanQueue-8     2407434               448.4 ns/op
// BenchmarkChanQueue-8     2662528               456.8 ns/op
// BenchmarkChanQueue-8     2361782               451.4 ns/op

// go test -bench=BenchmarkChanQueue -count=5
// goos: darwin
// goarch: arm64
// BenchmarkChanQueue-10                   17382958                71.89 ns/op
// BenchmarkChanQueue-10                   17791318                75.35 ns/op
// BenchmarkChanQueue-10                   18194452                74.87 ns/op
// BenchmarkChanQueue-10                   19312411                74.94 ns/op
// BenchmarkChanQueue-10                   19402941                75.26 ns/op
func BenchmarkChanQueue(b *testing.B) {
	length := 1 << 12
	inputs := make([]int, length)
	for i := 0; i < length; i++ {
		inputs = append(inputs, rand.Int()%2)
	}
	q := queue.NewChanQueue(1024)
	go func() {
		for {
			q.Dequeue()
			time.Sleep(time.Microsecond * 10)
		}
	}()

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
