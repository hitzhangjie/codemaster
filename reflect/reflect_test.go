package reflect

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"unsafe"

	pb "github.com/hitzhangjie/codemaster/reflect/pb"
	"github.com/stretchr/testify/assert"
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

type Thing struct {
	Name   string
	Price  float64
	Points []Point
}

type Point struct {
	x float64
	y float64
}

func TestReflectFillStruct(t *testing.T) {

	v := &Thing{}

	rv := reflect.ValueOf(v).Elem()

	// string, float fields
	rv.FieldByName("Name").SetString("something")
	rv.FieldByName("Price").SetFloat(100.0)

	// slice of user-defined type
	f := rv.FieldByName("Points")
	if f.Kind() != reflect.Slice {
		panic("field not slice")
	}

	// fill a new slice
	if f.IsNil() {
		points := reflect.MakeSlice(reflect.TypeOf([]Point{}), 4, 4)
		i := points.Interface()
		i.([]Point)[0] = Point{0, 0}
		i.([]Point)[1] = Point{0, 1}
		i.([]Point)[2] = Point{0, 2}
		i.([]Point)[3] = Point{0, 3}
		f.Set(points)
	}

	fmt.Printf("thing: %+v\n", v)

	f.Interface().([]Point)[0] = Point{4, 4}
	fmt.Printf("thing: %+v\n", v)

	fmt.Printf("Thing.Points addr: %0x\n", f.UnsafeAddr())

	// note: this breaks the unsafe.Pointer rules
	points := *(*[]Point)(unsafe.Pointer(f.UnsafeAddr()))
	fmt.Printf("Thing.Points: %+v\n", points)

	// change existed slice

	// note: prefer this method over unsafe.Pointer
	println(f.CanAddr())
	f.Interface().([]Point)[0] = Point{5, 5}
	fmt.Printf("Thing.Points: %+v\n", points)
}
