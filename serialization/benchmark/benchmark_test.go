// Package benchmark_test 这里整理了一些常见的json库的序列化、反序列化、get、set性能对比，
// 以及大致整理下json在这些方面面临的挑战，以及各个json库实现做的好的地方。
//
// 主要对比下面这些：
// - stdlib, encoding/json
// - sonic, https://www.libhunt.com/r/bytedance/sonic
// - fastjson, https://www.libhunt.com/r/valyala/fastjson
// - jsoniter, https://www.libhunt.com/r/jsoniter
// - encoding, https://www.libhunt.com/r/encoding
// - simdjson, https://www.libhunt.com/r/simdjson
// - simdjson-go, https://www.libhunt.com/r/simdjson-go
// - rapidjson, https://www.libhunt.com/r/rapidjson
package benchmark

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/bytedance/sonic"
	segmentio_json "github.com/segmentio/encoding/json"

	"github.com/hitzhangjie/codemaster/serialization/benchmark/def"
)

var testfiles []string

func TestMain(m *testing.M) {
	files, err := collectTestfiles()
	if err != nil {
		panic(err)
	}
	testfiles = files

	os.Exit(m.Run())
}

func Benchmark_Unmarshal_Slice_HasNoSchema(b *testing.B) {
	for _, f := range testfiles {
		testcase := filepath.Base(f)
		b.Run(testcase, func(b *testing.B) {
			dat, err := os.ReadFile(f)
			if err != nil {
				panic(err)
			}
			b.Run("Go/encoding/json", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m := []any{}
					err := json.Unmarshal(dat, &m)
					if err != nil {
						panic(err)
					}
				}
			})
			b.Run("Bytedance/sonic", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m := []any{}
					err := sonic.Unmarshal(dat, &m)
					if err != nil {
						panic(err)
					}
				}
			})
			b.Run("segmentio/encoding", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m := []any{}
					err := segmentio_json.Unmarshal(dat, &m)
					if err != nil {
						panic(err)
					}
				}
			})
		})
	}
}

func Benchmark_Unmarshal_Slice_HasSchema(b *testing.B) {
	for _, f := range testfiles {
		testcase := filepath.Base(f)
		b.Run(testcase, func(b *testing.B) {
			dat, err := os.ReadFile(f)
			if err != nil {
				panic(err)
			}
			b.Run("Go/encoding/json", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m := []def.Person{}
					err := json.Unmarshal(dat, &m)
					if err != nil {
						panic(err)
					}
				}
			})
			b.Run("Bytedance/sonic", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m := []def.Person{}
					err := sonic.Unmarshal(dat, &m)
					if err != nil {
						panic(err)
					}
				}
			})
			b.Run("segmentio/encoding", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					m := []any{}
					err := segmentio_json.Unmarshal(dat, &m)
					if err != nil {
						panic(err)
					}
				}
			})

		})
	}
}

func Benchmark_Marshal_Slice_HasNoSchema(b *testing.B) {
	for _, f := range testfiles {
		testcase := filepath.Base(f)
		b.Run(testcase, func(b *testing.B) {
			dat, err := os.ReadFile(f)
			if err != nil {
				panic(err)
			}
			m := []any{}
			if err := json.Unmarshal(dat, &m); err != nil {
				panic(err)
			}

			b.Run("Go/encoding/json", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_, err := json.Marshal(m)
					if err != nil {
						panic(err)
					}
				}
			})
			b.Run("Bytedance/sonic", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_, err := sonic.Marshal(m)
					if err != nil {
						panic(err)
					}
				}
			})
			b.Run("segmentio/encoding", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_, err := segmentio_json.Marshal(m)
					if err != nil {
						panic(err)
					}
				}
			})
		})
	}
}

func Benchmark_Marshal_Slice_HasSchema(b *testing.B) {
	type arg struct {
		desc  string
		elems int
	}
	// see fuzz_test.go
	args := []arg{
		{"4KB", 10}, {"10KB", 20}, {"24KB", 50}, {"49KB", 100},
		{"98KB", 200}, {"245KB", 500}, {"491KB", 1000}}

	for _, arg := range args {
		slice := reflect.MakeSlice(reflect.TypeOf([]def.Person{}), 0, arg.elems)
		rt := reflect.TypeOf(def.Person{})

		for i := 0; i < arg.elems; i++ {
			slice = reflect.Append(slice, newInstance(rt))
		}
		sliceValue := slice.Interface()

		b.Run(arg.desc, func(b *testing.B) {
			b.Run("Go/encoding/json", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_, err := json.Marshal(sliceValue)
					if err != nil {
						panic(err)
					}
				}
			})
			b.Run("Bytedance/sonic", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_, err := sonic.Marshal(sliceValue)
					if err != nil {
						panic(err)
					}
				}
			})
			b.Run("segmentio/encoding", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_, err := segmentio_json.Marshal(sliceValue)
					if err != nil {
						panic(err)
					}
				}
			})
		})
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
		i1, i2 := strings.Index(f1, "-"), strings.Index(f2, "-")
		s1, s2 := f1[:i1], f2[:i2]
		v1, _ := strconv.Atoi(s1)
		v2, _ := strconv.Atoi(s2)
		return v1 < v2
	})
	return files, nil
}
