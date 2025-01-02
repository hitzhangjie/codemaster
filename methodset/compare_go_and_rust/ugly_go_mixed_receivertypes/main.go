package main

// Receiver of custom defined struct type.
type student struct {
	age int
}

func (s student) Age() int {
	return s.age
}

func (s student) SetAge(v int) {
	s.age = v
}

func main() {
	s1 := student{}
	s1.SetAge(100)
	n := s1.Age()
	println(n)

	s2 := &student{}
	s2.SetAge(200) // 尽管s2是指针，但是实际上编译器
	m := s2.Age()
	println(m)
}
