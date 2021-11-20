package main

import (
	"debug/plan9obj"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	exe := os.Args[1]

	fp, err := filepath.Abs(exe)
	panicIfError(err)

	fin, err := os.Open(fp)
	panicIfError(err)

	//ef, err := elf.NewFile(fin)
	ef, err := plan9obj.NewFile(fin)
	panicIfError(err)

	fmt.Println(strings.Repeat("-", 78))
	reldata := ef.Section(".rel.data")
	panicIfError(err)
	fmt.Println(reldata)

	fmt.Println(strings.Repeat("-", 78))
	reltxt := ef.Section(".rel.txt")
	panicIfError(err)
	fmt.Println(reltxt)
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
