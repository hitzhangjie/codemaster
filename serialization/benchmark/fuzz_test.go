package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	// repeat: 10, 4k
	// repeat: 20, 10k
	// repeat: 50, 24k
	// repeat: 100, 49k
	// repeat: 200, 98k
	// repeat: 500, 245k
	// repeat: 1000, 491k
	repeatTimes := []int{10, 20, 50, 100, 200, 500, 1000}
	for _, repeatTime := range repeatTimes {
		slice := reflect.MakeSlice(reflect.TypeOf([]Person{}), 0, 3)
		rt := reflect.TypeOf(Person{})

		for i := 0; i < repeatTime; i++ {
			//slice = reflect.Append(slice, newInstance(rt))
			slice = reflect.Append(slice, newInstance(rt))
		}
		//spew.Dump(slice.Interface())
		buf, err := json.Marshal(slice.Interface())
		if err != nil {
			panic(err)
		}
		fmt.Printf("repeatTimes: %d size: %d-KB\n", repeatTime, len(buf)/1024)
		fp, err := filepath.Abs(fmt.Sprintf("testdata/%d-KB.gen.json", len(buf)/1024))
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(fp, buf, os.ModePerm); err != nil {
			panic(err)
		}
	}
}
