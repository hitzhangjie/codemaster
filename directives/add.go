package directives

import (
	_ "unsafe"
)

//go:linkname Add github.com/hitzhangjie/codemaster/directives/add.add
func Add(a, b int) int
