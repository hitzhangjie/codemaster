package main

import "fmt"

type Student struct {
	Name string
	Age  int
}

type Print func(s string, vals ...interface{})

func main() {
	s := Student{}
	fmt.Println(s)
}
