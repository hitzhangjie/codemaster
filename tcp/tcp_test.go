package tcp

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"
)

// Test_Write_Without_Read client一直发包，但是server一直不收包，测试server端拥塞窗口降为0，
// client这边拥塞控制机制不会继续发送请求， 这里后面client这边的tcpconn.Write会阻塞。
//
// ps: 拥塞窗口大小变化可以通过tcpdump抓包来查看，每次回包都会向对端通告自己的窗口大小。
func Test_Write_Without_Read(t *testing.T) {
	ch := make(chan int, 1)

	go func() {
		ln, err := net.Listen("tcp", ":8000")
		if err != nil {
			panic(err)
		}
		defer ln.Close()
		ch <- 1

		for {
			c, err := ln.Accept()
			if err != nil {
				fmt.Printf("server accept tcoconn, error: %v\n", err)
				return
			}
			fmt.Printf("server tcpconn established, peer: %s\n", c.RemoteAddr().String())
		}
	}()

	<-ch

	c, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		panic(err)
	}
	for {
		buf := bytes.Repeat([]byte{1}, 64*1024)
		n, err := c.Write(buf)
		if err != nil {
			fmt.Printf("client send err: %v\n", err)
		}
		if n != len(buf) {
			fmt.Printf("client send bytes mismatch, expected %d, got %d\n", len(buf), n)
		} else {
			fmt.Printf("client send bytes %d\n", n)
		}
		time.Sleep(time.Second)
	}

}
