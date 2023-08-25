package pb_benchmark

import (
	"reflect"
	"testing"

	gogofast "gogo/pb"

	"github.com/gogo/protobuf/proto"
)

// Benchmark_HelloRequest-10    	68195216	        16.45 ns/op	       0 B/op	       0 allocs/op
func Benchmark_HelloRequest(b *testing.B) {
	r3 := &gogofast.HelloRequest{}

	for i := 0; i < b.N; i++ {
		proto.Marshal(r3)
	}
}

func Benchmark_HelloRequest_NotEmpty(b *testing.B) {
	v := NewInstance(reflect.TypeOf(gogofast.HelloRequest{}))
	r2 := v.Interface().(gogofast.HelloRequest)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proto.Marshal(&r2)
	}
}
