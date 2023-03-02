package math_test

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
)

func Fuzz_overflow_int32(f *testing.F) {
	f.Add(int32(0x7fffffff), int32(0))
	f.Add(int32(-1<<31), int32(0))
	f.Fuzz(func(t *testing.T, a, b int32) {
		if c, err := safeSignedAdd(a, b); err != nil {
			t.Errorf("%v + %v = %v, err: %v", a, b, c, err)
		}
	})
}

func Fuzz_overflow_uint32(f *testing.F) {
	// 此输入下，mutator无能为力，执行了几分钟也发现不了问题，如果知道mutator原理的就很容易明白为什么。
	// 然后，进一步可以纠正下错误的测试思想，模糊测试!=漫无目的的测试，seed scorpus matters!
	//
	//f.Add(uint32(0xffffffff-1000), uint32(0))
	f.Add(uint32(0xffffffff), uint32(0))
	var count atomic.Int32
	f.Fuzz(func(t *testing.T, a, b uint32) {
		iter := count.Add(1)
		if c, err := safeUnsignedAdd(a, b); err != nil {
			t.Errorf("iter-%d %v + %v = %v, err: %v", iter, a, b, c, err)
		} else {
			t.Logf("iter-%d %v + %v = %v", iter, a, b, c)
		}
	})
}

var errOverflow = errors.New("overflow")

func safeSignedAdd[T ~int | ~int32 | ~int64](a, b T) (T, error) {
	c := a + b
	fmt.Fprintf(os.Stderr, "%v + %v = %v\n", a, b, c)
	if a > 0 && b > 0 && c <= 0 ||
		a < 0 && b < 0 && c >= 0 {
		return c, errOverflow
	}
	return c, nil
}

func safeUnsignedAdd[T ~uint | ~uint32 | ~uint64](a, b T) (T, error) {
	c := a + b
	fmt.Fprintf(os.Stderr, "%v + %v = %v\n", a, b, c)
	if c < a || c < b {
		return c, errOverflow
	}
	return c, nil
}
