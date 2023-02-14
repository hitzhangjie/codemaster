package lockfree_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/hitzhangjie/codemaster/slidingwindow_lockfree/internal/lockfree"
)

const maxTimes = 5

func BenchmarkLockFreeQueue_1Consumer(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < maxTimes; p++ {
		desc := fmt.Sprintf("parrallelism-%d", p*gomaxprocs)
		b.Run(desc, func(b *testing.B) {
			q := lockfree.NewQueue()
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
					q.Enqueue(1)
				}
			})
			close(done)
		})
	}
}

func BenchmarkMutexSliceQueue_1Consumer(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < maxTimes; p++ {
		desc := fmt.Sprintf("parrallelism-%d", p*gomaxprocs)
		b.Run(desc, func(b *testing.B) {
			q := newMutexQueue()
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
					q.Enqueue(1)
				}
			})
			close(done)
		})
	}
}

func BenchmarkChanQueue_1Consumer(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < maxTimes; p++ {
		desc := fmt.Sprintf("parrallelism-%d", p*gomaxprocs)
		b.Run(desc, func(b *testing.B) {
			q := newChanQueue(512)
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
					q.Enqueue(1)
				}
			})
			close(done)
		})
	}
}
