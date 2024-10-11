package tcp_test

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// 这个测试用例用来测试下tcpconn client主动关闭后，client还能否给对端发送tcpprobes，测试是不能。
func Test_SendProbes_after_ShutdownWrite(t *testing.T) {
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
				fmt.Printf("server: accept tcoconn, error: %v\n", err)
				return
			}
			fmt.Printf("server： tcpconn established, peer: %s\n", c.RemoteAddr().String())

			go func() {

				// 正常编码时需要显示关闭该连接，尽快完成四次挥手
				//
				// defer func() {
				// 	_ = c.Close()
				// }()

				time.Sleep(time.Minute * 2)
				for {
					buf := make([]byte, 4<<10)
					n, err := c.Read(buf)
					if err != nil {
						fmt.Printf("server: read tcpconn err: %v, read bytes: %d\n", err, n)
					}
					if n == 0 {
						return
					}
					fmt.Printf("server: read data: %s\n", string(buf[:n]))
				}
			}()
		}
	}()

	<-ch

	c, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		panic(err)
	}

	// client设置每隔1s发送一个tcp探活包
	c.(*net.TCPConn).SetKeepAlive(true)
	c.(*net.TCPConn).SetKeepAlivePeriod(time.Second)

	buf := []byte("hello world")
	n, err := c.Write(buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("client: send bytes: %d\n", n)

	// 1) sleep 30s这段时间，client端应该会每隔1s发送一个探活包：验证确实如此！
	// $ sudo tcpdump -i any tcp and dst port 8000 and tcp[tcpflags]=tcp-ack -vvv -nn
	// tcpdump: listening on any, link-type LINUX_SLL (Linux cooked v1), capture size 262144 bytes
	// 11:29:14.398522 IP (tos 0x0, ttl 64, id 51094, offset 0, flags [DF], proto TCP (6), length 52)
	//     127.0.0.1.44750 > 127.0.0.1.8000: Flags [.], cksum 0xfe28 (incorrect -> 0x15c9), seq 1416374172, ack 858335693, win 512, options [nop,nop,TS val 2844832609 ecr 2844832609], length 0
	// 11:29:15.444614 IP (tos 0x0, ttl 64, id 51096, offset 0, flags [DF], proto TCP (6), length 52)
	//     127.0.0.1.44750 > 127.0.0.1.8000: Flags [.], cksum 0xfe28 (incorrect -> 0x11a7), seq 10, ack 1, win 512, options [nop,nop,TS val 2844833656 ecr 2844832610], length 0
	// 11:29:16.484766 IP (tos 0x0, ttl 64, id 51097, offset 0, flags [DF], proto TCP (6), length 52)
	//     127.0.0.1.44750 > 127.0.0.1.8000: Flags [.], cksum 0xfe28 (incorrect -> 0x0981), seq 10, ack 1, win 512, options [nop,nop,TS val 2844834696 ecr 2844833656], length 0
	//
	// 2) 每隔15s应该能收到1个来自server端的探活包，server端默认15s一个探活包：验证确实如此！
	// $ sudo tcpdump -i any tcp and src port 8000 and tcp[tcpflags]=tcp-ack -vvv -nn
	// tcpdump: listening on any, link-type LINUX_SLL (Linux cooked v1), capture size 262144 bytes
	// 11:27:35.924615 IP (tos 0x0, ttl 64, id 5838, offset 0, flags [DF], proto TCP (6), length 52)
	//     127.0.0.1.8000 > 127.0.0.1.37908: Flags [.], cksum 0xfe28 (incorrect -> 0x61a3), seq 234932422, ack 1456125858, win 512, options [nop,nop,TS val 2844734136 ecr 2844718776], length 0
	// 11:27:51.284670 IP (tos 0x0, ttl 64, id 5839, offset 0, flags [DF], proto TCP (6), length 52)
	//     127.0.0.1.8000 > 127.0.0.1.37908: Flags [.], cksum 0xfe28 (incorrect -> 0xe9a2), seq 0, ack 1, win 512, options [nop,nop,TS val 2844749496 ecr 2844734136], length 0

	time.Sleep(time.Second * 30)
	c.Close()
	fmt.Printf("client: close tcpconn\n")

	// 在这之前继续观察下client端是否还会给server端发送tcp探活包? 测试表明，client
	// 主动关闭后，在最终四次挥手结束前，client也不会再继续发送探活包了。
	time.Sleep(time.Minute * 5)
}
