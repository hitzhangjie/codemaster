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
	"github.com/hitzhangjie/codemaster/gopanic"
	"google.golang.org/protobuf/proto"
)

// This file tries to investigate if this `badFastStringToBytes` results in the following core:
//
// ```
//    SIGSEGV: segmentation violation
//    92 SIGSEGV: segmentation violation                                                                                                                                                                                                                                            93 PC=0x41b077 m=4 sigcode=1 addr=0x20
//    94
//    95 goroutine 0 gp=0xc0000076c0 m=4 mp=0xc0000e5808 [idle]:
//    96 runtime.(*mspan).typePointersOfUnchecked(0xc009115c80?, 0xc009115c80?)
//    97     /data/home/user00/go/src/runtime/mbitmap_allocheaders.go:202 +0x37 fp=0x7faf1340dd48 sp=0x7faf1340dd28 pc=0x41b077
//    98 runtime.scanobject(0xc000080168?, 0xc000080168)
//    99     /data/home/user00/go/src/runtime/mgcmark.go:1441 +0x1ce fp=0x7faf1340ddd8 sp=0x7faf1340dd48 pc=0x42706e
//   100 runtime.gcDrain(0xc000080168, 0x3)
//   101     /data/home/user00/go/src/runtime/mgcmark.go:1242 +0x1f4 fp=0x7faf1340de40 sp=0x7faf1340ddd8 pc=0x4268b4
//   102 runtime.gcDrainMarkWorkerDedicated(...)
//   103     /data/home/user00/go/src/runtime/mgcmark.go:1124
//   104 runtime.gcBgMarkWorker.func2()
//   105     /data/home/user00/go/src/runtime/mgc.go:1387 +0xa5 fp=0x7faf1340de90 sp=0x7faf1340de40 pc=0x422f25
//   106 runtime.systemstack(0x800000)
//   107     /data/home/user00/go/src/runtime/asm_amd64.s:509 +0x4a fp=0x7faf1340dea0 sp=0x7faf1340de90 pc=0x47880a
// ```

// These testcases try to reproduce the bug:
// when we misuse uintptr and unsafe.Pointer, i.e, convert uintptr back to unsafe.Pointer in different statements,
// which is not guaranteed by the 8 rules of uintpr and unsafe.Pointer. This will result in a problem that reference
// to a garbage-collected object, means the reference is pointing to an unmmaped memory area.

// this testcase will trigger following panic:
// -------------------------------------------
// runtime: pointer 0xc000844e60 to unallocated span span.base()=0xc000844000 span.limit=0xc000846000 span.state=0
// fatal error: found bad pointer in Go heap (incorrect use of unsafe or cgo?)
func Test_faststring_sonic_marshal(t *testing.T) {
	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s := fmt.Sprintf("helloworld%d", rand.Int())

				buf := badFastStringToBytes(s)

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

//go:generate protoc --go_out=paths=source_relative:. test.proto

// this testcase will trigger following panic:
// -------------------------------------------
// runtime: pointer 0xc000171d40 to unallocated span span.base()=0xc000170000 span.limit=0xc000172000 span.state=0
// fatal error: found bad pointer in Go heap (incorrect use of unsafe or cgo?)
func Test_faststring_protobuf_marshal(t *testing.T) {
	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s := fmt.Sprintf("helloworld%d", rand.Int())

				buf := badFastStringToBytes(s)

				mm := &gopanic.TestMessage{
					F: buf,
				}

				_, err := proto.Marshal(mm)
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
func badFastStringToBytes(s string) []byte {
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

// this is safe, but typecasting is more efficient since go1.22
//
// see: https://groups.google.com/g/golang-nuts/c/Zsfk-VMd_fU/m/O1ru4fO-BgAJ
// see: https://cs.opensource.google/go/go/+/refs/tags/go1.23.1:src/unsafe/unsafe.go
func goodFastStringToBytes(s string) []byte {
	const max = 0x7fff0000
	if len(s) > max {
		panic("string too long")
	}
	return (*[max]byte)(unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&s)).Data))[:len(s):len(s)]
}
