package switchcase_test

import (
	"fmt"
)

func Example_switchcase_fallthrough() {
	var a int = 1
	switch a {
	case 1: // case 1分支满足，执行case 1对应的代码块
		fmt.Println("a is 1")
		fallthrough
	case 2: // 前一个分支加了fallthrough，case 2分支条件不满足也会执行对应的代码块
		fmt.Println("a is 2")
	}
	// Output:
	// a is 1
	// a is 2
}
