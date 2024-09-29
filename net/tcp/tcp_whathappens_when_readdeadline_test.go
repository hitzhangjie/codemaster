package tcp_test

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// 这个测试的是连接deadline超时后能否读取出之前写入的数据，tcpconn.Read(buf) (int, error)会同时返回超时错误、非0的datasize吗？
// 不会！
//
// 读超时连接设置，只是应用程序希望对读取耗时进行控制所进行的一种上层控制操作，不会破坏连接状态，
// 即使读超时了，也可以通过改大deadline或者去掉deadline继续读取出上面的数据。
func Test_WhatHappensWhenReadDeadlineExceeded(t *testing.T) {
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

				// 让服务端read的时候deadline超时，看它返回的(int, error)
				// 是会返回error=deadlineexceeded同时int!=0吗？不会！！！
				c.SetReadDeadline(time.Now().Add(time.Second))
				time.Sleep(time.Second * 5)

				for {
					buf := make([]byte, 4<<10)

					// 这里读取时，tcpconn其实已经deadlineExceeded了，我们关心的是会不会超时之前已经到达的数据能够返回出来，
					// 其实是不能的，这里会直接返回err为超时，并且n==0。为什么呢？
					// 看go源码实现的话，spliceDrain这里面会在读取socket数据时进入pollWait（非阻塞读大概率会走到这里），
					// 这里会检查当前socketfd对应的pollDesc上的状态，当我们设置tcpconn.SetReadDeadline时，其实会给连接更新
					// 一个定时器timer，并且关联一个回调函数，当timer超时时会更新连接的状态为读超时状态。
					// 这样当我们read时会进入到spliceDrain->pollWait->netpollercheckerror这样的一个路径，一看我擦读超时了，
					// 会直接返回错误读超时，管你socket里面有没有数据呢！就直接返回了！
					n, err := c.Read(buf)
					fmt.Printf("server: read tcpconn err: %v, read bytes: %d\n", err, n)

					// 那如果我把连接读deadline去掉再读呢？能读出来吗？是可以的！
					// 读超时，是用户态程序主动控制连接读的一种操作，并不是说连接不可用了，因为希望对读取操作的耗时进行控制而已，
					// 如果把连接deadline搞大或者去掉，是可以正常读取出来的。
					if strings.Contains(err.Error(), "i/o timeout") {
						c.SetDeadline(time.Time{})
						n, err = c.Read(buf)
						fmt.Printf("server: retry without deadline, read tcpconn err: %v, read bytes: %d\n", err, n)
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
	buf := []byte("hello world")
	// 发第一次后睡眠 5s
	n, err := c.Write(buf[:1])
	if err != nil {
		panic(err)
	}
	fmt.Printf("client: send bytes: %d\n", n)

	time.Sleep(time.Second * 5)

	// 发剩下的包，但是此时server端read已经timeout了，看看read时返回的(int, error)都是啥
	buf = buf[1:]
	n, err = c.Write(buf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("client: send bytes: %d\n", n)

	//
	// $ lsof -P -i tcp| grep -i 8000
	// tcp.test  27523 root    6u  IPv6 219574953      0t0  TCP *:8000 (LISTEN)
	// tcp.test  27523 root    7u  IPv6 219574959      0t0  TCP VM-147-116-centos:35372->VM-147-116-centos:8000 (ESTABLISHED)
	// tcp.test  27523 root    8u  IPv6 219574960      0t0  TCP VM-147-116-centos:8000->VM-147-116-centos:35372 (ESTABLISHED)
	time.Sleep(time.Second * 30)
	c.Close()
	fmt.Printf("client: close tcpconn\n")

	// 阻止测试提前退出，快去看下服务端的情况
	time.Sleep(time.Minute * 5)
}
