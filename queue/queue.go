package queue

type IQueue interface {
	Enqueue(interface{})
	Dequeue() interface{}
	Length() uint64
}
