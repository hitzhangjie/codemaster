package main_test

import (
	"debug/elf"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseELF(t *testing.T) {
	f, err := elf.Open("debug.test")
	assert.Nil(t, err)

	for _, s := range f.Sections {
		fmt.Println(s)
	}

	fmt.Println(strings.Repeat("-", 78))

	syms, err := f.Symbols()
	assert.Nil(t, err)
	for _, s := range syms {
		//fmt.Println(s, f.Sections[s.Section])
		if int(s.Section) >= len(f.Sections) {
			fmt.Println("skip:", s.Info)
			continue
		}
		fmt.Println(s.Name, s.Section, f.Sections[s.Section])
	}
}
