package lockfree

import (
	"sync/atomic"
	"unsafe"
)

// Queue implements lock-free FIFO freelist based queue.
// ref: https://dl.acm.org/citation.cfm?doid=248052.248106
type Queue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
	len  uint64
}

// NewQueue creates a new lock-free queue.
func NewQueue() *Queue {
	// allocate a free item
	head := node{next: nil, v: nil}
	return &Queue{
		// both head and tail points to the free item
		tail: unsafe.Pointer(&head),
		head: unsafe.Pointer(&head),
	}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *Queue) Enqueue(v interface{}) {
	// allocate new item
	i := &node{next: nil, v: v}
	var last, lastnext *node
	for {
		last = load(&q.tail)
		lastnext = load(&last.next)
		// are tail and next consistent?
		if load(&q.tail) == last {
			// was tail pointing to the last node or not
			if lastnext == nil {
				// try to link item at the end of linked list
				if cas(&last.next, lastnext, i) {
					// enqueue is done. try swing tail to the inserted node
					cas(&q.tail, last, i)
					atomic.AddUint64(&q.len, 1)
					return
				}
			} else {
				// try swing tail to the next node
				cas(&q.tail, last, lastnext)
			}
		}
	}
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *Queue) Dequeue() interface{} {
	var first, last, firstnext *node
	for {
		first = load(&q.head)
		last = load(&q.tail)
		firstnext = load(&first.next)

		// if head, tail and next not consistent
		if first != load(&q.head) {
			continue
		}

		// is queue empty?
		if first == last {
			// queue is empty, couldn't dequeue
			if firstnext == nil {
				return nil
			}
			// tail is falling behind, try to advance it
			cas(&q.tail, last, firstnext)
		} else {
			// read value before cas, otherwise another dequeue might free the next node
			v := firstnext.v
			// try to swing head to the next node
			if cas(&q.head, first, firstnext) {
				atomic.AddUint64(&q.len, ^uint64(0))
				// queue was not empty and dequeue finished.
				return v
			}
		}
	}
}

// Length returns the length of the queue.
func (q *Queue) Length() uint64 {
	return atomic.LoadUint64(&q.len)
}
