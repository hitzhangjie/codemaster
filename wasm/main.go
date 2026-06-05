package main

import (
	"fmt"
	"syscall/js"
)

func main() {
	fmt.Println("Go WASM module loaded")

	js.Global().Set("greet", js.FuncOf(greet))
	js.Global().Set("add", js.FuncOf(add))

	<-make(chan struct{})
}

func greet(this js.Value, args []js.Value) any {
	name := "World"
	if len(args) > 0 {
		name = args[0].String()
	}
	return fmt.Sprintf("Hello, %s from Go WASM!", name)
}

func add(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return "need two numbers"
	}
	a := args[0].Int()
	b := args[1].Int()
	return a + b
}
