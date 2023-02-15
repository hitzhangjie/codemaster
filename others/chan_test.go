package chan_test

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestXxx(t *testing.T) {

	var (
		t1_10us     = uint64(0) // 1-10微秒
		t10_100us   = uint64(0) // 10-100微秒
		t100_1000us = uint64(0) // 100-1000微秒
		t1_10ms     = uint64(0) // 1-10毫秒
		t10_100ms   = uint64(0) // 10-100毫秒
		t100_ms     = uint64(0) // 大于100毫秒
	)

	var (
		length     = 1024 * 1024
		goroutines = 100
		elsPerGo   = 10000
		totalCount = uint64(0)
		slowCount  = uint64(0)
	)
	ch := make(chan uint64, length)

	// 消费端
	go func() {
		var ts time.Time
		var count int32
		for {
			x := <-ch
			atomic.AddInt32(&count, 1)
			if count == 1 {
				ts = time.Now()
			}
			if x%100000 == 0 {
				fmt.Printf("read %d\n", x)
			}
			if count == int32(goroutines*elsPerGo) {
				tl := time.Since(ts)
				fmt.Printf("read time = %d ms\n", tl.Milliseconds())
			}
		}
	}()

	begin := time.Now()

	// 生产端
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < elsPerGo; j++ {
				x := atomic.AddUint64(&totalCount, 1)
				ts := time.Now()
				ch <- x
				us := time.Since(ts).Microseconds()
				if us > 1 {
					atomic.AddUint64(&slowCount, 1)
					if us < 10 { // t1_10us
						atomic.AddUint64(&t1_10us, 1)
					} else if us < 100 {
						atomic.AddUint64(&t10_100us, 1)
					} else if us < 1000 {
						atomic.AddUint64(&t100_1000us, 1)
					} else if us < 10000 {
						atomic.AddUint64(&t1_10ms, 1)
					} else if us < 100000 {
						atomic.AddUint64(&t10_100ms, 1)
					} else {
						atomic.AddUint64(&t100_ms, 1)
					}
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	timecost := time.Since(begin)
	fmt.Println(strings.Repeat("-", 64))
	fmt.Printf("write total time = [%d ms]\n", timecost.Milliseconds())
	fmt.Println()
	fmt.Printf("slow ratio = %.2f \n", float64(slowCount)*100.0/float64(totalCount))
	fmt.Printf("quick ratio = %.2f \n", float64(goroutines*elsPerGo-int(slowCount))*100.0/float64(goroutines*elsPerGo))
	fmt.Println()
	fmt.Printf("[<1us][%d] \n", totalCount-slowCount)
	fmt.Printf("[1-10us][%d] \n", t1_10us)
	fmt.Printf("[10-100us][%d] \n", t10_100us)
	fmt.Printf("[100-1000us][%d] \n", t100_1000us)
	fmt.Printf("[1-10ms][%d] \n", t1_10ms)
	fmt.Printf("[10-100ms][%d] \n", t10_100ms)
	fmt.Printf("[>100ms][%d] \n", t100_ms)
}

func BenchmarkXxx(b *testing.B) {
	ch := make(chan int, 1)
	done := make(chan int, 1)

	go func() {
		for {
			<-ch
			select {
			case <-done:
				return
			default:
			}
		}
	}()

	b.ResetTimer()
	b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch <- 1
		}
	})

	close(ch)
}
