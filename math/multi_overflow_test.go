package math_test

import (
	"fmt"
	"testing"
)

func Test_multiply_overflow(t *testing.T) {
	num := 307445734561825861
	price := 60

	totalPrice := num * price

	if totalPrice < price {
		panic("overflow")
	}

	fmt.Println(uint64(totalPrice))
}
