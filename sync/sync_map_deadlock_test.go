package sync_test

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// sync.Map赋值重做导致的死锁问题:
// ```go
// var cacheOld sync.Map
// var cacheNew sync.Map
//
// cacheNew.Store(1,1) // init sync.Map.readOnly
// cacheNew.Store(2,2) // init sync.Map.dirty
// cacheNew.Range(fn)  // multiple times, may lock/unlock sync.Map.mu
//
//	// and prompt sync.Map.dirty to sync.Map.readOnly
//
// cacheOld = cacheNew // if cacheNew.mu is locked, then copy a locked mutex to cacheOld
// cacheOld.Range(fn)  // if cacheOld.readOnly.amended && cacheOld.mu locked,
//
//	                    // then cacheOld.Range(fn) or cacheOld.Store/Load may try to lock cacheOld.mu again,
//	                    // ...
//						   // then deadlock!!!
//
// ```
//
// 因为这把锁被从cache拷贝到了cacheOld，cacheOld上仅剩range
// 读操作，会涉及到加锁操作，但是假如在copy期间要拷贝的这个map里的锁
// 状态已经是locked状态，就会导致cacheOld上的range操作死锁。
//
// --------------------------------------------------------------------------
//
// @@@panic1@@@:
//
//   - fatal error: sync: unlock of unlocked mutex
//
//   - panic: assignment to entry in nil map
//
//   - fatal error: concurrent map writes
//
//   - fatal error: concurrent map read and map write
//
//     ps: deadlock虽然没有复现出来，但是肯定是存在这种可能性的！！！！
func Test_SyncMap_CopyLockedMutex_Causes_Deadlock(t *testing.T) {
	var cacheOld sync.Map
	var cacheNew sync.Map

	go func() {
		for i := 0; ; i++ {
			func() {
				if e := recover(); e != nil {
					fmt.Println(e)
				}
				cacheNew.Store(i%1000, i)
				cacheNew.Store(i%1000+1, i)
				cacheNew.Store(i%1000+2, i)
				cacheNew.Store(i%1000+3, i)
				cacheNew.Store(i%1000+4, i)
			}()
		}
	}()

	go func() {
		for i := 0; ; i++ {
			func() {
				if e := recover(); e != nil {
					fmt.Println(e)
				}
				cacheOld = cacheNew
			}()
		}
	}()

	go func() {
		for i := 0; ; i++ {
			func() {
				if e := recover(); e != nil {
					fmt.Println(e)
				}
				cacheOld.Store(i%1000, i)
			}()
		}
	}()

	time.Sleep(10 * time.Second)
}
