package reflect

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	pb "github.com/hitzhangjie/codemaster/reflect/pb"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	Sex  int    // 0:male, 1:female
}

func TestInterfaceDynamicValueStruct(t *testing.T) {

	var inf interface{} = &Person{
		Name: "zhangjie",
		Age:  18,
		Sex:  0,
	}

	dt := reflect.TypeOf(inf).Elem()
	dv := reflect.ValueOf(inf).Elem()

	//for i := 0; i < dt.NumField(); i++ {
	//	f := dv.Field(i)
	//	fmt.Printf("%s %s = %v", dt.Field(i).Name, f.Type(), f.Interface())
	//}

	vals := url.Values{}
	for i := 0; i < dt.NumField(); i++ {
		f := dv.Field(i)
		tag := dt.Field(i).Tag.Get("json")
		if len(tag) != 0 {
			vals.Add(dt.Field(i).Tag.Get("json"), fmt.Sprintf("%v", f.Interface()))
			continue
		}
		vals.Add(dt.Field(i).Name, fmt.Sprintf("%v", f.Interface()))
	}
	s := vals.Encode()
	println(s)
}

func TestXXX(t *testing.T) {

	type TT struct {
		A int
		B string
	}

	tt := TT{23, "skidoo"}
	s := reflect.ValueOf(&tt).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("%d: %s %s = %v\n", i,
			typeOfT.Field(i).Name, f.Type(), f.Interface())
	}
}

func TestElementAndInterface(t *testing.T) {
	rv := reflect.ValueOf(&Student{
		Name: "xiaozhang",
		Age:  100,
	})
	stu := rv.Interface().(*Student)
	t.Logf("student: %v", stu)

	stu2 := rv.Elem().Interface().(Student)
	t.Logf("student: %v", stu2)
}

func TestJsonPBMessageInt64(t *testing.T) {
	m := &JSONPBSerialization{}
	p := pb.Player{
		Uid:   100,
		Level: 1,
	}
	s, err := m.Marshal(p)
	assert.Nil(t, err)
	t.Logf("marshal: %s\n", string(s))

	s, err = m.Marshal(&p)
	assert.Nil(t, err)
	t.Logf("marshal: %s\n", string(s))

	type Data struct {
		A int
		B int
	}
	d := Data{
		A: 1,
		B: 1,
	}
	_, err = m.Marshal(d)
	assert.NotNil(t, err)
}
