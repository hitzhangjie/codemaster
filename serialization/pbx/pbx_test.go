package pbx

import (
	"testing"

	"github.com/hitzhangjie/codemaster/serialization/pbx/git.code.oa.com/examples/hello" // 最新protoc-gen-go生成

	protoOld "github.com/golang/protobuf/proto"
	protoNew "google.golang.org/protobuf/proto"

	helloworldNew "github.com/hitzhangjie/codemaster/serialization/pbx/git.code.oa.com/examples/helloworld" // 最新protoc-gen-go生成
	helloworldOld "google.golang.org/grpc/examples/helloworld/helloworld"

	"github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////
// 新版本protoc-gen-go生成代码

func Test_PbNew_marshal_Pbold_unmarshal(t *testing.T) {
	req := hello.Req{Msg: "helloworld"}
	dat, err := protoNew.Marshal(&req)
	if err != nil {
		panic(err)
	}
	r := hello.Req{}
	err = protoOld.Unmarshal(dat, &r)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, req.Msg, r.Msg)
}

func Test_PbOld_marshal_PbNew_unmarshal(t *testing.T) {
	req := hello.Req{Msg: "helloworld"}
	dat, err := protoOld.Marshal(&req)
	if err != nil {
		panic(err)
	}
	r := hello.Req{}
	err = protoNew.Unmarshal(dat, &r)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, req.Msg, r.Msg)
}

///////////////////////////////////////////////////////////////////////
// 老版本protoc-gen-go生成代码

//func Test_PbNew_marshal_PbOld_unmarshal(t *testing.T) {
//	req := helloworld.HelloRequest{
//		Name: "xxxxxxxxxxx",
//	}
//
//	// fuck??? 老版本的protoc-gen-go生成的pb.go不满足google.golang.org/protobuf/proto.Message定义
//	dat, err := protoNew.Marshal(&req)
//	assert.Nil(t, err)
//}

//func Test_PbOld_marshal_PbNew_unmarshal(t *testing.T) {
//	req := helloworld.HelloRequest{
//		Name:                 "",
//	}
//
//	dat, err := protoOld.Marshal(&req)
//	assert.Nil(t, err)
//
//	r := helloworld.HelloReply{}
//
//	// fuck??? 老版本的protoc-gen-go生成的pb.go不满足google.golang.org/protobuf/proto.Message定义
//	err = protoNew.Unmarshal(dat, &r)
//	assert.Nil(t, err)
//}

func Test_PbOld_marshal_PbNew_unmarshal_X(t *testing.T) {
	req := helloworldOld.HelloRequest{
		Name: "helloworld",
	}
	dat, err := protoOld.Marshal(&req)
	assert.Nil(t, err)

	r := helloworldNew.HelloRequest{}
	err = protoNew.Unmarshal(dat, &r)
	assert.Nil(t, err)
	assert.Equal(t, req.Name, r.Name)
}
