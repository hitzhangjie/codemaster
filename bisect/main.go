// how to test this?
//
// 1. go build -o main
// 2. bisect MyPattern=PATTERN ./main
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	bisect "github.com/hitzhangjie/codemaster/bisect/internal"
)

var (
	matcher *bisect.Matcher
	err     error
)

const pattern = "MyPattern"

func init() {
	flag.Parse()

	matcher, err = bisect.New(os.Getenv(pattern))
	if err != nil {
		log.Fatalf("failed to create matcher: %v", err)
	}
}

func main() {
	var a = 1
	var b = 2
	var sum int

	if matcher.ShouldEnable(bisect.Hash(pattern)) {
		sumv2 := addv2(a, b)
		if sumv2 != sum {
			if matcher.ShouldReport(bisect.Hash(pattern)) {
				fmt.Println("bug detected in v2: addv2 != addv1")
			}
			os.Exit(1)
		} else {
			sum = sumv2
		}
	} else {
		sum = addv1(a, b)
		fmt.Println("Using addv1")
	}

	fmt.Printf("%d + %d = %d\n", a, b, sum)
}

func addv1(a, b int) int {
	return a + b
}

func addv2(a, b int) int {
	return a - b
}
