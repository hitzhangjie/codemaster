package error

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type myerr struct {
}

func (e *myerr) Error() string {
	return "xxxxxxx"
}

func Test_XXXX(t *testing.T) {
	err := get()
	if err != nil {
		panic(err)
	}
	assert.Nil(t, err)
	assert.IsType(t, err, &myerr{})
}

func get() error {
	var e *myerr
	return e
}
