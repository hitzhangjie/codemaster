package benchmark

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var pattern = regexp.MustCompile(`(\w+)\((.+)\)`)

func newInstance(rt reflect.Type) reflect.Value {
	el := reflect.New(rt).Elem()

	for fieldIdx := 0; fieldIdx < rt.NumField(); fieldIdx++ {
		field := rt.Field(fieldIdx)
		if !field.IsExported() {
			continue
		}

		fuzzVal := field.Tag.Get("fuzz")
		if fuzzVal == "ignore" {
			continue
		}

		var action string
		var matches []string
		var params []string

		if field.Type.Kind() != reflect.Struct {
			matches = pattern.FindStringSubmatch(fuzzVal)
			if len(matches) < 3 {
				fmt.Println("warn: invalid fuzz tag", matches)
				continue
			}
			action = matches[1]
			params = strings.Split(matches[2], ",")
		} else {
			action = "STRUCT"
		}

		spec, ok := supported.Find(field.Type.Kind(), action)
		if !ok {
			fmt.Println("warn: no spec found")
			continue
		}

		args, err := spec.ParserE(params)
		if err != nil {
			panic(fmt.Errorf("parse field:%s err:%v, spec:%v", field.Name, err, spec.Rule))
		}
		fieldInstance := spec.New(args, field.Type)
		el.Field(fieldIdx).Set(fieldInstance)
	}

	return el
}

type Spec struct {
	Kind    reflect.Kind
	Rule    string
	Desc    string
	ParserE func(args []string) ([]any, error)
	New     func(args []any, elType reflect.Type) reflect.Value
}

func (s Spec) action() string {
	if s.Rule == "" {
		return "STRUCT"
	}

	matches := pattern.FindStringSubmatch(s.Rule)
	if len(matches) < 3 {
		fmt.Println("warn: invalid fuzz tag", matches)
		return ""
	}
	return matches[1]
}

type SpecConfig []Spec

var supported = SpecConfig{}

func init() {
	args := SpecConfig{
		{
			// 字符串规则1：生成字符串，填充指定长度即可
			Kind: reflect.String,
			Rule: "length(minLen,maxLen)",
			Desc: "limit the length of generated string, minLen~maxLen runes",
			ParserE: func(args []string) ([]any, error) {
				vals := make([]any, len(args))
				for i, arg := range args {
					v, err := strconv.Atoi(arg)
					if err != nil {
						return nil, fmt.Errorf("args[%d]==%v err: %v", i, args[i], err)
					}
					vals[i] = v
				}
				return vals, nil
			},
			New: func(args []any, elType reflect.Type) reflect.Value {
				if len(args) != 2 {
					panic("length(a,b): string length btw a and b, err: miss 'a' or 'b'")
				}
				minLen, maxLen := args[0].(int), args[1].(int)
				if !(minLen <= maxLen) {
					panic("length(a,b): 'a <= b' not ok")
				}
				//fmt.Println("length:", minLen, maxLen)
				chooseLen := rand.Int()%(maxLen-minLen) + minLen
				ins := reflect.New(reflect.TypeOf("")).Elem()
				ins.SetString(strings.Repeat("x", chooseLen))
				return ins
			},
		},
		{
			// 字符串规则2：生成日期，简单按照格式格式化即可
			Kind: reflect.String,
			Rule: "date(2006-01-02)",
			Desc: "limit the date format of generated string",
			ParserE: func(args []string) ([]any, error) {
				vals := make([]any, len(args))
				for i, arg := range args {
					vals[i] = time.Now().Format(arg)
				}
				return vals, nil
			},
			New: func(args []any, elType reflect.Type) reflect.Value {
				if len(args) != 1 {
					panic("date(yyyy-MM-dd): limit date format, err: miss time layout")
				}
				ins := reflect.New(reflect.TypeOf("")).Elem()
				ins.SetString(args[0].(string))
				return ins
			},
		},
		{
			// 整数规则：限定生成整数的范围，即可
			Kind: reflect.Int,
			Rule: "range(min,max)",
			Desc: "limit the value range, min~max",
			ParserE: func(args []string) ([]any, error) {
				vals := make([]any, len(args))
				for i, arg := range args {
					v, err := strconv.Atoi(arg)
					if err != nil {
						return nil, fmt.Errorf("args[%d] not int", i)
					}
					vals[i] = v
				}
				return vals, nil
			},
			New: func(args []any, elType reflect.Type) reflect.Value {
				min, max := args[0].(int), args[1].(int)
				if !(min <= max) {
					panic("range(a,b): 'a <= b' not ok")
				}
				//fmt.Println("range:", min, max)
				chooseVal := rand.Int()%(max-min) + min
				ins := reflect.New(reflect.TypeOf(int(0))).Elem()
				ins.SetInt(int64(chooseVal))
				return ins
			},
		},
		{
			// 数组规则：生成指定的切片，即可
			Kind: reflect.Slice,
			Rule: "length(n)",
			Desc: "limit the slice length, n elements",
			ParserE: func(args []string) ([]any, error) {
				if len(args) != 1 {
					return nil, errors.New("slice rule 'length(n)' has only 1 argument")
				}
				v, err := strconv.Atoi(args[0])
				if err != nil {
					return nil, err
				}
				return []any{v}, nil
			},
			New: func(args []any, elType reflect.Type) reflect.Value {
				length := args[0].(int)
				slice := reflect.MakeSlice(elType, 0, length)
				for i := 0; i < length; i++ {
					v := newInstance(elType.Elem())
					slice = reflect.Append(slice, v)
				}
				return slice
			},
		},
		{
			// 结构体规则：生成对应的struct，即可
			Kind: reflect.Struct,
			Rule: "",
			Desc: "recursive fill the struct",
			ParserE: func(args []string) ([]any, error) {
				return nil, nil
			},
			New: func(args []any, elType reflect.Type) reflect.Value {
				return newInstance(elType)
			},
		},
	}
	supported = args
}

func (c SpecConfig) Find(kind reflect.Kind, action string) (Spec, bool) {
	for _, v := range c {
		if v.Kind == kind && v.action() == action {
			return v, true
		}
	}
	return Spec{}, false
}
