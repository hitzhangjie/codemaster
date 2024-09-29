package tcp_test

import (
	"log"
	"net"
	"os"
	"testing"
	"time"
)

// readdeadline控制的是什么呢？
//
// 可以理解成如果无数据到达时要等多久，如果有数据到达，也会立即返回的，不一定填满buffer或者等够readdeadline的时间。
// 其实联想下epoll的工作原理，这些就都很好理解。其实可以简单理解为，read notready时，最大等多长时间
//
// see also: `int epoll_wait(int epfd, struct epoll_event *events, int maxevents, int timeout)`
func Test_ReadDeadline(t *testing.T) {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("accept incoming conn fail: %v", err)
				continue
			}
			go func() {
				buf := make([]byte, 1024)
				for {
					// 如果有数据到达，会立即返回的，不一定填满buffer或者等够readdeadline的时间。
					// 其实可以简单理解为，read notready时，最大等多长时间
					_ = conn.SetReadDeadline(time.Now().Add(time.Second * 30))
					n, err := conn.Read(buf)
					if err != nil {
						if os.IsTimeout(err) {
							continue
						}
					}
					log.Printf("read data: %s", string(buf[:n]))
				}
			}()
		}
	}()

	time.Sleep(time.Second)

	conn, err := net.Dial("tcp4", ln.Addr().String())
	if err != nil {
		panic(err)
	}

	s := []byte("helloworld")
	time.Sleep(time.Second)
	conn.Write(s[0:5])

	time.Sleep(time.Second * 15)
	conn.Write(s[5:])

	time.Sleep(time.Second)
}
