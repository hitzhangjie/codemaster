package pb_v1v2compat

import (
	"testing"

	"github.com/hitzhangjie/codemaster/serialization/pb_v1v2compat/git.code.oa.com/examples/hello" // 最新protoc-gen-go生成
	"github.com/hitzhangjie/codemaster/serialization/pb_v1v2compat/git.code.oa.com/examples/helloworld"

	protoOld "github.com/golang/protobuf/proto"
	protoNew "google.golang.org/protobuf/proto"

	// 最新protoc-gen-go生成
	helloworldOld "google.golang.org/grpc/examples/helloworld/helloworld"

	"github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////
// 新版本protoc-gen-go生成代码

// 未避免protobuf同一个pb文件多次注册后panic的问题，测试时指定如下选项：
// - go test -v -ldflags='-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn'
// - 也可以通过环境变量的方式
// see: https://developers.google.com/protocol-buffers/docs/reference/go/faq#namespace-conflict
//
// 考虑到IDE里测试的方便，我们用第一种方式，将上述-ldflags加到工程的默认testEnvVars设置里。

// 测试case：
//   - pb桩代码使用最新的google.golang.org下提供的protoc-gen-go生成，实际上github.com/golang/protobuf是基于前面这个库重写的
//     后者没有反射能力，但是生成的消息在序列化、反序列化时的兼容性还是有保证的
//   - 先用最新版的proto库marshal，然后再用旧版本的proto库unmarshal（旧只是说repo是旧的，实际上 (v1.4) 是基于新repo重写后的）
//
// 新旧repo的序列化、反序列化逻辑应该保证是兼容的。
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

// 用老的repo marshal，再用新的repo unmarshal，也应该保证是ok的
// 原因前面已经提过了，旧的repo也是用最新的重写过的……我们引用的是重写后的version
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
//
// github.com/golang/protobuf, git checkout v1.3.5, 然后 go install .../protoc-gen-go
// 用这个旧版本的protoc-gen-go，来生成helloworld.proto对应的pb

func Test_OldStub_PbNew_marshal_PbOld_unmarshal(t *testing.T) {
	t.Skip()

	req := helloworld.HelloRequest{
		Name: "xxxxxxxxxxx",
	}
	_ = req

	// 老版本的protoc-gen-go生成的pb.go不满足google.golang.org/protobuf/proto.Message定义
	// 编译不过！

	// _, err := protoNew.Marshal(&req)
	// assert.Nil(t, err)
}

func Test_OldStub_PbOld_marshal_PbNew_unmarshal(t *testing.T) {
	t.Skip()

	req := helloworld.HelloRequest{
		Name: "",
	}

	// 新版本桩代码是满足原来的MessageV1定义的，但是不满足MessageV2定义
	dat, err := protoOld.Marshal(&req)
	assert.Nil(t, err)
	_ = dat

	r := helloworld.HelloReply{}
	_ = r

	// 老版本的protoc-gen-go生成的pb.go不满足google.golang.org/protobuf/proto.Message定义
	// 编译不过！

	// err = protoNew.Unmarshal(dat, &r)
	// assert.Nil(t, err)
}

func Test_OldStub_PbOld_marshal_PbNew_unmarshal_X(t *testing.T) {
	req := helloworldOld.HelloRequest{
		Name: "helloworld",
	}
	dat, err := protoOld.Marshal(&req)
	assert.Nil(t, err)
	_ = dat

	// 老版本的protoc-gen-go生成的pb.go不满足google.golang.org/protobuf/proto.Message定义
	// 编译不过！

	// r := helloworldNew.HelloRequest{}
	// err = protoNew.Unmarshal(dat, &r)
	// assert.Nil(t, err)
	// assert.Equal(t, req.Name, r.Name)
}
