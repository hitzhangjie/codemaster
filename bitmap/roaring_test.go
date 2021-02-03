package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/RoaringBitmap/roaring/roaring64"
)

func Test_RoaringBitmap(t *testing.T) {
	// example inspired by https://github.com/fzandona/goroar
	fmt.Println("==roaring==")
	rb1 := roaring.BitmapOf(1, 2, 3, 4, 5, 100, 1000)
	fmt.Println("rb1: ", rb1.String())

	rb2 := roaring.BitmapOf(3, 4, 1000)
	fmt.Println("rb2: ", rb2.String())

	rb3 := roaring.New()
	fmt.Println("rb3: ", rb3.String())

	fmt.Println("rb1 cardinality: ", rb1.GetCardinality())

	fmt.Println("rb1 contains 3? ", rb1.Contains(3))

	rb1.And(rb2)
	fmt.Println("rb1 intersect rb2 cardinality: ", rb1.GetCardinality())
	fmt.Println("rb1 intersect rb2: ", rb1.String())

	rb3.Add(1)
	rb3.Add(5)
	fmt.Println("rb3 cardinality: ", rb3.GetCardinality())
	fmt.Println("rb3: ", rb3.String())

	rb3.Or(rb1)
	fmt.Println("rb3 union rb1 cardinality: ", rb3.GetCardinality())
	fmt.Println("rb3 union rb2: ", rb3.String())

	// computes union of the three bitmaps in parallel using 4 workers
	roaring.ParOr(4, rb1, rb2, rb3)
	// computes intersection of the three bitmaps in parallel using 4 workers
	roaring.ParAnd(4, rb1, rb2, rb3)

	// prints 1, 3, 4, 5, 1000
	i := rb3.Iterator()
	for i.HasNext() {
		fmt.Println(i.Next())
	}
	fmt.Println()

	// next we include an example of serialization
	buf := new(bytes.Buffer)
	rb1.WriteTo(buf) // we omit error handling
	newrb := roaring.New()
	newrb.ReadFrom(buf)
	if rb1.Equals(newrb) {
		fmt.Println("I wrote the content to a byte stream and read it back.")
	}
	// you can iterate over bitmaps using ReverseIterator(), Iterator, ManyIterator()
}

func Test_RoaringBitmap64(t *testing.T) {
	// example inspired by https://github.com/fzandona/goroar
	fmt.Println("==roaring64==")
	rb1 := roaring64.BitmapOf(1, 2, 3, 4, 5, 100, 1000)
	fmt.Println(rb1.String())

	rb2 := roaring64.BitmapOf(3, 4, 1000)
	fmt.Println(rb2.String())

	rb3 := roaring64.New()
	fmt.Println(rb3.String())

	fmt.Println("Cardinality: ", rb1.GetCardinality())

	fmt.Println("Contains 3? ", rb1.Contains(3))

	rb1.And(rb2)

	rb3.Add(1)
	rb3.Add(5)

	rb3.Or(rb1)

	// prints 1, 3, 4, 5, 1000
	i := rb3.Iterator()
	for i.HasNext() {
		fmt.Println(i.Next())
	}
	fmt.Println()

	// next we include an example of serialization
	buf := new(bytes.Buffer)
	rb1.WriteTo(buf) // we omit error handling
	newrb := roaring64.New()
	newrb.ReadFrom(buf)
	if rb1.Equals(newrb) {
		fmt.Println("I wrote the content to a byte stream and read it back.")
	}
	// you can iterate over bitmaps using ReverseIterator(), Iterator, ManyIterator()
}
