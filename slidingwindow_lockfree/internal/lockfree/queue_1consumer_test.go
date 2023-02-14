package lockfree_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/hitzhangjie/codemaster/slidingwindow_lockfree/internal/lockfree"
)

func BenchmarkLockFreeQueue_1Consumer(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < 1024; p++ {
		parrallelism := p * gomaxprocs
		desc := fmt.Sprintf("parrallelism-%d", parrallelism)
		b.Run(desc, func(b *testing.B) {
			q := lockfree.NewQueue()
			go func() {
				for {
					q.Dequeue()
					time.Sleep(time.Millisecond * 10)
				}
			}()

			b.ResetTimer()
			b.SetParallelism(parrallelism)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					q.Enqueue(1)
				}
			})
		})
	}

}

func BenchmarkMutexSliceQueue_1Consumer(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < 1024; p++ {
		parrallelism := p * gomaxprocs
		desc := fmt.Sprintf("parrallelism-%d", parrallelism)
		b.Run(desc, func(b *testing.B) {
			q := newMutexQueue()
			go func() {
				for {
					q.Dequeue()
					time.Sleep(time.Millisecond * 10)
				}
			}()

			b.ResetTimer()
			b.SetParallelism(parrallelism)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					q.Enqueue(1)
				}
			})
		})
	}
}

func BenchmarkChanQueue_1Consumer(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < 1024; p++ {
		parrallelism := gomaxprocs * p
		desc := fmt.Sprintf("parrallelism-%d", parrallelism)
		b.Run(desc, func(b *testing.B) {
			q := newChanQueue(1024)
			go func() {
				for {
					q.Dequeue()
					time.Sleep(time.Millisecond * 10)
				}
			}()

			b.ResetTimer()
			b.SetParallelism(parrallelism)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					q.Enqueue(1)
				}
			})
		})
	}
}
