package pb_benchmark

import (
	"reflect"
	"strings"
)

func NewInstance(rt reflect.Type) reflect.Value {
	el := reflect.New(rt).Elem()

	for fieldIdx := 0; fieldIdx < rt.NumField(); fieldIdx++ {
		field := rt.Field(fieldIdx)
		if !field.IsExported() {
			continue
		}

		spec, ok := supported.Find(field.Type.Kind())
		if !ok {
			//fmt.Println("warn: no spec found")
			continue
		}

		fieldInstance := spec.New(field.Type)
		el.Field(fieldIdx).Set(fieldInstance)
	}

	return el
}

type Spec struct {
	Kind reflect.Kind
	New  func(elType reflect.Type) reflect.Value
}

type SpecConfig []Spec

var supported = SpecConfig{}

func init() {
	args := SpecConfig{
		{
			Kind: reflect.String,
			New: func(elType reflect.Type) reflect.Value {
				ins := reflect.New(reflect.TypeOf("")).Elem()
				ins.SetString(strings.Repeat("x", 16))
				return ins
			},
		},
		{
			Kind: reflect.Int,
			New: func(elType reflect.Type) reflect.Value {
				ins := reflect.New(reflect.TypeOf(int(0))).Elem()
				ins.SetInt(1234567890)
				return ins
			},
		},
		{
			Kind: reflect.Int32,
			New: func(elType reflect.Type) reflect.Value {
				ins := reflect.New(reflect.TypeOf(int32(0))).Elem()
				ins.SetInt(1234567890)
				return ins
			},
		},
		{
			Kind: reflect.Uint32,
			New: func(elType reflect.Type) reflect.Value {
				ins := reflect.New(reflect.TypeOf(uint32(0))).Elem()
				ins.SetUint(1234567890)
				return ins
			},
		},
		{
			Kind: reflect.Int64,
			New: func(elType reflect.Type) reflect.Value {
				ins := reflect.New(reflect.TypeOf(int64(0))).Elem()
				ins.SetInt(1234567890)
				return ins
			},
		},
		{
			Kind: reflect.Uint64,
			New: func(elType reflect.Type) reflect.Value {
				ins := reflect.New(reflect.TypeOf(uint64(0))).Elem()
				ins.SetUint(1234567890)
				return ins
			},
		},

		{
			Kind: reflect.Slice,
			New: func(elType reflect.Type) reflect.Value {
				length := 32
				if elType.Elem().Kind() == reflect.Uint8 {
					return reflect.ValueOf([]byte("hello world"))
				}
				slice := reflect.MakeSlice(elType, 0, length)
				for i := 0; i < length; i++ {
					var v reflect.Value
					if elType.Elem().Kind() != reflect.Pointer {
						v = NewInstance(elType.Elem())
					} else {
						v = NewInstance(elType.Elem().Elem())
						v = v.Addr()
					}
					slice = reflect.Append(slice, v)
				}
				return slice
			},
		},
		{
			Kind: reflect.Struct,
			New: func(elType reflect.Type) reflect.Value {
				return NewInstance(elType)
			},
		},
		{
			Kind: reflect.Map,
			New: func(elType reflect.Type) reflect.Value {
				m := map[int32]int32{}
				for i := 0; i < 10; i++ {
					m[int32(i)] = int32(i)
				}
				return reflect.ValueOf(m)
			},
		},
	}
	supported = args
}

func (c SpecConfig) Find(kind reflect.Kind) (Spec, bool) {
	for _, v := range c {
		if v.Kind == kind {
			return v, true
		}
	}
	return Spec{}, false
}
