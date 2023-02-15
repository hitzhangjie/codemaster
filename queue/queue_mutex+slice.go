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
