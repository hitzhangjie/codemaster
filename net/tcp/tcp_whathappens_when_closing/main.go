package main

import (
	"flag"
	"fmt"
	"net"
	"time"
)

var mode = flag.String("mode", "server", "run in server or client mode")

// 测试下客户端给服务端发包，发完后立即close连接，服务端过一会再收包，
// 测试这个时候服务端read返回的 `n, err = conn.Read(buf)`，
//
// 读socket和读文件行为是一样的，内核会保证在有数据时返回数据，没数据时返回EOF。
//
//   - 尽管从时序上看，client在server读之前就close了连接（只是一个发起动作而已），
//     直觉上感觉server read的时候应该返回下面之一：
//     1）n != 0 && err == io.EOF，
//     2) n == 0 && err == io.EOF
//
//   - 实际上不是的，server在socket buffer里面有数据时即使收到FIN包也只是相当于数据流
//     末尾插入了一个EOF， 等到read读走数据之后，再次读取才返回EOF.
func main() {
	flag.Parse()

	ch := make(chan int, 1)

	switch *mode {
	case "server":
		startServer()
	case "client":
		startClient()
	default:
		panic("invalid mode")
	}

}

func startServer() {
	ln, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

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

			// 这里for循环实际执行了2次，第一次是正常读数据，n>0 && err==nil；第二次读取是读取到的EOF错误
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
}

func tartClient() {
	// clientside dial and send data to server
	c, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		panic(err)
	}
	buf := []byte("hello world")
	n, err := c.Write(buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("client: send bytes: %d\n", n)

	// close the conn nearly at clocktime +30s, then server wait +2min,
	// then server read the data sent before and after the close event.
	//
	// here, we guarantee that server read happens after the client close event.
	time.Sleep(time.Second * 30)
	c.Close()
	fmt.Printf("client: close tcpconn\n")

	time.Sleep(time.Minute * 5)
}
