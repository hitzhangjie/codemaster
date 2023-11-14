package main

import (
	"runtime"
	"time"
)

// 方法1
//var ballast = make([]byte, 1<<30)

// 方法2
// var ballast []byte
// func init() {
//     ballast = make([]byte, 1<<30)
// }

func main() {
	// 方法3
	ballast := make([]byte, 1<<30)

	// do something
	// ...
	time.Sleep(time.Minute)

	runtime.KeepAlive(&ballast)
}
