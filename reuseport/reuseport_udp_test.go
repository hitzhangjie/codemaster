package reuseport

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/libp2p/go-reuseport"
)

// 测试下Linux UDP reuseport distribute datagrams evenly
//
// 思考：如果服务端创建的udpconn，如果没有线程在epoll read，那么内核为什么还要将datagrams转发给这个udpconn对应的socket呢？
//
//	这不就导致bug了？尽管是业务层自己也有bug，但是操作系统不能规避这个嘛
func TestReuseport_UDP_DistributeDatagramsEvenly(t *testing.T) {
	for i := 0; i < runtime.NumCPU(); i++ {
		ln, err := reuseport.ListenPacket("udp4", ":8888")
		if err != nil {
			panic(err)
		}
		udpconn := ln.(*net.UDPConn)
		// f, _ := udpconn.File()
		// fd := f.Fd()
		// fmt.Printf("i-%d file: %p, listenfd:%d\n", i, f, fd)

		pkg := make([]byte, 64)
		go func(i int, udpconn *net.UDPConn) {
			// 1. 检测下Linux负载均衡怎么做的，如果我注释掉，那么各个udpconn会均匀收包，
			// 2. 如果我uncomment这段代码，意味着只有一个在收包，内核怎么做负载均衡呢？还是每个都发，但是(NUMCPU-1)/NUMCPU实际上都没有被正常处理
			if i != 0 {
				return
			}
			for {
				n, err := udpconn.Read(pkg)
				if err != nil {
					if err == io.EOF {
						return
					}
					fmt.Printf("goroutine-%v, read bytes:%d, err:%v\n", i, n, err)
					continue
				}
				fmt.Printf("goroutine-%v, read bytes:%d\n", i, n)
			}
		}(i, udpconn)
	}

	time.Sleep(time.Second)
	dat := []byte("helloworld")
	for i := 0; i < 10; i++ {
		udpconn, err := net.Dial("udp4", "127.0.0.1:8888")
		if err != nil {
			fmt.Printf("connect error: %v\n", err)
			break
		}

		n, err := udpconn.Write(dat)
		if err != nil {
			fmt.Printf("write bytes:%d, err:%v\n", n, err)
			break
		}
		fmt.Printf("write bytes: %d\n", len(dat))
	}

	time.Sleep(time.Second * 10)
}
