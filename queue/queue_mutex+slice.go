package queue

import "sync"

type MutexSliceQueue[T any] struct {
	v  []T
	mu sync.Mutex
}

func NewMutexSliceQueue[T any]() IQueue[T] {
	return &MutexSliceQueue[T]{v: make([]T, 0)}
}

func (q *MutexSliceQueue[T]) Enqueue(v T) {
	q.mu.Lock()
	q.v = append(q.v, v)
	q.mu.Unlock()
}

func (q *MutexSliceQueue[T]) Dequeue() (value T, ok bool) {
	q.mu.Lock()
	if len(q.v) == 0 {
		q.mu.Unlock()
		ok = false
		return
	}
	v := q.v[0]
	// FIXME 这里实现有问题，底层slice会被复用，slice在入队的时候被扩容，
	// 出队的时候，有底层存储被浪费
	//
	// 把这个buffer做成一个环形，或者直接用ring.Ring代替
	q.v = q.v[1:]
	q.mu.Unlock()
	return v, true
}

func (q *MutexSliceQueue[T]) Length() uint64 {
	q.mu.Lock()
	n := uint64(len(q.v))
	q.mu.Unlock()
	return n
}
