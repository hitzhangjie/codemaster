package golang

import (
	"fmt"
	"sync"
	"testing"
)

func Test_syncmap(t *testing.T) {
	m := sync.Map{}
	v, ok := m.LoadOrStore("hello", 1)
	fmt.Println(v)
	fmt.Println(ok)
}
