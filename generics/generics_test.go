package generics_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 泛型类型，T是类型参数，any表示限制接口，泛型实例化的参数要实现any这个限制接口
type Lockable[T any] struct {
	mu   sync.Mutex
	data T
}

// func (l *Lockable[T]) Add(v T) {
// 	l.mu.Lock()
// 	l.data += v // 这样不行，+= 操作不符合类型参数T的约束条件，约束条件any并没有要求实现+=这个操作，所以要换种写法
//              // 方法1： 一种是下下面那样我们定义个方法DoAdd接受一个参数为*T的函数去操作指针
//              // 方法2：我们要调整下这里的限制，告诉它，我这里实际上是一个可以进行+=运算的数值型的玩意，
//                       - c++里面我们可以通过operator重载来实现任意类型的+=操作
//                       - go里面如何表示任意类型的+=操作呢？
//                         - 可以肯定的是不能，我们没办法在运算符+=上做文章，go也不支持这个
//                         - 但是我们可以借助go提供的constraint以及type set来实现这个，
//                           比如定义这里的约束为 interface{uint32|float32}，这样我们就可以将底层类型为uint32或者float32的类型作为我们的参数类型。
//                           但是对于其他的要支持+=操作的自定义类型，比如type xxx struct{}，那只能通过增加接口来解决，毕竟不是c++没法重载运算符+=
// 	l.mu.Unlock()
// }

func (l *Lockable[T]) DoAdd(f func(*T)) {
	l.mu.Lock()
	f(&l.data)
	l.mu.Unlock()
}

func Equal[T comparable](a, b T) bool {
	return a == b
}

func Test_generics(t *testing.T) {
	var l1 Lockable[uint32]
	l1.mu.Lock()
	l1.data++
	l1.mu.Unlock()

	l1.DoAdd(func(t *uint32) {
		(*t)++
	})

	var l2 Lockable[float32]
	l2.mu.Lock()
	l2.data += 1.0
	l2.mu.Unlock()

	l2.DoAdd(func(t *float32) {
		(*t) += 1.0
	})

	assert.True(t, Equal[int32](1, 1))
	assert.True(t, Equal(1, 1))
	assert.False(t, Equal(1, 2))

	assert.True(t, Equal[string]("hello", "hello"))
	assert.False(t, Equal("hello", "world"))

	//assert.False(t, Equal([]int{}, []int{})) // []int 没有实现限制comparable（不可比较）
}

// type set
type numberal interface {
	~uint | ~uint32 | ~uint64 | ~int | ~int32 | ~int64 | ~float32 | ~float64
}

type MyLockable[T numberal] struct {
	mu   sync.Mutex
	data T
}

func Test_mylockable(t *testing.T) {
	var l1 MyLockable[uint32]
	l1.mu.Lock()
	l1.data += 1
	l1.mu.Unlock()

	var l2 MyLockable[int32]
	l2.mu.Lock()
	l2.data += 1
	l2.mu.Unlock()

	var l3 MyLockable[float32]
	l3.mu.Lock()
	l3.data += 1.0
	l3.mu.Unlock()
}
