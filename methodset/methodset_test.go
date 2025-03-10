package methodset_test

import "testing"

type student struct {
	age  int
	name string
}

// 我们写这个函数，实际上编译器干了什么呢？
func (s student) SetAge(v int) {
	// 检测到ineffective assign，有助于发现这里的赋值可能存在问题，但是并不总是奏效，比如我加一个 _ = s.age后，就不会再报错。
	// 所以 …… 这里的写法 s.age = v 并没有靠谱的办法被检查出来，这里修改的s的拷贝的字段，而非s的字段
	//
	// 对于go语言设计者来说，一个小struct的拷贝比一个指针容易减轻GC开销，如果struct比较小，且没有对其进行修改的诉求，建议用value接收器类型，
	// 反之用pointer接收器类型。
	s.age = v
	// _ = s.age
}

func TestHelloWorld(t *testing.T) {
	s := student{}
	s.SetAge(100)
	println(s.age)

	s2 := &student{}
	s2.SetAge(200) // (*s2).SetAge(200) or (*student).SetAge(*s2, 200)
	println(s2.age)
}
