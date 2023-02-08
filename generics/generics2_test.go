package generics

import (
	"fmt"
	"testing"
	"unsafe"

	"google.golang.org/protobuf/proto"

	"github.com/hitzhangjie/codemaster/pb/hello"
)

var dat []byte

func Test_doSomething(t *testing.T) {
	req := hello.HelloReq{
		Code: 100,
		Msg:  "xxxx",
	}

	dat, _ = proto.Marshal(&req)

	doSomething[hello.HelloReq](&req)

}

func doSomething[T any, E interface {
	proto.Message
	*T
}](val E) {
	var value T
	v := (E)(unsafe.Pointer(&value))
	_ = proto.Unmarshal(dat, v)
	fmt.Println(v)
}
