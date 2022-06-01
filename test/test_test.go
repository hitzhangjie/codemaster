package test

import (
	"fmt"
	"strconv"
	"testing"
)

func TestXXX(t *testing.T) {
	// t.Errorf only marks the testcase failed
	t.Errorf("errorf")
	fmt.Println("helloworld")

	// t.Fatalf not only marks the testcase failed, but also runtime.Goexit,
	// so following statements following t.Fatalf won't be executed
	t.Fatalf("fatalf")
	fmt.Println("helloworld2")
}

//BenchmarkSprintfInt2Str-8   	13052944	        83.63 ns/op
func BenchmarkSprintfInt2Str(b *testing.B) {
	var n int = 1024
	var s string
	for i := 1; i < b.N; i++ {
		s = fmt.Sprintf("%d", n)
		_ = s
	}
}

//BenchmarkItoa2Str-8   	55205221	        19.01 ns/op
func BenchmarkItoa2Str(b *testing.B) {
	var n int = 1024
	var s string
	for i := 1; i < b.N; i++ {
		s = strconv.Itoa(n)
		_ = s
	}
}
