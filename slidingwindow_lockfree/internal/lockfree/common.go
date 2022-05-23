package lockfree

import (
	"sync/atomic"
	"unsafe"
)

type node struct {
	next unsafe.Pointer
	v    interface{}
}

func load(p *unsafe.Pointer) *node {
	return (*node)(atomic.LoadPointer(p))
}
func cas(p *unsafe.Pointer, old, new *node) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
