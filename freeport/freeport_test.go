package freeport

import (
	"net"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestGetFreePort(t *testing.T) {
	GetPort()
}

// GetFreePort asks the kernel for a free open port that is ready to use.
//
// FIXME：能预留这个端口号吗，有可能会有bug，比如刚l.Close()之后就被其他人使用了，这个端口号就有问题了
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	spew.Printf("addr: %+v\n", addr)

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

// GetPort is deprecated, use GetFreePort instead
// Ask the kernel for a free open port that is ready to use
func GetPort() int {
	port, err := GetFreePort()
	if err != nil {
		panic(err)
	}
	return port
}
