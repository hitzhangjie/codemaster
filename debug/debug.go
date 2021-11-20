package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hitzhangjie/codemaster/debug/internal/objfile"
)

func main() {
	exe := os.Args[1]
	fp, err := filepath.Abs(exe)
	checkErr(err)

	f, err := objfile.Open(fp)
	checkErr(err)

	syms, err := f.Symbols()
	checkErr(err)

	for _, s := range syms {
		//fmt.Printf("%x\t%c\t%s\t%s\t%v\t%v\n", s.Addr, s.Code, s.Name, s.Type, s.Size, s.Relocs)
		fmt.Printf("%x\t%c\t%s\t%s\t%v\n", s.Addr, s.Code, s.Name, s.Type, s.Size)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
