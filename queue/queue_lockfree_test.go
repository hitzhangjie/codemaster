package queue_test

import (
	"fmt"
	"testing"

	"github.com/hitzhangjie/codemaster/queue"
)

func TestQueueDequeueEmpty(t *testing.T) {
	q := queue.NewLockfreeQueue()
	if q.Dequeue() != nil {
		t.Fatalf("dequeue empty queue returns non-nil")
	}
}

func TestQueue_Length(t *testing.T) {
	q := queue.NewLockfreeQueue()
	if q.Length() != 0 {
		t.Fatalf("empty queue has non-zero length")
	}

	q.Enqueue(1)
	if q.Length() != 1 {
		t.Fatalf("count of enqueue wrong, want %d, got %d.", 1, q.Length())
	}

	q.Dequeue()
	if q.Length() != 0 {
		t.Fatalf("count of dequeue wrong, want %d, got %d", 0, q.Length())
	}
}

func ExampleQueue() {
	q := queue.NewLockfreeQueue()

	q.Enqueue("1st item")
	q.Enqueue("2nd item")
	q.Enqueue("3rd item")

	fmt.Println(q.Dequeue())
	fmt.Println(q.Dequeue())
	fmt.Println(q.Dequeue())

	// Output:
	// 1st item
	// 2nd item
	// 3rd item
}
