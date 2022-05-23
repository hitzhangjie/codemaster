package slidingwindow

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlidingWindow(t *testing.T) {
	winsz := time.Millisecond * 100
	// new slidingwindow
	w := NewSlidingWindow(winsz)
	assert.NotNil(t, w.curr)
	assert.NotNil(t, w.prev)
	assert.Equal(t, winsz, w.size)

	// count
	assert.Zero(t, w.Count())

	// record
	w.Record()
	time.Sleep(winsz / 2)
	assert.Equal(t, int64(1), w.Count())

	// record
	time.Sleep(winsz * 2)
	count := 5
	for i := 0; i < count; i++ {
		w.Record()
	}
	time.Sleep(winsz / 2)
	assert.Equal(t, int64(count), w.Count())
}

var sizes = []time.Duration{
	time.Millisecond * 100,
	time.Millisecond * 500,
	time.Second,
}

//BenchmarkSlidingWindow_Record/window-100ms-8         	 4267891	       332.0 ns/op
//BenchmarkSlidingWindow_Record/window-500ms-8         	 2454961	       734.1 ns/op
//BenchmarkSlidingWindow_Record/window-1s-8            	 1486333	       847.1 ns/op
func BenchmarkSlidingWindow_Record(b *testing.B) {
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("window-%v", sz), func(b *testing.B) {
			w := NewSlidingWindow(sz)
			for i := 0; i < b.N; i++ {
				w.Record()
			}
		})
	}
}

//BenchmarkSlidingWindow_RecordCount/window-100ms-8         	 4047170	       309.0 ns/op
//BenchmarkSlidingWindow_RecordCount/window-500ms-8         	 2068082	       528.3 ns/op
//BenchmarkSlidingWindow_RecordCount/window-1s-8            	 1000000	      1348 ns/op
func BenchmarkSlidingWindow_RecordCount(b *testing.B) {
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("window-%v", sz), func(b *testing.B) {
			w := NewSlidingWindow(sz)
			for i := 0; i < b.N; i++ {
				if i%10 != 0 {
					w.Record()
				} else {
					w.Count()
				}
			}
		})
	}
}

//BenchmarkSlidingWindow_RecordParrallel-8   	 4655830	       242.9 ns/op
func BenchmarkSlidingWindow_RecordParrallel(b *testing.B) {
	w := NewSlidingWindow(time.Second)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.Record()
		}
	})
}

//BenchmarkSlidingWindow_RecordCountParrallel-8   	63516374	        21.28 ns/op
func BenchmarkSlidingWindow_RecordCountParrallel(b *testing.B) {
	length := 1 << 12
	inputs := make([]int, length)
	for i := 0; i < length; i++ {
		inputs = append(inputs, rand.Int()%10)
	}

	var c int64
	w := NewSlidingWindow(time.Second)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := int(atomic.AddInt64(&c, 1)-1) % length
			v := inputs[i]
			if v != 0 {
				w.Record()
			} else {
				w.Count()
			}
		}
	})
}
