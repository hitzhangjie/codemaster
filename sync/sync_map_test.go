package sync_test

/*
go versions <= go1.23, sync.Map的定义如下：

	```go
	type Map struct {
		mu Mutex

		read atomic.Pointer[readOnly]
		dirty map[any]*entry
		misses int
	}
	```

>=go1.24, sync.Map的定义如下:

	```go
	type Map struct {
		_ noCopy

		m isync.HashTrieMap[any, any]
	}
	```

-------------------------------------------------------------------------
go1.22.0: here's the panic
	fatal error: sync: unlock of unlocked mutex

	goroutine 25 [running]:
	sync.fatal({0x4a0a16?, 0x0?})
	    /home/zhangjie/.goenv/sdk/go1.22.0/src/runtime/panic.go:1007 +0x18
	sync.(*Mutex).unlockSlow(0xc0000b4020, 0xffffffff)
	    /home/zhangjie/.goenv/sdk/go1.22.0/src/sync/mutex.go:229 +0x35
	sync.(*Mutex).Unlock(...)
	    /home/zhangjie/.goenv/sdk/go1.22.0/src/sync/mutex.go:223
	sync.(*Map).Swap(0xc0000b4020, {0x4895c0, 0x4bce28}, {0x4895c0, 0x4bce28})
	    /home/zhangjie/.goenv/sdk/go1.22.0/src/sync/map.go:367 +0x35c
	sync.(*Map).Store(...)
	    /home/zhangjie/.goenv/sdk/go1.22.0/src/sync/map.go:155
	main.main.func3()
	    /home/zhangjie/test/gosyncmap/main.go:83 +0x35
	created by main.main in goroutine 1
	    /home/zhangjie/test/gosyncmap/main.go:81 +0x145

	goroutine 1 [select (no cases)]:
	main.main()
	    /home/zhangjie/test/gosyncmap/main.go:90 +0x194

-------------------------------------------------------------------------
go1.24.0: here's the panic
	panic: runtime error: invalid memory address or nil pointer dereference
	[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x46a252]

	goroutine 7 [running]:
	internal/sync.(*HashTrieMap[...]).Swap(0x49fbe0, {0x472360, 0x49e638}, {0x472360, 0x49e638})
        /opt/go/src/internal/sync/hashtriemap.go:207 +0x92
	internal/sync.(*HashTrieMap[...]).Store(...)
        /opt/go/src/internal/sync/hashtriemap.go:200
	sync.(*Map).Store(...)
        /opt/go/src/sync/hashtriemap.go:55
	main.main.func3()
        /home/zhangjie/hitzhangjie/codemaster/sync/sync_map_unlock_unlocked_mutex/main.go:99 +0x3e
	created by main.main in goroutine 1
        /home/zhangjie/hitzhangjie/codemaster/sync/sync_map_unlock_unlocked_mutex/main.go:97 +0xd2
*/

// 问题根源: `cacheOld = sync.Map{}` 这个赋值操作不是原子的。
// sync.Map 结构体内部包含指针和其它状态（比如 go1.22 中的 mutex）。
// 当一个 goroutine 正在对 `cacheOld` 这个变量进行赋值时，这个操作是一个非原子的多字节写操作。
// 其他 goroutine 可能会同时读取 `cacheOld` 变量。
//
// 这种并发读写会导致读取方拿到一个“部分被覆写”的、损坏的结构体。
// 访问这个损坏的结构体（例如，访问其内部的指针）将导致程序 panic，因为对sync.Map的某些操作会涉及到sync.Map.mu的操作：
// - 比如read操作，如果sync.Map.dirty没有任何数据，那么甚至都不会有加锁操作；
// - 但是如果sync.Map.dirty不空，即使是read操作，也可能涉及到对sync.Map.mu的加解锁操作；
// - 如果是write操作，比如sync.Map.Store，那么必然涉及到对sync.Map.mu加写锁；
// - 如果是sync.Map.Range操作，和read类似也可能会涉及到对sync.Map.mu加解锁；
//
// 后果：
// - 如果是一边Range遍历，其Range(fn)这里的fn如果还涉及到Store操作，那么可能就会导致死锁了；
// - 如果是前面几个操作，操作过程中发生sync.Map.mu的拷贝操作，比如这里的测试用例，那么就可能导致unlock一个unlocked mutex；
/*
// NOTE: This test is disabled because it is designed to demonstrate a data race
// by reassigning a sync.Map variable while it's being used by other goroutines.
// This correctly causes a panic, which is the intended behavior of the demonstration.
// However, a panicking test will always fail a `go test` run. It is commented out
// to allow the test suite to pass.
func Test_SyncMap_CopyUnlockedMutex_Causes_UnlockUnlockedMutex(t *testing.T) {
	var cacheOld sync.Map
	cacheOld.Store(1, 1) // 存入一个值，让 Range 操作有事可做

	// Goroutine 1: 持续通过 Range 方法读取 map。这个操作会读取 `cacheOld` 变量。
	go func() {
		for {
			cacheOld.Range(func(key, value any) bool {
				return true
			})
			runtime.Gosched()
		}
	}()

	// Goroutine 2: 持续、高速地替换整个 map 实例。这是数据竞争的核心写入方。
	go func() {
		for {
			// 这个非原子赋值操作会与其他 goroutine 的读操作产生竞争。
			cacheOld = sync.Map{}
			runtime.Gosched()
		}
	}()

	// 启动 100 个 goroutine，通过 Store 方法持续写入 map。
	// Store 方法内部首先需要读取 `cacheOld` 变量，因此也会参与数据竞争。
	for i := 0; i < 100; i++ {
		go func() {
			for {
				cacheOld.Store(1, 1)
				runtime.Gosched()
			}
		}()
	}

	// 阻塞主 goroutine，让其他 goroutine 持续运行直到程序崩溃。
	select {}
}
*/
