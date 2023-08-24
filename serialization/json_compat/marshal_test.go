package json_compat_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/option"
	sjson "github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/require"
)

func TestMarshal_Compat(t *testing.T) {
	msg := map[string]any{
		"a1": 1,
		"b1": 2,
		"a2": 3,
		"b2": 4,
		"a0": 5,
	}

	out := `{"a0":5,"a1":1,"a2":3,"b1":2,"b2":4}`

	t.Run("map-sortkeys", func(t *testing.T) {
		t.Run("encoding/json", func(t *testing.T) {
			b, _ := json.Marshal(msg)
			fmt.Println(string(b))
			require.Equal(t, out, string(b))
		})
		// t.Run("gogo/jsonpb", func(t *testing.T) {
		// 	m := jsonpb.Marshaler{}
		// 	b := bytes.Buffer{}
		// 	// map 不满足 MessageV1
		// 	//m.Marshal(&b, msg)
		// })
		t.Run("segmentio/encoding/json", func(t *testing.T) {
			// 默认，也是强制escapeHTML+sortMapKeys，
			// 目的是在行为上与标准库对齐
			b, _ := sjson.Marshal(msg)
			fmt.Println(string(b))
			require.Equal(t, out, string(b))

			// 但是可以控制是否关闭这些选项
			buf := bytes.NewBuffer(nil)
			enc := sjson.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			enc.SetAppendNewline(false)
			enc.SetSortMapKeys(false) // 去掉mapkeys排序，也就意味去掉了排序前的各种反射操作，获取key名之类的
			enc.SetTrustRawMessage(true)
			_ = enc.Encode(msg)
			fmt.Println(buf.String())
			require.NotEqual(t, out, buf.String())
		})
	})
}

// map序列化时，实测segmentio/encoding胜出，bytedance/sonic性能不及segmentio/encoding，
// 这应该是这个用例中的map太小，没有体现出sonic的优势，我们在大数据结构16KB+测试时发现sonic优势明显。
//
// goos: linux
// goarch: amd64
// BenchmarkMarshal_Map/encoding_json-16             			    600128                1885 ns/op
// BenchmarkMarshal_Map/segmentio_json_compatmode-16                2517423               481.3 ns/op
// BenchmarkMarshal_Map/segmentio_json_perfmode-16                  2951804               404.5 ns/op
// BenchmarkMarshal_Map/bytedance_sonic_compatmode-16               1583278               759.9 ns/op
// BenchmarkMarshal_Map/bytedance_sonic_perfmode-16                 2082262               571.2 ns/op
// BenchmarkMarshal_Map/bytedance_sonic_perfmode#01-16              2082891               579.0 ns/op
func BenchmarkMarshal_Map(b *testing.B) {
	msg := map[string]any{
		"a1": 1,
		"b1": 2,
		"a2": 3,
		"b2": 4,
		"a0": 5,
	}

	// 标准库json序列化
	b.Run("encoding_json", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			json.Marshal(msg)
		}
	})
	// segmentio兼容标注库模式模式 (去掉了一点点不必要的反射)
	b.Run("segmentio_json_compatmode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sjson.Marshal(msg)
		}
	})
	// segmentio性能first模式 (去掉了很多不必要的反射)
	b.Run("segmentio_json_perfmode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(nil)
			enc := sjson.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			enc.SetAppendNewline(false)
			enc.SetSortMapKeys(false)
			enc.SetTrustRawMessage(true)
			_ = enc.Encode(msg)
		}
	})
	// sonic标准库兼容模式（要看amd64架构下的效果，其他平台下未之后simd优化）
	sonic.Pretouch(reflect.TypeOf(msg), option.WithCompileMaxInlineDepth(8), option.WithCompileRecursiveDepth(8))
	b.Run("bytedance_sonic_compatmode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sonic.ConfigStd.Marshal(msg)
		}
	})
	// sonic性能first模式（要看amd64架构下的效果，其他平台下未之后simd优化）
	b.Run("bytedance_sonic_perfmode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sonic.ConfigFastest.Marshal(msg)
		}
	})
	// sonic默认模式（要看amd64架构下的效果，其他平台下未之后simd优化）
	b.Run("bytedance_sonic_perfmode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sonic.ConfigDefault.Marshal(msg)
		}
	})
}

// sonic胜出，看来有schema的情况下可以最大化减少反射，这种情况下sonic表现更突出
//
// goos: linux
// goarch: amd64
// BenchmarkMarshal_Struct/encoding_json-16                         1702755               704.7 ns/op
// BenchmarkMarshal_Struct/segmentio_json_compatmode-16             2798784               427.4 ns/op
// BenchmarkMarshal_Struct/segmentio_json_perfmode-16               2228810               536.2 ns/op
// BenchmarkMarshal_Struct/bytedance_sonic_compatmode-16            2090330               574.5 ns/op
// BenchmarkMarshal_Struct/bytedance_sonic_perfmode-16              2690434               448.2 ns/op
// BenchmarkMarshal_Struct/bytedance_sonic_perfmode#01-16           2681014               450.1 ns/op
func BenchmarkMarshal_Struct(b *testing.B) {
	type XYZ struct {
		F1 string `json:"f1,omitempty"`
		F2 uint64 `json:"f2,omitempty"`
	}

	type Msg struct {
		A1 string `json:"a1,omitempty"`
		A2 string `json:"a2,omitempty"`
		A3 int32  `json:"a3,omitempty"`
		A4 int64  `json:"a4,omitempty"`
		A5 int64  `json:"a5,string,omitempty"`
		X  XYZ    `json:"xxx"`
	}

	msg := Msg{
		A1: "hello world",
		A2: "hello world",
		A3: 1,
		A4: 2,
		A5: 3,
		X: XYZ{
			F1: "hello world",
			F2: 1,
		},
	}

	// 标准库json序列化
	b.Run("encoding_json", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			json.Marshal(msg)
		}
	})
	// segmentio兼容标注库模式模式 (去掉了一点点不必要的反射)
	b.Run("segmentio_json_compatmode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sjson.Marshal(msg)
		}
	})
	// segmentio性能first模式 (去掉了很多不必要的反射)
	b.Run("segmentio_json_perfmode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(nil)
			enc := sjson.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			enc.SetAppendNewline(false)
			enc.SetSortMapKeys(false)
			enc.SetTrustRawMessage(true)
			_ = enc.Encode(msg)
		}
	})
	// sonic标准库兼容模式（要看amd64架构下的效果，其他平台下未之后simd优化）
	sonic.Pretouch(reflect.TypeOf(msg), option.WithCompileMaxInlineDepth(8), option.WithCompileRecursiveDepth(8))
	b.Run("bytedance_sonic_compatmode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sonic.ConfigStd.Marshal(msg)
		}
	})
	// sonic性能first模式（要看amd64架构下的效果，其他平台下未之后simd优化）
	b.Run("bytedance_sonic_perfmode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sonic.ConfigFastest.Marshal(msg)
		}
	})
	// sonic默认模式（要看amd64架构下的效果，其他平台下未之后simd优化）
	b.Run("bytedance_sonic_perfmode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sonic.ConfigDefault.Marshal(msg)
		}
	})

}
