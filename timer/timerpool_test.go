package timer_test

import (
	"sync"
	"testing"
	"time"
)

type TimerPool struct {
	pool sync.Pool
}

func NewTimerPool() *TimerPool {
	return &TimerPool{
		pool: sync.Pool{
			New: func() interface{} {
				return time.NewTimer(0)
			},
		},
	}
}

func (p *TimerPool) Get(d time.Duration) *time.Timer {
	t := p.pool.Get().(*time.Timer)
	t.Reset(d)
	return t
}

func (p *TimerPool) Put(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	p.pool.Put(t)
}

func TestTimerPool(t *testing.T) {
	pool := NewTimerPool()
	for i := 0; i < 10000; i++ {
		timer := pool.Get(time.Millisecond)
		<-timer.C
		pool.Put(timer)
	}
}
