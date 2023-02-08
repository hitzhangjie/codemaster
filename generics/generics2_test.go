package generics

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/hitzhangjie/codemaster/pb/hello"
)

var dat []byte

func Test_doSomething(t *testing.T) {
	req := hello.HelloReq{
		Code: 100,
		Msg:  "xxxx",
	}

	dat, _ = json.Marshal(&req)

	doSomething(&req)
}

func doSomething[T any, E interface {
	proto.Message
	*T
}](val E) {
	var value T
	json.Unmarshal(dat, &value)
}
