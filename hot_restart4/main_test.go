package main

import (
	"syscall"
	"testing"
)

func TestNewUnixDomainSocket(t *testing.T) {
	addr := "/tmp/helloworld.sock"
	c, err := NewUnixDomainSocket(addr)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Unlink(addr)
	defer c.Close()
	t.Log("success")
}

func TestSendToDomainSocket(t *testing.T) {

	syscall.Unlink("xxxx.sock")

	conn, err := NewUnixDomainSocket("xxxx.sock")
	if err != nil {
		t.Fatal(err)
	}

	addr := ResolveUnixAddr("xxxx.sock")

	for i := 0; i < 10; i++ {

		n, err := SendToDomainSocket(conn, []byte("helloworld"), addr)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("send ok, bytes:%d", n)

		buf := make([]byte, 128, 128)
		n, peer, err := ReadFromDomainSocket(conn, buf)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("read ok, bytes:%d, peer:%v", n, peer)
	}
}
