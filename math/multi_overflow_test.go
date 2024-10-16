package math_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 乘法溢出的一个案例，攻击者盗刷道具的一个场景：
// 实际上按玩家当前拥有的金币，只能兑换少量的道具，但是它故意填了一个超大的数字，使得乘法溢出，结果乘法结果和正常情况下的乘法结果一样，
// 通过这种情况不当获得了很多道具 …… 由这个case引申出来，对乘法溢出的思考
func Test_multiply_overflow(t *testing.T) {
	num := 307445734561825861
	price := 60

	totalPrice := num * price

	if totalPrice < price {
		panic("overflow")
	}
	fmt.Println(uint64(totalPrice))
}

// 有符号数的乘法
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

// 无符号数的乘法
func safeMultiUnsignedInt[T ~uint | uint32 | uint64](a, b T) (T, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}

	result := a * b
	if result/b != a {
		return 0, fmt.Errorf("integer overflow")
	}

	return result, nil
}

func Test_multiply_uint_overflow(t *testing.T) {
	type testcase struct {
		a, b     uint64
		overflow bool
	}

	testcases := []testcase{
		{1, 1, false},
		{0, math.MaxUint64, false},
		{2, math.MaxUint64, true},
		{math.MaxUint64, math.MaxUint64, true},
	}

	for _, tt := range testcases {
		_, err := safeMultiUnsignedInt(tt.a, tt.b)
		if (err != nil) != tt.overflow {
			t.Fatalf("%d * %d = ?, want overflow = %v, got = %v", tt.a, tt.b, tt.overflow, err != nil)
		} else {
			t.Logf("%d * %d = ?, want overflow = %v, got = %v (err = %v)", tt.a, tt.b, tt.overflow, err != nil, err)
		}
	}
}

// 浮点数的乘法
func safeMultiFloat32[T ~float32 | float64](a, b T) (T, error) {
	result := a * b

	if math.IsInf(float64(result), 0) {
		return 0, fmt.Errorf("float overflow")
	}

	if math.IsNaN(float64(result)) {
		return 0, fmt.Errorf("result is NaN")
	}

	return result, nil
}

func Test_multiply_float32_overflow(t *testing.T) {
	type testcase struct {
		a, b     float32
		overflow bool
	}

	testcases := []testcase{
		{1, 1, false},
		{2, math.MaxFloat32, true},
		{-1.5, math.MaxFloat32, true},
		{-1.5, -math.MaxFloat32, true},
		{math.MaxFloat32, -math.MaxFloat32, true},
	}

	for _, tt := range testcases {
		_, err := safeMultiFloat32(tt.a, tt.b)
		if (err != nil) != tt.overflow {
			t.Fatalf("%.2f * %.2f = ?, want overflow = %v, got = %v", tt.a, tt.b, tt.overflow, err != nil)
		} else {
			t.Logf("%.2f * %.2f = ?, want overflow = %v, got = %v (err = %v)", tt.a, tt.b, tt.overflow, err != nil, err)
		}
	}
}
