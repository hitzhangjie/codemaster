package queue

type ChanQueue struct {
	ch chan interface{}
}

func NewChanQueue(size int) IQueue {
	return &ChanQueue{
		ch: make(chan interface{}, size),
	}
}

func (c *ChanQueue) Enqueue(i interface{}) {
	c.ch <- i
}

func (c *ChanQueue) Dequeue() interface{} {
	select {
	case v := <-c.ch:
		return v
	default:
		return nil
	}
}

func (c *ChanQueue) Length() uint64 {
	return uint64(len(c.ch))
}
