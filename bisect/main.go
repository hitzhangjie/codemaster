// how to test this?
//
// 1. go build -o main
// 2. bisect ADD_V2_PATTERN=PATTERN ./main
package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"runtime"

	bisect "github.com/hitzhangjie/codemaster/bisect/internal"
)

var (
	/*
		$ bisect PRINT=PATTERN ADD_V2=PATTERN ./main

		bisect: checking target with all changes disabled
		bisect: run: PRINT=n ADD_V2=n ./main... ok (0 matches)
		bisect: run: PRINT=n ADD_V2=n ./main... ok (0 matches)
		bisect: checking target with all changes enabled
		bisect: run: PRINT=y ADD_V2=y ./main... FAIL (0 matches)
		bisect: fatal error: target failed without printing any matches
		Enable addv2
		bug detected, call stack:
		        main.Add at /home/zhangjie/hitzhangjie/codemaster/bisect/main.go:65
		        main.main at /home/zhangjie/hitzhangjie/codemaster/bisect/main.go:40
		        runtime.main at /usr/local/go/src/runtime/proc.go:283
		        runtime.goexit at /usr/local/go/src/runtime/asm_amd64.s:1700
		Shit ... overflow happened
	*/
	changelist1Matcher *bisect.Matcher

	/*
		// bisect运行原理是如果想测试某个pattern，那么这个pattern=n时target必须成功(即os.Exit(0))
		// 因为用例设计的原因，ADD_V2=y时总是会失败，所以我们必须显示禁用掉之后再测试

		$ bisect PRINT=PATTERN ADD_V2=n ./main

		```bash
		bisect: checking target with all changes disabled
		bisect: run: PRINT=n ADD_V2=n ./main... ok (0 matches)
		bisect: run: PRINT=n ADD_V2=n ./main... ok (0 matches)
		bisect: checking target with all changes enabled
		bisect: run: PRINT=y ADD_V2=n ./main... FAIL (0 matches)
		bisect: fatal error: target failed without printing any matches
		Enable addv1
		        Good ... overflow detected

		9223372036854775807 + 1 = -9223372036854775808
		Enable printHelloWorld
		printHelloWorld err: not implemented
		bug detected, call stack:
		        main.main.func1 at /home/zhangjie/hitzhangjie/codemaster/bisect/main.go:57
		        runtime.gopanic at /usr/local/go/src/runtime/panic.go:792
		        main.printHelloWorld at /home/zhangjie/hitzhangjie/codemaster/bisect/main.go:129
		        main.main at /home/zhangjie/hitzhangjie/codemaster/bisect/main.go:57
		        runtime.main at /usr/local/go/src/runtime/proc.go:283
		        runtime.goexit at /usr/local/go/src/runtime/asm_amd64.s:1700
		```
	*/
	changelist2Matcher *bisect.Matcher
	err                error
)

const changelist1 = "ADD_V2"
const changelist2 = "PRINT"

func init() {
	changelist1Matcher, err = bisect.New(os.Getenv(changelist1))
	if err != nil {
		log.Fatalf("failed to create matcher: %v", err)
	}
	changelist2Matcher, err = bisect.New(os.Getenv(changelist2))
	if err != nil {
		log.Fatalf("failed to create matcher: %v", err)
	}
}

func main() {
	var a int64 = math.MaxInt64
	var b int64 = 1
	var c = Add(a, b)

	fmt.Printf("%d + %d = %d\n", a, b, c)

	if changelist2Matcher.ShouldEnable(bisect.Hash(changelist2)) {
		fmt.Println("Enable printHelloWorld")
		defer func() {
			if e := recover(); e != nil {
				if changelist2Matcher.ShouldReport(bisect.Hash(changelist2)) {
					fmt.Println("printHelloWorld err:", e)
					fmt.Printf("bug detected, call stack:\n")
					stack := callstack(+6)
					printstack(stack)
				}
				os.Exit(1)
			}
		}()
		printHelloWorld()
	}
}

func Add(a, b int64) (sum int64) {
	// change: 这里根据
	if changelist1Matcher.ShouldEnable(bisect.Hash(changelist1)) {
		fmt.Println("Enable addv2")
		v := addv2(a, b)
		if v < a || v < b {
			if changelist1Matcher.ShouldReport(bisect.Hash(changelist1)) {
				fmt.Printf("bug detected, call stack:\n")
				stk := callstack(-4)
				printstack(stk)
			}
			fmt.Println("Shit ... overflow happened")
			os.Exit(1)
		} else {
			sum = v
		}
	} else {
		fmt.Println("Enable addv1")
		sum = addv1(a, b)
	}
	return sum
}

func addv1(a, b int64) int64 {
	// Supposing we solved the overflow here by returning error,
	// or logging error, or report metrics, etc.
	if math.MaxInt64-a < b {
		fmt.Println("\tGood ... overflow detected\n")
	}
	return a + b
}

// BUG: someone don't understand how overflow happens, he want to
// simpilify `addv1`, so here addv2 comes :)
func addv2(a, b int64) int64 {
	return a + b
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

func printHelloWorld() {
	panic("not implemented")
}
