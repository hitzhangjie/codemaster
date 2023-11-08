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
//
// 测试结果说明：
// - 这里是用的不同大小的json数据去unmarshal成slice、或者将这样大小的slice再反过来marshal，
// - 然后，slice分 有schema定义([]def.Person)、无schema([]any)定义 两种情况，
//
// 测试结果显示，序列化反序列化效率基本保持这样的顺序，只是“差距”大小不同，
// 性能排序：bytedance/sonic > segmentio/encoding > jsoniter > stdlib
// 多数数据大小情况下，bytedance/sonic性能大概是stdlib的4~5倍左右。
package benchmark

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/bytedance/sonic"
	go_json "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
	segmentio_json "github.com/segmentio/encoding/json"

	"github.com/hitzhangjie/codemaster/serialization/json_benchmark/def"
)

var testfiles []string

func TestMain(m *testing.M) {
	prepareTestData()

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
	{"goccy/go-json", go_json.Marshal, go_json.Unmarshal},
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
