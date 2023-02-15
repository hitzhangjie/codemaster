package queue

type ChanQueue[T any] struct {
	ch chan T
}

func NewChanQueue[T any](size int) IQueue[T] {
	return &ChanQueue[T]{
		ch: make(chan T, size),
	}
}

func (c *ChanQueue[T]) Enqueue(i T) {
	c.ch <- i
}

func (c *ChanQueue[T]) Dequeue() (value T, ok bool) {
	select {
	case v := <-c.ch:
		return v, true
	default:
		ok = false
		return
	}
}

func (c *ChanQueue[T]) Length() uint64 {
	return uint64(len(c.ch))
}
