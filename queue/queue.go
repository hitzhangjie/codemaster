package queue

type IQueue[T any] interface {
	Enqueue(T)
	Dequeue() (T, bool)
	Length() uint64
}
