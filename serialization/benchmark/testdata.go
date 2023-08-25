package benchmark

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/hitzhangjie/codemaster/serialization/benchmark/def"
)

func prepareTestData() {
	d, err := filepath.Abs("testdata")
	if err != nil {
		panic(err)
	}
	_ = os.RemoveAll(d)
	_ = os.MkdirAll(d, os.ModePerm)

	// repeat: 10, 4k
	// repeat: 20, 10k
	// repeat: 50, 24k
	// repeat: 100, 49k
	// repeat: 200, 98k
	// repeat: 500, 245k
	// repeat: 1000, 491k
	repeatTimes := []int{1, 5, 10, 20, 50, 100, 200, 500, 1000}
	for _, repeatTime := range repeatTimes {
		slice := reflect.MakeSlice(reflect.TypeOf([]def.Person{}), 0, 3)
		rt := reflect.TypeOf(def.Person{})

		for i := 0; i < repeatTime; i++ {
			//slice = reflect.Append(slice, newInstance(rt))
			slice = reflect.Append(slice, newInstance(rt))
		}
		//spew.Dump(slice.Interface())
		buf, err := json.Marshal(slice.Interface())
		if err != nil {
			panic(err)
		}
		var fp string
		if n := len(buf); n >= 1024 {
			//fmt.Printf("repeatTimes: %d size: %d-KB\n", repeatTime, n/1024)
			fp, err = filepath.Abs(fmt.Sprintf("testdata/%d-KB.gen.json", n/1024))
		} else {
			//fmt.Printf("repeatTimes: %d size: %d-B\n", repeatTime, n)
			fp, err = filepath.Abs(fmt.Sprintf("testdata/%d-B.gen.json", n))
		}
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(fp, buf, os.ModePerm); err != nil {
			panic(err)
		}
	}
}

func collectTestfiles() ([]string, error) {
	var files []string

	testdata, err := filepath.Abs("testdata")
	if err != nil {
		return nil, err
	}

	_, err = os.Lstat(testdata)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		_ = os.MkdirAll(testdata, os.ModePerm)
	}

	err = filepath.Walk(testdata, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool {
		f1, f2 := filepath.Base(files[i]), filepath.Base(files[j])
		sz1, _, _ := strings.Cut(f1, ".")
		sz2, _, _ := strings.Cut(f2, ".")
		ss1 := strings.Split(sz1, "-")
		ss2 := strings.Split(sz2, "-")
		v1, _ := strconv.Atoi(ss1[0])
		if ss1[1] == "KB" {
			v1 *= 1024
		}
		v2, _ := strconv.Atoi(ss2[0])
		if ss2[1] == "KB" {
			v2 *= 1024
		}
		return v1 < v2
	})
	return files, nil
}
