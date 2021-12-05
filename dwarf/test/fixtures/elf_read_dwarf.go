package main

import "fmt"

type Student struct {
	Name string
	Age  int
}

func main() {
	s := Student{}
	fmt.Println(s)
}
