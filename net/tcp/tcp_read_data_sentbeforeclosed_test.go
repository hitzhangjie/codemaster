package tcp_test

import (
	"fmt"
	"net"
	"testing"
	"time"
)

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

// 这就需要关心linux kernel什么时候会做tcp连接状态转换：
// 1) 当服务端收到FIN包+响应ACK后，服务端连接状态会进入CLOSE_WAIT，此时服务端还是可以正常收取buffer里面的数据的；
// 2) 读完数据后，再次读取才会返回EOF错误。这就是前面说的内核会在有数据时先返回数据。
//
// 服务端发现EOF后，判定对端不再继续进行请求，可以考虑主动关闭连接，才会进入LAST_ACK->CLOSED转换，这是一般的正常处理流程。
// 当然也可能有其他场景，比如我们精心设计的测试用例 tcp_whathappens_when_closing_test.go中，server没关闭连接也不读数据，
// 而是让tcp probes不断发给client，client处于FIN_WAIT2状态，超时之前它都可以发送tcp probes响应，直到超时发送RST。
// 服务器收到RST后，在读完skbuff数据后，会直接从CLOSE_WAIT状态迁移到CLOSED状态。
//
// see also: tcp_whathappens_when_closing_test.go
func Test_ReadDataSentBeforeClosed(t *testing.T) {
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
	}()

	<-ch

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

	//
	time.Sleep(time.Minute * 5)
}
