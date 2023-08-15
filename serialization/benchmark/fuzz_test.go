package benchmark

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type Person struct {
	Name        string      `fuzz:"length(8,32)"`
	Age         int         `fuzz:"range(1,100)"`
	Address     string      `fuzz:"length(20,60)"`
	Education   []Education `fuzz:"length(4)"`
	ContactInfo ContactInfo

	XXXX string `fuzz:"ignore"`
}

type Education struct {
	School string `fuzz:"length(8,16)"`
	From   string `fuzz:"date(2006-01-02)"`
	To     string `fuzz:"date(2006-01-02)"`
}

type ContactInfo struct {
	Mobile string `fuzz:"length(7,11)"`
	Home   string `fuzz:"length(7,11)"`
	Work   string `fuzz:"length(7,11)"`
	Email  string `fuzz:"length(10,64)"`
}

func Test_Generate(t *testing.T) {
	slice := reflect.MakeSlice(reflect.TypeOf([]Person{}), 0, 3)
	rt := reflect.TypeOf(Person{})
	// repeat: 10, 4k
	// repeat: 20, 10k
	// repeat: 50, 24k
	// repeat: 100, 49k
	// repeat: 200, 98k
	// repeat: 500, 245k
	// repeat: 1000, 491k
	for repeat := 0; repeat < 1000; repeat++ {
		//slice = reflect.Append(slice, newInstance(rt))
		slice = reflect.Append(slice, newInstance(rt))
	}
	//spew.Dump(slice.Interface())
	buf, _ := json.Marshal(slice.Interface())
	fmt.Println("size:", len(buf))
}
