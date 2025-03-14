/*
	how to run this test?


	```bash
	$ bisect FEAT1=PATTERN FEAT2=PATTERN FEAT3=PATTERN ./main

	bisect FEAT1=PATTERN FEAT2=PATTERN FEAT3=PATTERN ./main
	bisect: checking target with all changes disabled
	bisect: run: FEAT1=n FEAT2=n FEAT3=n ./main... ok (0 matches)
	bisect: run: FEAT1=n FEAT2=n FEAT3=n ./main... ok (0 matches)
	bisect: checking target with all changes enabled
	bisect: run: FEAT1=y FEAT2=y FEAT3=y ./main... FAIL (3 matches)
	bisect: run: FEAT1=y FEAT2=y FEAT3=y ./main... FAIL (3 matches)
	bisect: target succeeds with no changes, fails with all changes
	bisect: searching for minimal set of enabled changes causing failure
	bisect: run: FEAT1=+0 FEAT2=+0 FEAT3=+0 ./main... ok (0 matches)
	bisect: run: FEAT1=+0 FEAT2=+0 FEAT3=+0 ./main... ok (0 matches)
	bisect: run: FEAT1=+1 FEAT2=+1 FEAT3=+1 ./main... ok (0 matches)
	bisect: run: FEAT1=+1 FEAT2=+1 FEAT3=+1 ./main... ok (0 matches)
	bisect: run: FEAT1=+0+1 FEAT2=+0+1 FEAT3=+0+1 ./main... FAIL (2 matches)
	bisect: run: FEAT1=+0+1 FEAT2=+0+1 FEAT3=+0+1 ./main... FAIL (2 matches)
	bisect: run: FEAT1=+00+1 FEAT2=+00+1 FEAT3=+00+1 ./main... FAIL (1 matches)
	bisect: run: FEAT1=+00+1 FEAT2=+00+1 FEAT3=+00+1 ./main... FAIL (1 matches)
	bisect: run: FEAT1=+1+x70 FEAT2=+1+x70 FEAT3=+1+x70 ./main... FAIL (1 matches)
	bisect: run: FEAT1=+1+x70 FEAT2=+1+x70 FEAT3=+1+x70 ./main... FAIL (1 matches)
	bisect: confirming failing change set
	bisect: run: FEAT1=v+x70+x89 FEAT2=v+x70+x89 FEAT3=v+x70+x89 ./main... FAIL (2 matches)
	bisect: run: FEAT1=v+x70+x89 FEAT2=v+x70+x89 FEAT3=v+x70+x89 ./main... FAIL (2 matches)
	bisect: FOUND failing change set
	--- change set #1 (enabling changes causes failure)
	main.go:56 feat2() called
	main.go:63 feat1() called
	---
	bisect: checking for more failures
	bisect: run: FEAT1=-x70-x89 FEAT2=-x70-x89 FEAT3=-x70-x89 ./main... ok (0 matches)
	bisect: run: FEAT1=-x70-x89 FEAT2=-x70-x89 FEAT3=-x70-x89 ./main... ok (0 matches)
	bisect: target succeeds with all remaining changes enabled
	```
*/

package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	bisect "github.com/hitzhangjie/codemaster/bisectv2/internal"
)

var (
	changelist1Matcher *bisect.Matcher // 1 and 2 enabled, will causes fail
	changelist2Matcher *bisect.Matcher // 1 and 2 enabled, will causes fail
	changelist3Matcher *bisect.Matcher // always success

	err error
)

const changelist1 = "FEAT1"
const changelist2 = "FEAT2"
const changelist3 = "FEAT3"

var (
	h1 = bisect.Hash(changelist1)
	h2 = bisect.Hash(changelist2)
	h3 = bisect.Hash(changelist3)
)

func init() {
	changelist1Matcher, err = bisect.New(os.Getenv(changelist1))
	if err != nil {
		log.Fatalf("failed to create matcher: %v", err)
	}
	changelist2Matcher, err = bisect.New(os.Getenv(changelist2))
	if err != nil {
		log.Fatalf("failed to create matcher: %v", err)
	}
	changelist3Matcher, err = bisect.New(os.Getenv(changelist3))
	if err != nil {
		log.Fatalf("failed to create matcher: %v", err)
	}
}

func main() {
	ch := make(chan int)
	go func() {
		// lock, missing unlock
		if changelist2Matcher.ShouldEnable(h2) {
			feat2()
		} else {
			fmt.Println("disable feat2()")
		}

		// lock, and unlock
		if changelist1Matcher.ShouldEnable(h1) {
			feat1()
		} else {
			fmt.Println("disable feat1()")
		}

		// no lock operation
		if changelist3Matcher.ShouldEnable(h3) {
			feat3()
		} else {
			fmt.Println("disable feat3()")
		}
		ch <- 1
	}()

	select {
	case <-ch:
		os.Exit(0)
	case <-time.After(time.Second):
		if changelist2Matcher.ShouldReport(h2) {
			fmt.Printf("main.go:56 %s feat2() called\n", bisect.Marker(h2))
		}
		if changelist1Matcher.ShouldReport(h1) {
			fmt.Printf("main.go:63 %s feat1() called\n", bisect.Marker(h1))
		}
		if changelist3Matcher.ShouldReport(h3) {
			fmt.Printf("main.go:70 %s feat3() called\n", bisect.Marker(h3))
		}
		os.Exit(1)
	}
}

var (
	mu  sync.Mutex
	val int64
)

// feat2 doesn't release, it will cause following feat1() feat2() call deadlock
func feat1() error {
	mu.Lock()
	val++
	mu.Unlock()
	return nil
}

// feat2 doesn't release, it will cause following feat1() feat2() call deadlock
func feat2() error {
	mu.Lock()
	val--
	//mu.Unlock()
	return nil
}

func feat3() error {
	return nil
}

func callstack(offset int) []string {
	callerspc := make([]uintptr, 8)
	n := runtime.Callers(2, callerspc)
	callerspc = callerspc[:n]
	frames := runtime.CallersFrames(callerspc)
	var stack []string

	calcOffset := true
	for {
		f, more := frames.Next()
		if calcOffset {
			stack = append(stack, fmt.Sprintf("%s at %s:%d", f.Function, f.File, f.Line+offset))
			calcOffset = false
		} else {
			stack = append(stack, fmt.Sprintf("%s at %s:%d", f.Function, f.File, f.Line))
		}
		if !more {
			break
		}
	}
	return stack
}

func printstack(stack []string) {
	for i := 0; i < len(stack); i++ {
		fmt.Printf("\t%s\n", stack[i])
	}
}
