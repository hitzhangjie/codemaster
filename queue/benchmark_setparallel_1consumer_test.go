package queue_test

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"

	lockfree3rd "github.com/bruceshao/lockfree/lockfree"
	"github.com/hitzhangjie/codemaster/queue"
	disruptor "github.com/smartystreets-prototypes/go-disruptor"
)

// macbook 14, m1 pro, darwin+arm64
// 协程数从10~1000，平均耗时380~498ns，随着竞争加剧平均耗时有所增加
//
// 这个性能之所以比较差，可能原因主要在于：
// - Yes 调度器方面：避免一直cas+spin，实时gosched，从400、500ns，直接干到200ns
// - No  内存使用方面，enqueue对象每次分配对象，GC开销大停顿延迟1ms：syncpool优化不明显
// - No  interface转换，改成泛型实现：但是效果不明显
func BenchmarkLockFreeQueue_1Consumer(b *testing.B) {
	//debug.SetGCPercent(-1)
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < 3; p++ {
		desc := fmt.Sprintf("parrallelism-%d", p*gomaxprocs)
		b.Run(desc, func(b *testing.B) {
			q := queue.NewLockfreeQueue[int]()
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
			q := queue.NewMutexSliceQueue[int]()
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

// LMAX Disruptor porting to golang：go-disruptor
//
// 这个实现的速度就很快，比上面这个(https://github.com/bruceshao/lockfree)速度还要快很多
// reading more: https://lmax-exchange.github.io/disruptor/
const buffersize = 1024 * 1024

var buffermask = buffersize - 1
var sequence int64 = 0
var ringbuffer = [buffersize]int{}

func BenchmarkLockFreeQueue_godisruptor(b *testing.B) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	for p := 1; p < maxTimes; p++ {
		desc := fmt.Sprintf("parrallelism-%d", p*gomaxprocs)
		b.Run(desc, func(b *testing.B) {

			d := disruptor.New(
				disruptor.WithCapacity(int64(buffersize)),
				disruptor.WithConsumerGroup(MyConsumer{}))

			go d.Read()

			b.ResetTimer()
			b.SetParallelism(p)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					reservation := d.Reserve(1)
					ringbuffer[int(sequence)&buffermask] = 1
					atomic.AddInt64(&sequence, 1)
					d.Commit(reservation, reservation)
				}
			})
			d.Close()
		})
	}
}

type MyConsumer struct{}

func (m MyConsumer) Consume(lowerSequence, upperSequence int64) {
	for sequence := lowerSequence; sequence <= upperSequence; sequence++ {
		index := sequence & int64(buffermask)
		message := ringbuffer[index]
		_ = message
		//fmt.Println(message)
	}
}
