package codec

import (
	"bytes"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	person "github.com/hitzhangjie/codemaster/codec/testcase"
	person2 "github.com/hitzhangjie/codemaster/codec/testcase2"
	"github.com/stretchr/testify/assert"
)

type Person struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       uint16 `json:"age"`
	Sex       uint8  `json:"sex"`
}

func TestJsonMarshal_WithoutOrigName(t *testing.T) {

	p := &person.Person{
		FirstName: "Jie",
		LastName:  "Zhang",
		Age:       29,
		Sex:       1,
	}

	marshaler := jsonpb.Marshaler{
		OrigName:     false, // json序列化的时候是否使用pb中原始的字段名，而非tag名
		EnumsAsInts:  true,  // json序列化的时候是否将枚举值作为int
		EmitDefaults: false, // json序列化的时候是否要包含0值字段
		Indent:       "",
		AnyResolver:  nil,
	}

	// don't use original pb message field name, use the one defined in message field option instead
	buf := bytes.Buffer{}
	err := marshaler.Marshal(&buf, p)
	assert.Nil(t, err)
	t.Logf("marshalled person: %s", buf.String())

	// use original pb message field name instead of the one defined in message field option
	marshaler.OrigName = true
	buf2 := bytes.Buffer{}
	err2 := marshaler.Marshal(&buf2, p)
	assert.Nil(t, err2)
	t.Logf("marshalled person: %s", buf2.String())

	// if no message field option specified
	marshaler.OrigName = false
	px := &person.PersonX{
		FirstName: "Jie",
		LastName:  "Zhang",
	}
	buf3 := bytes.Buffer{}
	err3 := marshaler.Marshal(&buf3, px)
	assert.Nil(t, err3)
	t.Logf("marshaled personx: %s", buf3.String())
}

func TestJsonMarshal_WithDefaultZeroValue(t *testing.T) {
	p := &person.Person{
		FirstName: "Jie",
		LastName:  "Zhang",
		Age:       29,
		Sex:       0,
	}

	marshaler := jsonpb.Marshaler{
		OrigName:     false, // json序列化的时候是否使用pb中原始的字段名，而非tag名
		EnumsAsInts:  true,  // json序列化的时候是否将枚举值作为int
		EmitDefaults: true,  // json序列化的时候是否要包含0值字段
		Indent:       "",
		AnyResolver:  nil,
	}

	buf := bytes.Buffer{}
	err := marshaler.Marshal(&buf, p)
	if err != nil {
		t.Fatalf("jsonpb marshal error: %v", err)
	}
	t.Logf("jsonpb marshal ok, data:\n%s", buf.String())
}

func TestJsonMarshaler_WithPBSyntax2(t *testing.T) {
	p := &person2.Person{
		FirstName: proto.String("Jie"),
		LastName:  proto.String("Zhang"),
		Age:       proto.Uint32(29),
		//Sex:       proto.Uint32(0),
	}

	marshaler := jsonpb.Marshaler{
		OrigName:     false, // json序列化的时候是否使用pb中原始的字段名，而非tag名
		EnumsAsInts:  true,  // json序列化的时候是否将枚举值作为int
		EmitDefaults: true,  // json序列化的时候是否要包含0值字段
		Indent:       "",
		AnyResolver:  nil,
	}

	buf := bytes.Buffer{}
	err := marshaler.Marshal(&buf, p)
	if err != nil {
		t.Fatalf("jsonpb marshal error: %v", err)
	}
	t.Logf("jsonpb marshal ok, data:\n%s", buf.String())
}
