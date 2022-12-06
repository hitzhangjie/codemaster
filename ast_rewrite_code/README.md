# README

This demo introduces how to use go AST to analyze and modify your code.

Here's the source code for testing:

```go
// Package main just outputs hello world
package main

// comment 1

// comment 2

// comment 3
// comment 4

import (
	"fmt"
)

func add(a, b int) int {
	return a + b
}

func main() {
	fmt.Println("hello world")
}
```

This demo:
- reads the source code, 
- and removes unused comments, 
- and add comments for function `main.add`,
- then it rewrites the AST as the source code,
- the new code will be written to main.go.txt.

You can run `go test` to test.

The expected new code should be:

```go
package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello world")
}
```