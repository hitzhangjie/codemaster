package error

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type myerr struct {
}

func (e *myerr) Error() string {
	return "this is an error"
}

func Test_CheckNilError(t *testing.T) {
	err := doSomething()
	// 动态类型不为nil，err就不为nil（甭管动态值是否为nil）
	if err != nil {
		assert.IsType(t, err, &myerr{})
	}
	// 注意，动态值确实为nil
	assert.Nil(t, err)
}

func doSomething() error {
	var e *myerr
	return e
}
