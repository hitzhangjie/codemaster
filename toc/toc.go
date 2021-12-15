// Program `toc` scans your markdown directory and automatically
// generates `SUMMARY.md`, you can use it in this way:
//
// ```
// toc [-d <dir>] > SUMMARY.md
// ```
// It is useful when you organize markdowns in different folders,
// similar cases are gitbook, hugo, etc. By `toc` we don't need
// to manually write SUMMARY.md now.
//
// `toc` will rewrite SUMMARY.md according to your filesystem
// hierarchy, and liquid tag `weight` in same folder.
//
// For example, if we have following files:
//
// ``````````````````````````````````````````````````````````
// .
// |-1
//   - 1.1
//   - 1.2
// |-2
//   - 2.1
//   - 2.2
//
// ``````````````````````````````````````````````````````````
// then `toc > SUMMARY.md` will generate following content:
//
// file: SUMMARY.md
// ``````````````````````````````````````````````````````````
// # Summary
// ---
// headless: true
// bookhidden: true
// ---
//
// * [1](1)
//   * [1.1](1.1)
//   * [1.2](1.2)
// * [2](2)
//   * [2.1](2.1)
//   * [2.2](2.2)
//
// ``````````````````````````````````````````````````````````
//
// Hope `toc` could ease your hand!
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
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

	ignored := map[string]bool{
		"_index.md": true,
		".DS_Store": true,
	}

	var nodes nodes
	err = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if dir == path {
			return nil
		}
		if err != nil {
			return nil
		}
		if _, ok := ignored[filepath.Base(path)]; ok {
			return nil
		}

		var (
			n = &node{
				name:     fileName(path),
				path:     path,
				subnodes: nil,
			}
			ppath = filepath.Join(filepath.Dir(path), "_index.md")
		)

		if info.IsDir() {
			n.path = filepath.Join(path, "_index.md")
		}
		if w, err := readWeight(n.path); err == nil {
			n.weight = w
		}

		p, ok := nodes.find(ppath)
		if !ok {
			nodes = append(nodes, n)
		} else {
			p.addSubNode(n)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("filewalk err: %v", err)
	}

	fmt.Println(nodes.String())
}

func fileName(fp string) string {
	baseName := filepath.Base(fp)
	return strings.TrimSuffix(baseName, ".md")
}

type node struct {
	name     string
	path     string
	weight   int
	indent   int
	subnodes nodes
}

func readWeight(fp string) (int, error) {
	fin, err := os.Open(fp)
	if err != nil {
		return 0, err
	}
	sc := bufio.NewScanner(fin)
	for sc.Scan() {
		s := strings.TrimSpace(sc.Text())
		if !strings.Contains(s, "weight") {
			continue
		}
		w := strings.TrimSpace(strings.Split(s, ":")[1])
		v, err := strconv.ParseInt(w, 10, 64)
		if err != nil {
			return 0, err
		}
		return int(v), nil
	}
	return 0, errors.New("weight not found")
}
