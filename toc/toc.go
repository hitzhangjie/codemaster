package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var dir string

func init() {
	flag.StringVar(&dir, "d", ".", "specify the documents directory")
	flag.Parsed()
}

func main() {
	if !filepath.IsAbs(dir) {
		d, err := filepath.Abs(dir)
		if err != nil {
			log.Fatalf("cannot locate %s", dir)
		}
		dir = d
	}

	fin, err := os.Lstat(dir)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if !fin.IsDir() {
		log.Fatalf("%s isn't a directory", dir)
	}

	sb := &strings.Builder{}

	filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if dir == path {
			return nil
		}
		if err != nil {
			return nil
		}
		fmt.Fprintf(sb, "%s")
	})
}
