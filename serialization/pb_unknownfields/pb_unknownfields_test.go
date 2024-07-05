package pbunknownfields_test

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hitzhangjie/codemaster/serialization/pb_unknownfields/hello"
	"github.com/stretchr/testify/require"
)

func Test_marshal_with_unknownfields(t *testing.T) {
	// 假定这是新版本协议对应的数据
	req := hello.HelloReq{
		Code: 100,
		Msg:  "hello",
	}
	dat, err := proto.Marshal(&req)
	require.Nil(t, err)
	require.NotEmpty(t, dat)

	// 假定这是旧版本协议对应的数据，缺少字段Msg定义
	reqOld := hello.HelloReqX{}
	err = proto.Unmarshal(dat, &reqOld)
	require.Nil(t, err)

	// 现在尝试对旧版本协议对应数据进行marshal，再用新版本协议进行unmarshal，看看pb对unknownfields是如何处理的，会不会丢失
	dat2, err := proto.Marshal(&reqOld)
	require.Nil(t, err)
	require.NotEmpty(t, dat2)

	// 再次反序列化，看看新协议中添加的字段是否还存在
	req2 := hello.HelloReq{}
	err = proto.Unmarshal(dat2, &req2)
	require.Nil(t, err)

	// 存在，原因是：
	// - 按照旧协议进行unmarshal的时候，如果遇到不认识的tag，会放到unknownfields里面，这些并没有丢，只是当前无法处理
	// - 对unmarshal后的message再次进行marshal时，会直接将knownfields里面的数据（tagvalue...）给追加进去，所以是不会丢的。
	//   尽管不会丢，但是逻辑层面因为前一步确实不认识这些新增字段，可能应该更新但是没更新，还是会导致一些不一致数据、程序后续的异常等。
	//
	// see:
	//
	// ```go
	// func (o MarshalOptions) marshalMessageSlow(...) ([]byte, error) {
	//     ....
	//     b = append(b, m.GetUnknown()...)
	//	   return b, nil
	// }
	//
	require.Equal(t, req.Code, req2.Code)
	require.Equal(t, req.Msg, req2.Msg)
}
