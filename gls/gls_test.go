package tvars

import (
	"bytes"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"

	"github.com/v2pro/plz/gls"
	"go.uber.org/atomic"
)

///////////////////////////////////////////////////////////////////////////////
// - 普通整数++，性能最高，但是不适用于并发环境，必须加写保护
//   - mutex，写性能差，12.5ns
//   - atomic，写性能好与mutex，无并发写场景下7.21ns/op，有并发写冲突情况下约20.4ns/op，且和并发协程数没多少关系

// BenchmarkIntegerAdd
// BenchmarkIntegerAdd-8   	1000000000	         0.313 ns/op
func BenchmarkIntegerAdd(b *testing.B) {
	var a int64
	for i := 0; i < b.N; i++ {
		a++
	}
}

// BenchmarkIntegerAddWithMutex
// BenchmarkIntegerAddWithMutex-8   	96309058	        12.5 ns/op
func BenchmarkIntegerAddWithMutex(b *testing.B) {
	var a int64
	var m sync.Mutex
	for i := 0; i < b.N; i++ {
		m.Lock()
		a++
		m.Unlock()
	}
}

// BenchmarkIntegerAtomic
// BenchmarkIntegerAtomic-8   	166263684	         7.21 ns/op
func BenchmarkIntegerAtomic(b *testing.B) {
	var a atomic.Int64
	for i := 0; i < b.N; i++ {
		a.Inc()
	}
}

// BenchmarkIntegerAtomic_2G
// BenchmarkIntegerAtomic_2G-8   	58196326	        20.4 ns/op
func BenchmarkIntegerAtomic_2G(b *testing.B) {
	var a atomic.Int64
	b.SetParallelism(2)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.Inc()
		}
	})
}

// BenchmarkIntegerAtomic_4G
// BenchmarkIntegerAtomic_4G-8   	59256405	        20.4 ns/op
func BenchmarkIntegerAtomic_4G(b *testing.B) {
	var a atomic.Int64
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.Inc()
		}
	})
}

// BenchmarkIntegerAtomic_8G
// BenchmarkIntegerAtomic_8G-8   	62279982	        20.6 ns/op
func BenchmarkIntegerAtomic_8G(b *testing.B) {
	var a atomic.Int64
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.Inc()
		}
	})
}

// BenchmarkIntegerAtomic_64
// BenchmarkIntegerAtomic_64-8   	71354692	        16.9 ns/op
func BenchmarkIntegerAtomic_64(b *testing.B) {
	var a atomic.Int64
	b.SetParallelism(64)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.Inc()
		}
	})
}

// BenchmarkIntegerAtomic_128
// BenchmarkIntegerAtomic_128-8   	58768611	        20.1 ns/op
func BenchmarkIntegerAtomic_128(b *testing.B) {
	var a atomic.Int64
	b.SetParallelism(128)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.Inc()
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// 获取goid性能对比
// - 通过汇编方式获取tls中的g，再获取goid是性能比较高的，3704ns左右
// - 通过runtime stack解析，配合对象池优化，能优化到3873ns左右
//
// 如果要用的话，建议采用第一种方案

// BenchmarkGoID 通过线程局部存储获取到g地址，然后取g中的goid
//
// 当前这个库的实现方式参考自golang源码, see:
// https://sourcegraph.com/github.com/golang/go@9ece63f0647ec34cc729ad71a87254193014dcca/-/blob/src/runtime/stubs.go#L18
//
// BenchmarkGoID-8   	  426937	      3704 ns/op
func BenchmarkGoID(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			go func() {
				b.Logf("G load LoadGoIDByRuntimeStack: %d", gls.GoID())
			}()
		}
	})
}

var pool = sync.Pool{
	New: func() interface{} { return make([]byte, 16) },
}

var delimi = []byte(" ")

// LoadGoIDByRuntimeStack 通过运行时栈信息，解析出goid
//
// runtime stack info example:
//
// goroutine 18 [running]:
//   awesomeProject/goid.TestGetStackInfo(0xc000082600)
//     /Users/liangyaopei/go/src/awesomeProject/goid/goid_test.go:18 +0x6f
//   testing.tRunner(0xc000082600, 0x114d5c0)
//     /usr/local/go/src/testing/testing.go:1127 +0xef
//   created by testing.(*T).Run
//     /usr/local/go/src/testing/testing.go:1178 +0x386
func LoadGoIDByRuntimeStack() int {
	b := pool.Get()
	buf := b.([]byte)
	n := runtime.Stack(buf, false)
	id := bytes.Split(buf[:n], delimi)

	//fmt.Println(string(buf))

	v := 0
	for _, b := range id[1] {
		v = v*10 + int(b) - 48
	}
	return v
}

func Test_goid(t *testing.T) {
	ch := make(chan int)
	go func() {
		t.Logf("G load LoadGoIDByRuntimeStack: %d", LoadGoIDByRuntimeStack())
		ch <- 1
	}()
	<-ch
}

// BenchmarkGoID_ByRuntimeStack 性能和通过汇编方式获取，多了169ns
//
// BenchmarkGoID_ByRuntimeStack-8   	  277622	      3873 ns/op
func BenchmarkGoID_ByRuntimeStack(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			go func() {
				b.Logf("G load LoadGoIDByRuntimeStack: %d", LoadGoIDByRuntimeStack())
			}()
		}
	})
}

//////////////////////////////////////////////////////////////////////////////////
// 对比10000个协程并发写的性能对比
// - 10000g，写atomic/读atomic，平均耗时
// - 10000g，写gls中的counter/合并读，平均耗时 --> 注意这里通过map<key,map<goid,value>>模拟gls，并非真正的gls
// - 10000g，写gls中的counter/合并读，平均耗时 --> 注意这里通过map<key,map<tid,value>>模拟gls，近似gls

//deprecated 难以实现，go官方也不推荐这种方式，纯粹是为了实验、对比性能
func LoadThreadID() int {
	return -1
}

// BenchmarkWrite_10000G
// BenchmarkWrite_10000G-8   	 5587188	       225 ns/op
func BenchmarkWrite_10000G(b *testing.B) {
	vals := map[string]*atomic.Int32{}
	mux := &sync.Mutex{}

	keys := []string{"a", "b", "c", "d", "e"}

	b.SetParallelism(10000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {

			k := rand.Int() % 5
			s := keys[k]

			mux.Lock()
			m, ok := vals[s]
			mux.Unlock()

			if !ok {
				m = atomic.NewInt32(0)
				mux.Lock()
				vals[s] = m
				mux.Unlock()
			}
			m.Inc()
		}
	})
}

// BenchmarkGLSWrite_10000G 自己实现了写操作，通过map<key,map<goid,value>>来减轻并发写压力
//
// BenchmarkGLSWrite_10000G-8   	 3478429	       348 ns/op map[key]gvalues{mutex, map[goid]value}
// BenchmarkGLSWrite_10000G-8   	 2792016	       417 ns/op map[key]gvalues{sync.Map}
//
// 对map、sync.Map、concurrent_map进行了benchmark，结果如下：
//
// BenchmarkDeleteEmptyMap-8               20000000                86.9 ns/op
// BenchmarkDeleteEmptySyncMap-8           300000000                5.16 ns/op
// BenchmarkDeleteEmptyCMap-8              50000000                34.8 ns/op
//
// BenchmarkDeleteMap-8                    10000000               131 ns/op
// BenchmarkDeleteSyncMap-8                10000000               135 ns/op
// BenchmarkDeleteCMap-8                   30000000                37.0 ns/op
//
// BenchmarkLoadEmptyMap-8                 20000000                87.9 ns/op
// BenchmarkLoadEmptySyncMap-8             300000000                5.03 ns/op
// BenchmarkLoadEmptyCMap-8                100000000               17.1 ns/op
//
// BenchmarkLoadMap-8                      20000000               111 ns/op
// BenchmarkLoadSyncMap-8                  100000000               12.8 ns/op
// BenchmarkLoadCMap-8                     100000000               22.5 ns/op
//
// BenchmarkSetMap-8                       10000000               187 ns/op
// BenchmarkSetSyncMap-8                    5000000               396 ns/op
// BenchmarkSetCMap-8                      20000000                84.9 ns/op
func BenchmarkGLSWrite_10000G(b *testing.B) {
	// key,
	// val=map[goid]value
	type gvalues struct {
		vals *sync.Map
	}

	vals := map[string]*gvalues{}
	mux := &sync.Mutex{}

	keys := []string{"a", "b", "c", "d", "e"}
	b.SetParallelism(10000)
	b.RunParallel(func(pb *testing.PB) {
		goid := int(gls.GoID())
		for pb.Next() {
			idx := rand.Int() % 5
			k := keys[idx]

			mux.Lock()
			v, ok := vals[k]
			if !ok {
				v = &gvalues{vals: &sync.Map{}}
			}
			vals[k] = v
			mux.Unlock()

			gv, ok := v.vals.Load(goid)
			if !ok {
				v.vals.Store(goid, 1)
				continue
			}
			v.vals.Store(goid, gv.(int)+1)
		}
	})

	for k, m := range vals {
		m.vals.Range(func(id, vv interface{}) bool {
			fmt.Printf("key:%s, goid:%d, val:%d\n", k, id.(int), vv.(int))
			return true
		})
	}
}
