package std

import (
	"container/ring"
	"fmt"
	"testing"
)

func Test_Ring(t *testing.T) {
	// ring 1, 1->2->3
	r := ring.New(3)
	r.Value = 1

	r2 := r.Next()
	r2.Value = 2

	r3 := r2.Next()
	r3.Value = 3

	fmt.Println("len:", r.Len())

	// ring2, 4->5->6
	rr := ring.New(3)
	rr.Value = 4

	rr2 := rr.Next()
	rr2.Value = 5

	rr3 := rr2.Next()
	rr3.Value = 6

	fmt.Println("len:", rr.Len())

	// ring link, 1->4->5->6->2->3

	r.Link(rr)
	fmt.Println("link ring el:", r.Value)

	for p := r.Next(); p != r; p = p.Next() {
		fmt.Println("link ring el:", p.Value)
	}

	// ring unlink, 1->2->3

	r.Unlink(3)
	fmt.Println("unlink ring el:", r.Value)

	for p := r.Next(); p != r; p = p.Next() {
		fmt.Println("unlink ring el:", p.Value)
	}

}
