package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DecodeAsNumberRatherThanFloat64(t *testing.T) {
	m := map[string]uint64{
		"key": math.MaxUint64,
	}
	dat, err := json.Marshal(m)
	assert.Nil(t, err)

	n := map[string]interface{}{}
	err = json.Unmarshal(dat, &n)
	fmt.Printf("type: %T, value: %v\n", n["key"], n["key"])

	d := json.NewDecoder(bytes.NewBuffer(dat))
	d.UseNumber()
	err = d.Decode(&n)
	fmt.Printf("type: %T, value: %v\n", n["key"], n["key"])
}
