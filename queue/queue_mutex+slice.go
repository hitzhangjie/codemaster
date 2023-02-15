package queue

import "sync"

type MutexSliceQueue struct {
	v  []interface{}
	mu sync.Mutex
}

func NewMutexSliceQueue() IQueue {
	return &MutexSliceQueue{v: make([]interface{}, 0)}
}

func (q *MutexSliceQueue) Enqueue(v interface{}) {
	q.mu.Lock()
	q.v = append(q.v, v)
	q.mu.Unlock()
}

func (q *MutexSliceQueue) Dequeue() interface{} {
	q.mu.Lock()
	if len(q.v) == 0 {
		q.mu.Unlock()
		return nil
	}
	v := q.v[0]
	q.v = q.v[1:]
	q.mu.Unlock()
	return v
}

func (q *MutexSliceQueue) Length() uint64 {
	q.mu.Lock()
	n := uint64(len(q.v))
	q.mu.Unlock()
	return n
}
