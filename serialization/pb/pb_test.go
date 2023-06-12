package pb_test

import (
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/hitzhangjie/codemaster/serialization/pb/hello"
)

func Test_JSONPB_Marshal_DefaultValue(t *testing.T) {
	// 模拟testone代理返回的trpc服务的响应数据
	req := hello.HelloReq{
		Code: 0,
		Msg:  "",
	}

	buf, err := proto.Marshal(&req)
	assert.Nil(t, err)
	assert.Len(t, buf, 0, "pb3 message with default value fields, after marshalled length should be 0")

	// 模拟协议中台先proto反序列化，然后jsonpb再marshal
	req2 := hello.HelloReq{}

	err = proto.Unmarshal(buf, &req2)
	assert.Nil(t, err)

	m := jsonpb.Marshaler{
		OrigName:     true,
		EnumsAsInts:  true,
		EmitDefaults: true,
	}
	s, err := m.MarshalToString(&req2)
	assert.Nil(t, err)
	assert.Equal(t, "{\"code\":0,\"msg\":\"\"}", s)
}

func Test_ProtoMarshalUnMarshal_DifferentMessage_WithSameStructure(t *testing.T) {
	r1 := hello.HelloReq{
		Code: 100,
		Msg:  "hello",
	}
	b, err := proto.Marshal(&r1)
	assert.Nil(t, err)

	r2 := hello.HelloReqX{}
	err = proto.Unmarshal(b, &r2)
	assert.Nil(t, err)

	assert.Equal(t, r1.Code, r2.Code)
	assert.Equal(t, r1.Msg, r2.Msg)
}
