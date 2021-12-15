package net

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestDial(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:3000")
	if err != nil {
		panic(err)
	}
	addr := conn.LocalAddr()
	time.Sleep(time.Hour)
	fmt.Println(addr.String())
}
