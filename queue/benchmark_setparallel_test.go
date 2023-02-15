package queue_test

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"

	"github.com/hitzhangjie/codemaster/queue"
)

const maxTimes = 5

// macbook 14, m1 pro, darwin+arm64
// 协程数从10~1000，平均耗时156~173ns，随着竞争加剧平均耗时有所增加，但是也是10ns级别的
func BenchmarkLockFreeQueue_withParallel(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < maxTimes; p++ {
		desc := fmt.Sprintf("parrallelism-%d", p*gomaxprocs)
		b.Run(desc, func(b *testing.B) {
			q := queue.NewLockfreeQueue[int]()
			b.ResetTimer()
			b.SetParallelism(p)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if rand.Int()%2 == 0 {
						q.Enqueue(1)
					} else {
						q.Dequeue()
					}
				}
			})
		})
	}
}

// macbook 14, m1 pro, darwin+arm64
// 协程数从10~1000，平均耗时153~210ns，随着竞争加剧平均耗时有所增加，但是也是10ns级别的
func BenchmarkMutexSliceQueue_withParallel(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < maxTimes; p++ {
		desc := fmt.Sprintf("parrallelism-%d", p*gomaxprocs)
		b.Run(desc, func(b *testing.B) {
			q := queue.NewMutexSliceQueue[int]()
			b.ResetTimer()
			b.SetParallelism(p)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if rand.Int()%2 == 0 {
						q.Enqueue(1)
					} else {
						q.Dequeue()
					}

				}
			})
		})
	}
}

// macbook 14, m1 pro, darwin+arm64
// 协程数从10~1000，平均耗时150~175ns，随着竞争加剧平均耗时有所增加，但是也是10ns级别的
//
// 现在这个workload下，chan这个效果杠杠的，跟lockfree没什么区别
func BenchmarkChanQueue_withParallel(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < maxTimes; p++ {
		desc := fmt.Sprintf("parrallelism-%d", p*gomaxprocs)
		b.Run(desc, func(b *testing.B) {
			q := queue.NewChanQueue[int](512)
			done := make(chan int, 1)
			go func() {
				for {
					q.Dequeue()
					select {
					case <-done:
						return
					default:
					}
				}
			}()
			b.ResetTimer()
			b.SetParallelism(p)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if rand.Int()%2 == 0 {
						q.Enqueue(1)
					} else {
						q.Dequeue()
					}
				}
			})
			close(done)
		})
	}
}
