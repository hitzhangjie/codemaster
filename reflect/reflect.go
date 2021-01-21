package reflect

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
)

type JSONPBSerialization struct{}

func (s *JSONPBSerialization) Marshal(body interface{}) ([]byte, error) {

	var input proto.Message

	switch v := body.(type) {
	// 如果是protobuf message，直接marshal
	case proto.Message:
		input = v
	default:
		// 如果不是，很可能是传了value过来，自动构造一个，获取地址，然后再marshal
		rt := reflect.TypeOf(body)
		oldValue := reflect.ValueOf(body)
		newValue := reflect.New(rt)

		for i := 0; i < rt.NumField(); i++ {
			c := rt.Field(i).Name[0]
			if c >= 'a' && c <= 'z' {
				continue
			}
			newValue.Elem().Field(i).Set(oldValue.Field(i))

			fmt.Println(newValue.Elem().Field(i))
		}

		//fmt.Printf("%s\n", newValue.Type().String())
		//input = newValue.Interface().(proto.Message)

		fmt.Printf("%s\n", newValue.Type().String())
		var ok bool
		input,ok = newValue.Interface().(proto.Message)
		if !ok {
			return nil, errors.New("value of body or pointer to value of body, not proto.Message")
		}
	}

	//input, ok := body.(proto.Message)
	//if !ok {
	//	var vv interface{}
	//	switch v := body.(type) {
	//	default:
	//		vv = &v
	//	}
	//	return JSONAPI.Marshal(vv)
	//}

	buf := []byte{}
	w := bytes.NewBuffer(buf)
	err := Marshaler.Marshal(w, input)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
