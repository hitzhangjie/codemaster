package test

import (
	"fmt"
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
