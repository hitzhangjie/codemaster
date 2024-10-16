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

func safeMultiSignedInt[T ~int | int32 | int64](a, b T) (T, error) {

	if a == 0 || b == 0 {
		return 0, nil
	}

	result := T(int64(a) * int64(b))

	if (a > 0 && b > 0 && result < 0) || (a < 0 && b < 0 && result < 0) || (a > 0 && b < 0 && result > 0) || (a < 0 && b > 0 && result > 0) {
		return 0, fmt.Errorf("integer overflow")
	}

	if a != 0 && result/a != b {
		return 0, fmt.Errorf("integer overflow")
	}

	return result, nil
}

func TestSafeMultiSignedInt(t *testing.T) {
	tests := []struct {
		name     string
		a        int64
		b        int64
		expected int64
		hasError bool
	}{
		{"Positive multiplication", 5, 7, 35, false},
		{"Negative multiplication", -3, 4, -12, false},
		{"Zero multiplication", 0, 10, 0, false},
		{"Overflow positive", math.MaxInt64, 2, 0, true},
		{"Overflow negative", math.MinInt64, 2, 0, true},
		{"No overflow at limits", math.MaxInt64, 1, math.MaxInt64, false},
		{"No overflow at negative limits", math.MinInt64, 1, math.MinInt64, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := safeMultiSignedInt(tt.a, tt.b)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func safeMultiUnsignedInt[T ~uint | uint32 | uint64](a, b T) (T, error) {
	return 0, nil
}

func safeMultiFloat32[T ~float32 | float64](a, b T) (T, error) {
	return 0, nil
}
