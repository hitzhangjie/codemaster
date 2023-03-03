package math_test

import (
	"errors"
	"fmt"
	"math"
	"os"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
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

/////////////////////////////////////////////////////////////////////

func safeFloat32Add_Version1(a, b float32) (float32, error) {
	// 联想下浮点数是如何相加的：
	// - 先把a、b的阶码（指数部分）调成一样的，比如按a的
	// - 然后把a、b的尾数相加一下，但是b在调整了解码后，尾数部分已经不再IEEE 754 float32的尾数bits范围内了，如果b比较小，很可能就被扔了
	c := float64(a) + float64(b)
	if math.Abs(c) > math.MaxFloat32 {
		return 0, errOverflow
	}
	fmt.Println("还差:", c-float64(a)-float64(b)) // 一做减法，预期为0，实际呢为 -b，说明前面的推断是对的，在float32表示中b直接被扔了
	return float32(c), nil
}

func safeFloat32Add(a, b float32) (float32, error) {
	c := a + b
	if a > 0 && b > 0 {
		if c > math.MaxFloat32 { // 这种肯定是overflow
			return 0, errOverflow
		} else if c == math.MaxFloat32 { // 这种有可能是因为尾数精度给丢弃了，实际上也超了
			v := c - a - b
			if v < 0 {
				return 0, errOverflow
			}
			return c, nil
		} else {
			return c, nil
		}
	}
	if a < 0 && b < 0 {
		if c < -1*math.MaxFloat32 { // 肯定是overflow，只不过是负数
			return 0, errOverflow
		} else if c == -1*math.MaxFloat32 { // 尾数部分有可能丢了，实际上也超了
			v := c - a - b
			if v > 0 {
				return 0, errOverflow
			}
			return c, nil
		} else {
			return c, nil
		}
	}
	return c, nil
}

func Test_add_float32(t *testing.T) {
	t.Run("not working version", func(t *testing.T) {
		a := float32(math.MaxFloat32)
		b := float32(10000) // not working, why? fraction part is dropped
		_, err := safeFloat32Add_Version1(a, b)
		assert.NotNil(t, err)

		a = float32(math.MaxFloat32)
		b = a
		_, err = safeFloat32Add_Version1(a, b)
		assert.NotNil(t, err)
	})
	t.Run("working version", func(t *testing.T) {
		a := float32(math.MaxFloat32)
		b := float32(10000)
		_, err := safeFloat32Add(a, b)
		assert.NotNil(t, err)

		a = float32(-1 * math.MaxFloat32)
		b = float32(-1)
		_, err = safeFloat32Add(a, b)
		assert.NotNil(t, err)
	})

}

func safeFloat64Add(a, b float64) (float64, error) {
	c := a + b
	if a > 0 && b > 0 {
		if c > math.MaxFloat64 { // 这种肯定是overflow
			return 0, errOverflow
		} else if c == math.MaxFloat64 { // 这种有可能是因为尾数精度给丢弃了，实际上也超了
			v := c - a - b
			if v < 0 {
				return 0, errOverflow
			}
			return c, nil
		} else {
			return c, nil
		}
	}
	if a < 0 && b < 0 {
		if c < -1*math.MaxFloat64 { // 肯定是overflow，只不过是负数
			return 0, errOverflow
		} else if c == -1*math.MaxFloat64 { // 尾数部分有可能丢了，实际上也超了
			v := c - a - b
			if v > 0 {
				return 0, errOverflow
			}
			return c, nil
		} else {
			return c, nil
		}
	}
	return c, nil
}
