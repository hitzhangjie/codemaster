package queue_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/hitzhangjie/codemaster/queue"
)

// macbook 14, m1 pro, darwin+arm64
// 协程数从10~1000，平均耗时380~498ns，随着竞争加剧平均耗时有所增加
func BenchmarkLockFreeQueue_1Consumer(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < maxTimes; p++ {
		desc := fmt.Sprintf("parrallelism-%d", p*gomaxprocs)
		b.Run(desc, func(b *testing.B) {
			q := queue.NewLockfreeQueue()
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

// macbook 14, m1 pro, darwin+arm64
// 协程数从10~1000，平均耗时145~148ns，随着竞争加剧平均耗时没有明显增加，但是内存分配次数有增加，跟slice操作有关系
func BenchmarkMutexSliceQueue_1Consumer(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < maxTimes; p++ {
		desc := fmt.Sprintf("parrallelism-%d", p*gomaxprocs)
		b.Run(desc, func(b *testing.B) {
			q := queue.NewMutexSliceQueue()
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

// macbook 14, m1 pro, darwin+arm64
// 协程数从10~1000，平均耗时115~564ns，随着竞争加剧平均耗时有所增加，
//
// 但是绝大部分时候（包括协程最多时）耗时在500ms以下，超过490ms的极少，
// 从这个意义上说，消费者为1，生产者写消息百万级别以上（生产者数量从10到1000），
// chan表现的至少和当前的lockfree实现没特别大差异。
func BenchmarkChanQueue_1Consumer(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < maxTimes; p++ {
		desc := fmt.Sprintf("parrallelism-%d", p*gomaxprocs)
		b.Run(desc, func(b *testing.B) {
			q := queue.NewChanQueue(512)
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
