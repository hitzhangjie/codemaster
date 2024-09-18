package gopanic_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/bytedance/sonic"
)

func Benchmark_faststring(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			s := fmt.Sprintf("helloworld%d", rand.Int())
			buf := fastStringToBytes(s)
			_ = buf
		}
	})
}

func Test_faststring(t *testing.T) {
	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s := fmt.Sprintf("helloworld%d", rand.Int())

				buf := fastStringToBytes(s)

				mm := map[string]any{
					"data": buf,
				}

				_, err := sonic.Marshal(mm)
				if err != nil {
					panic(err)
				}
			}
		}()
	}

	wg.Wait()
}

// using uintptr and unsafe.Pointer break the rules
//
// see: https://cs.opensource.google/go/go/+/refs/tags/go1.23.1:src/unsafe/unsafe.go
func fastStringToBytes(s string) []byte {
	strHeader := (*[2]uintptr)(unsafe.Pointer(&s))
	byteSliceHeader := [3]uintptr{
		strHeader[0],
		strHeader[1],
		strHeader[1],
	}
	verbose := rand.Int()%100 == 0
	if verbose {
		runtime.GC()
		time.Sleep(time.Second)
		runtime.GC()
	}
	return *(*[]byte)(unsafe.Pointer(&byteSliceHeader))
}

// this is safe
//
// see: https://groups.google.com/g/golang-nuts/c/Zsfk-VMd_fU/m/O1ru4fO-BgAJ
// see: https://cs.opensource.google/go/go/+/refs/tags/go1.23.1:src/unsafe/unsafe.go
func StringToBytes(s string) []byte {
	const max = 0x7fff0000
	if len(s) > max {
		panic("string too long")
	}
	return (*[max]byte)(unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&s)).Data))[:len(s):len(s)]
}
