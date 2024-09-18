package slices_test

import (
	"fmt"
	"slices"
	"testing"
)

func Test_slices(t *testing.T) {
	nums := []int{0, 1, 2, 3, 4, 5}
	nums = slices.Delete(nums, 5, 6)
	fmt.Println(nums)
}
