package queue_test

import (
	"fmt"
	"runtime"
	"testing"

	lockfree3rd "github.com/bruceshao/lockfree/lockfree"
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

// 这个实现号称快的多：https://github.com/bruceshao/lockfree
//
// 实测效果确实不错，耗时稳定在100ns上下。
func BenchmarkLockFreeQueue_3rdparty(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < maxTimes; p++ {
		desc := fmt.Sprintf("parrallelism-%d", p*gomaxprocs)
		b.Run(desc, func(b *testing.B) {
			h := &longEventHandler[uint64]{}
			q := lockfree3rd.NewSerialDisruptor[uint64](1024*1024, h, &lockfree3rd.SchedWaitStrategy{})
			if err := q.Start(); err != nil {
				panic(err)
			}
			defer q.Close()
			producer := q.Producer()

			b.ResetTimer()
			b.SetParallelism(p)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_ = producer.Write(1)
				}
			})
		})
	}
}

type longEventHandler[T uint64] struct {
}

func (h *longEventHandler[T]) OnEvent(v uint64) {
	//fmt.Printf("value = %v\n", v)
}
