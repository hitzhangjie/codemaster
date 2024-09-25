package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/hitzhangjie/codemaster/gc/rss"
)

func main() {
	// 方法3
	ballast := make([]byte, 1<<30)

	// do something
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)
	fmt.Println("alloc:", humanReadableString(m.Sys))
	fmt.Println("freed", humanReadableString(m.HeapReleased))
	fmt.Println("rss", humanReadableString(rss.GetRSS()))
	time.Sleep(time.Minute)

	runtime.KeepAlive(&ballast)
}

func humanReadableString(numOfBytes uint64) string {
	var sizes = []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for numOfBytes >= 1024 && i < len(sizes)-1 {
		numOfBytes /= 1024
		i++
	}
	return fmt.Sprintf("%d%s", numOfBytes, sizes[i])
}
