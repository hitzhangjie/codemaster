package pbxx

import (
	"testing"

	protoGoGo "github.com/gogo/protobuf/proto"
	protoGithub "github.com/golang/protobuf/proto"
	protocol "github.com/hitzhangjie/codemaster/pbxx/xxx"
	"github.com/stretchr/testify/assert"
	protoGoogle "google.golang.org/protobuf/proto"
)

func Test_XXXX(t *testing.T) {
	req := HelloRequest{
		Msg: "helloworld",
	}

	// 压测客户端用的gogo/protobuf
	dat, err := protoGoGo.Marshal(&req)
	assert.Nil(t, err)
	assert.NotEmpty(t, dat)

	// 测试github.../protobuf反序列化
	r := HelloRequest{}
	err = protoGithub.Unmarshal(dat, &r)
	assert.Nil(t, err)
	assert.Equal(t, r.Msg, req.Msg)

	// 测试google.../protobuf反序列化
	r2 := protocol.HelloRequestXX{}
	err = protoGoogle.Unmarshal(dat, &r2)
	assert.Nil(t, err)
	assert.Equal(t, r2.Msg, req.Msg)

	// 编译时错误，r3 not a message
	//r3 := HelloRequest{}
	//err = protoGoogle.Unmarshal(dat, &r3)
	//assert.Nil(t, err)
	//assert.Equal(t, r3.Msg, req.Msg)
}
