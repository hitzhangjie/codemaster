package pb_benchmark

import (
	"reflect"
	"testing"

	pbapiv2 "apiv2/pb"

	pbapiv2_proto "github.com/golang/protobuf/proto"
)

func Benchmark_HelloRequest(b *testing.B) {
	r2 := &pbapiv2.HelloRequest{}

	for i := 0; i < b.N; i++ {
		pbapiv2_proto.Marshal(r2)
	}
}

func Benchmark_HelloRequest_NotEmpty(b *testing.B) {
	v := NewInstance(reflect.TypeOf(pbapiv2.HelloRequest{}))
	r2 := v.Interface().(pbapiv2.HelloRequest)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pbapiv2_proto.Marshal(&r2)
	}
}
