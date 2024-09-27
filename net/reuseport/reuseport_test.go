package reuseport_test

import (
	"fmt"
	"net"
	"runtime"
	"testing"

	"github.com/libp2p/go-reuseport"
)

func Test_Reuseport(t *testing.T) {
	for i := 0; i < runtime.NumCPU(); i++ {
		ln, err := reuseport.Listen("tcp", ":8888")
		if err != nil {
			panic(err)
		}
		tcpln := ln.(*net.TCPListener)
		f, _ := tcpln.File()
		fd := f.Fd()
		fmt.Println("listenfd:", fd)
	}
}
