package queue

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

// LockFreeQueue implements lock-free FIFO freelist based queue.
// ref: https://dl.acm.org/citation.cfm?doid=248052.248106
type LockFreeQueue[T any] struct {
	head unsafe.Pointer
	tail unsafe.Pointer
	len  uint64

	pool sync.Pool
}

// NewLockfreeQueue creates a new lock-free queue.
func NewLockfreeQueue[T any]() IQueue[T] {
	// allocate a free item
	head := node[T]{next: nil}
	return &LockFreeQueue[T]{
		// both head and tail points to the free item
		tail: unsafe.Pointer(&head),
		head: unsafe.Pointer(&head),
		pool: sync.Pool{New: func() any {
			return &node[T]{}
		}},
	}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *LockFreeQueue[T]) Enqueue(v T) {
	// allocate new item
	//el := &node[T]{next: nil, v: v}
	el := q.pool.New().(*node[T])
	el.next = nil
	el.v = v
	var last, lastnext *node[T]
	failed := 0
	for {
		last = load[T](&q.tail)
		lastnext = load[T](&last.next)
		// are tail and next consistent?
		if load[T](&q.tail) == last {
			// was tail pointing to the last node or not
			if lastnext == nil {
				// try to link item at the end of linked list
				if cas(&last.next, lastnext, el) {
					// enqueue is done. try swing tail to the inserted node
					cas(&q.tail, last, el)
					atomic.AddUint64(&q.len, 1)
					return
				}
			} else {
				// try swing tail to the next node
				cas(&q.tail, last, lastnext)
			}
		}
		failed++
		if failed > 1 {
			runtime.Gosched()
			failed = 0
		}
	}
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *LockFreeQueue[T]) Dequeue() (value T, ok bool) {
	var first, last, firstnext *node[T]
	for {
		first = load[T](&q.head)
		last = load[T](&q.tail)
		firstnext = load[T](&first.next)

		// if head, tail and next not consistent
		if first != load[T](&q.head) {
			continue
		}

		// is queue empty?
		if first == last {
			// queue is empty, couldn't dequeue
			if firstnext == nil {
				ok = false
				return
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
				q.pool.Put(first)
				return v, true
			}
		}
	}
}

// Length returns the length of the queue.
func (q *LockFreeQueue[T]) Length() uint64 {
	return atomic.LoadUint64(&q.len)
}

type node[T any] struct {
	next unsafe.Pointer
	v    T
}

func load[T any](p *unsafe.Pointer) *node[T] {
	return (*node[T])(atomic.LoadPointer(p))
}

func cas[T any](p *unsafe.Pointer, old, new *node[T]) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
