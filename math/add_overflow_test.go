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
		if c, err := safeAddSignedInt(a, b); err != nil {
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
		if c, err := safeAddUnsignedInt(a, b); err != nil {
			t.Errorf("iter-%d %v + %v = %v, err: %v", iter, a, b, c, err)
		} else {
			t.Logf("iter-%d %v + %v = %v", iter, a, b, c)
		}
	})
}

var errOverflow = errors.New("overflow")

func safeAddSignedInt[T ~int | ~int32 | ~int64](a, b T) (T, error) {
	c := a + b
	fmt.Fprintf(os.Stderr, "%v + %v = %v\n", a, b, c)
	// 按补码表示法表示的话，负数比正数多一个可表示的数字，再加上0，实际上maxint32+2==-2，不可能加成0
	if a > 0 && b > 0 && c < 0 ||
		a < 0 && b < 0 && c > 0 {
		return c, errOverflow
	}
	return c, nil
}

func safeAddUnsignedInt[T ~uint | ~uint32 | ~uint64](a, b T) (T, error) {
	c := a + b
	fmt.Fprintf(os.Stderr, "%v + %v = %v\n", a, b, c)
	if c < a || c < b {
		return c, errOverflow
	}
	return c, nil
}

func safeFloat32Add(a, b float32) (float32, error) {
	// 符号不同，不可能溢出
	if (a >= 0 && b <= 0) || (a <= 0 && b >= 0) {
		return a + b, nil
	}
	// 符号相同，检查下上限
	var a1 float32 = a
	var b1 float32 = b

	// 浮点数的表示是对称的，符号位+阶码+尾数，符号位不同则数值不同，但是正数、负数值域范围是一致的，所以也没有MinFloat32这种定义
	if a < 0 && b < 0 {
		a1 = -a
		b1 = -b
	}
	if math.MaxFloat32-a1 < b1 {
		return 0, errOverflow
	}
	return a + b, nil
}

func Test_add_float32(t *testing.T) {
	a := float32(math.MaxFloat32)
	b := float32(10000)
	_, err := safeFloat32Add(a, b)
	assert.NotNil(t, err)

	a = float32(-1 * math.MaxFloat32)
	b = float32(-1)
	_, err = safeFloat32Add(a, b)
	assert.NotNil(t, err)
}
