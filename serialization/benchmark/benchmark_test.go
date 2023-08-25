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
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/bytedance/sonic"
	jsoniter "github.com/json-iterator/go"
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

		// foreach testcase
		b.Run(testcase, func(b *testing.B) {
			dat, err := os.ReadFile(f)
			if err != nil {
				panic(err)
			}
			// foreach marshaler
			for _, marshaler := range marshalers {
				b.Run(marshaler.Name, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						m := []any{}
						err := marshaler.Unmarshal(dat, &m)
						if err != nil {
							panic(err)
						}
					}
				})
			}
		})
	}
}

func Benchmark_Unmarshal_Slice_HasSchema(b *testing.B) {
	for _, f := range testfiles {
		testcase := filepath.Base(f)

		// foreach testcase
		b.Run(testcase, func(b *testing.B) {
			dat, err := os.ReadFile(f)
			if err != nil {
				panic(err)
			}
			// foreach marshaler
			for _, marshaler := range marshalers {
				b.Run(marshaler.Name, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						m := []def.Person{}
						err := marshaler.Unmarshal(dat, &m)
						if err != nil {
							panic(err)
						}
					}
				})
			}
		})
	}
}

func Benchmark_Marshal_Slice_HasNoSchema(b *testing.B) {
	for _, f := range testfiles {
		testcase := filepath.Base(f)

		// foreach testfile
		b.Run(testcase, func(b *testing.B) {
			dat, err := os.ReadFile(f)
			if err != nil {
				panic(err)
			}
			m := []any{}
			if err := json.Unmarshal(dat, &m); err != nil {
				panic(err)
			}

			// foreach marshaler
			for _, marshaler := range marshalers {
				b.Run(marshaler.Name, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						_, err := marshaler.Marshal(m)
						if err != nil {
							panic(err)
						}
					}
				})
			}
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

		// foreach testfile
		b.Run(arg.desc, func(b *testing.B) {
			// foreach marshaler
			for _, marshaler := range marshalers {
				b.Run(marshaler.Name, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						_, err := marshaler.Marshal(sliceValue)
						if err != nil {
							panic(err)
						}
					}
				})
			}
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

type Marshaler struct {
	Name      string
	Marshal   func(any) ([]byte, error)
	Unmarshal func([]byte, any) error
}

var marshalers = []Marshaler{
	{"Go/encoding/json", json.Marshal, json.Unmarshal},
	{"Bytedance/sonic-default", sonic.Marshal, sonic.Unmarshal},
	{"Bytedance/sonic-compatmode", sonic.ConfigStd.Marshal, sonic.ConfigStd.Unmarshal},
	{"Bytedance/sonic-perfmode", sonic.ConfigFastest.Marshal, sonic.ConfigFastest.Unmarshal},
	{"Segmentio/json-compatmode", segmentio_json.Marshal, segmentio_json.Unmarshal},
	{"Segmentio/json-perfmode", segmentio_marshal_fast, segmentio_unmarshal_fast},
	{"jsoniter/go", jsoniter.Marshal, jsoniter.Unmarshal},
}

var bp = sync.Pool{
	New: func() any { return make([]byte, 0, 256) },
}

func segmentio_marshal_fast(v any) ([]byte, error) {
	b := bp.Get().([]byte)
	defer bp.Put(b)

	buf := bytes.NewBuffer(b)
	enc := segmentio_json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetAppendNewline(false)
	enc.SetSortMapKeys(false)
	enc.SetTrustRawMessage(true)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func segmentio_unmarshal_fast(b []byte, v any) error {
	dec := segmentio_json.NewDecoder(bytes.NewBuffer(b))
	dec.ZeroCopy()
	dec.DontMatchCaseInsensitiveStructFields()
	return dec.Decode(v)
}
