package pb_benchmark

import (
	"reflect"
	"testing"

	pbapiv1 "apiv1/pb"

	pbapiv1_proto "github.com/golang/protobuf/proto"
)

func Benchmark_HelloRequest_Empty(b *testing.B) {
	r1 := &pbapiv1.HelloRequest{}

	for i := 0; i < b.N; i++ {
		pbapiv1_proto.Marshal(r1)
	}
}

func Benchmark_HelloRequest_NotEmpty(b *testing.B) {
	v := NewInstance(reflect.TypeOf(pbapiv1.HelloRequest{}))
	r1 := v.Interface().(pbapiv1.HelloRequest)

	// serialized bytes: 3859b
	//buf, _ := pbapiv1_proto.Marshal(&r1)
	//fmt.Println("serialized bytes:", len(buf))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pbapiv1_proto.Marshal(&r1)
	}
}
