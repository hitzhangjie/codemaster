package math_test

import (
	"errors"
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
	f.Add(uint32(0xffffffff), uint32(0))
	f.Fuzz(func(t *testing.T, a, b uint32) {
		if c, err := safeUnsignedAdd(a, b); err != nil {
			t.Errorf("%v + %v = %v, err: %v", a, b, c, err)
		}
	})
}

var errOverflow = errors.New("overflow")

func safeSignedAdd[T ~int | ~int32 | ~int64](a, b T) (T, error) {
	c := a + b
	if a > 0 && b > 0 && c <= 0 ||
		a < 0 && b < 0 && c >= 0 {
		return c, errOverflow
	}
	return c, nil
}

func safeUnsignedAdd[T ~uint | ~uint32 | ~uint64](a, b T) (T, error) {
	c := a + b
	if c < a || c < b {
		return c, errOverflow
	}
	return c, nil
}
