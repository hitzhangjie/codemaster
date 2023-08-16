package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hitzhangjie/codemaster/serialization/benchmark/def"
)

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
		slice := reflect.MakeSlice(reflect.TypeOf([]def.Person{}), 0, 3)
		rt := reflect.TypeOf(def.Person{})

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
