package methodset_test

import "testing"

type student struct {
	age  int
	name string
}

func (s student) SetAge(v int) {
	s.age = v
}

func TestHelloWorld(t *testing.T) {
	s := student{}
	s.SetAge(100)
	println(s.age)

	s2 := &student{}
	s2.SetAge(200) // (*s2).SetAge(200) or (*student).SetAge(*s2, 200)
	println(s2.age)
}
