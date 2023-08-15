package benchmark

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

type Person struct {
	Name    string `fuzz:"length(8,32)"`
	Age     int    `fuzz:"range(1,100)"`
	Address string `fuzz:"length(20,60)"`

	XXXX string `fuzz:"ignore"`
}

var pattern = regexp.MustCompile(`(\w+)\((.+)\)`)

func Test_Generate(t *testing.T) {
	slice := reflect.MakeSlice(reflect.TypeOf([]Person{}), 0, 3)

	rt := reflect.TypeOf(Person{})
	for repeat := 0; repeat < 1; repeat++ {
		el := reflect.New(rt).Elem()

		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			v := f.Tag.Get("fuzz")
			if v == "ignore" {
				continue
			}

			matches := pattern.FindStringSubmatch(v)
			if len(matches) < 3 {
				fmt.Println("warn: invalid fuzz tag", matches)
				continue
			}
			action := matches[1]
			args := strings.Split(matches[2], ",")

			switch action {
			case "length":
				vals, err := parseArgs(args)
				if err != nil {
					panic(err)
				}
				if len(vals) != 2 {
					panic("length(a,b): miss 'a' or 'b'")
				}
				minLen, maxLen := vals[0], vals[1]
				if !(minLen <= maxLen) {
					panic("length(a,b): 'a <= b' not ok")
				}
				fmt.Println("length:", minLen, maxLen)
				chooseLen := rand.Int()%(maxLen-minLen) + minLen
				el.Field(i).SetString(strings.Repeat("x", chooseLen))
			case "range":
				vals, err := parseArgs(args)
				if err != nil {
					panic(err)
				}
				if len(vals) != 2 {
					panic("range(a,b): miss 'a' or 'b'")
				}
				min, max := vals[0], vals[1]
				if !(min <= max) {
					panic("range(a,b): 'a <= b' not ok")
				}
				fmt.Println("range:", min, max)
				chooseVal := rand.Int()%(max-min) + min
				el.Field(i).SetInt(int64(chooseVal))
			default:
			}
			//spew.Dump(el)
		}
		slice = reflect.Append(slice, el)
	}
	spew.Dump(slice.Interface())
}

func parseArgs(args []string) ([]int, error) {
	if len(args) != 2 {
		return nil, errors.New("length mismatch")
	}
	vals := make([]int, len(args))
	for i, arg := range args {
		v, err := strconv.Atoi(arg)
		if err != nil {
			return nil, fmt.Errorf("args[%d] not int", i)
		}
		vals[i] = v
	}
	return vals, nil
}
