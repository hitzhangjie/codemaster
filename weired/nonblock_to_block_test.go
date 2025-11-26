package main

import (
	"log"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_NonblockToBlock(t *testing.T) {
	// t.Fatal("not implemented")
	addr, err := net.ResolveUDPAddr("udp4", ":8888")
	require.Nil(t, err)

	udpconn, err := net.ListenUDP("udp4", addr)
	require.Nil(t, err)

	ch := make(chan struct{})

	go func() {
		var times int
		for {
			times++
			if err := udpconn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				panic(err)
			}

			<-ch

			buf := make([]byte, 64<<10)
			n, err := udpconn.Read(buf)
			log.Printf("times-%d read data: %d, err: %v", times, n, err)
		}
	}()

	go func() {
		time.Sleep(time.Second * 5)
		f, err := udpconn.File()
		if err != nil {
			panic(err)
		}
		fd := f.Fd()
		log.Printf("fd: %d", fd)

		val, err := syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF)
		if err != nil {
			panic(err)
		}
		log.Printf("SO_RCVBUF: %d", val)

		close(ch)
	}()

	<-ch
	err = udpconn.Close()
	require.Nil(t, err)
}
