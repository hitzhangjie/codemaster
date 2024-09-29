package tcp_test

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// 测试下客户端给服务端发包，发完后立即close连接，服务端过一会再收包，
// 测试这个时候服务端read返回的n, err，
// 读socket和读文件行为是一样的，内核会保证在有数据时返回数据，没数据时返回EOF。
//
//   - 尽管从时序上看，client在server读之前就close了连接（只是一个发起动作而已），
//     直觉上感觉server read的时候应该返回下面之一：
//     1）n != 0 && err == io.EOF，
//     2) n == 0 && err == io.EOF
//
//   - 实际上不是的，server在socket buffer里面有数据时即使收到FIN包也只是相当于数据流
//     末尾插入了一个EOF， 等到read读走数据之后，再次读取才返回EOF.
//
//     这就需要关心linux kernel什么时候会做tcp连接状态转换，当服务端收到FIN包+响应ACK后，
//     服务端连接状态会进入CLOSE_WAIT，此时服务端还是可以正常收取buffer里面的数据的，
//     读完后再次读取，操作系统才会返回EOF错误，这就是前面说的内核会在有数据时先返回数据。
//     服务端发现EOF后，判定对端不再继续进行请求，主动关闭连接后，才会进入LAST_ACK->CLOSED转换。
//
// 接下来，可以看下这个测试用例中的一些注释，根据这个提示去看下netstat中的输出来印证下。
func Test_tcp_close(t *testing.T) {
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

				// 阻止服务端很快收完包，赶紧看下当前tcp链接的状态，CLOSE_WAIT状态，netstat可以看到
				//
				// $ lsof -P -i tcp| grep -i 800
				// tcp.test  27523 root    6u  IPv6 219574953      0t0  TCP *:8000 (LISTEN)
				// tcp.test  27523 root    8u  IPv6 219574960      0t0  TCP VM-147-116-centos:8000->VM-147-116-centos:35372 (CLOSE_WAIT)

				time.Sleep(time.Minute * 2)
				for {
					buf := make([]byte, 4<<10)

					// 收完包后，内核才会真正关闭连接，但是关闭也要主动关闭才行，需要调用close来完成
					//
					// $ lsof -P -i tcp| grep -i 8000
					// tcp.test  27523 root    6u  IPv6 219574953      0t0  TCP *:8000 (LISTEN)
					//
					// 实际上，执行完Read后，lsof命令输出的连接列表就把之前的连接给剔除掉了，
					// 但是，服务器并没有完成tcp连接关闭的四次挥手的最后两步，tcp连接状态还处在连接状态还应该处在CLOSE_WAIT状态才对，
					// 正常编码时服务器是需要显示关闭该tcp连接的。我们这里注释掉了`defer func() {c.Close()}`，所以没有按照正常四次
					// 挥手来终止连接。
					//
					// 有意思的是，尽管clientside已经发送FIN给serverside表示没有数据发送了，serverside在未调用close关闭连接之前，
					// serverside还每隔一段时间就发一个空消息体给clientside，这可能是检测clientside连接是否还ok？为什么这么做呢？
					// 因为TCP本身就是fullduplex设计，clientside结束数据发送，serverside还可以继续发送数据给clientside，这是有
					// 可能的，所以serverside read eof之后是否关闭连接，应该根据serverside的处理场景来定。
					// 比如，如果serverside是一个一发一收的RPC服务端，那么关闭就没有任何问题，但是如果是一个支持streaming的流式服务，
					// 那么关闭就可能有问题，因为serverside完全可能收到1个请求后会发送多个响应回去。
					//
					// anyway，这里当连接处于 CLOSE_WAIT + skbuff有数据时，lsof是可以显示出该连接的；
					// 但是，当skbuff中数据被读取后，lsof就不显示了，这可能跟lsof这个工具的实现有关系。
					// 实际上，这个连接还是没有正常关闭的，因为tcpdump抓包，没有看到read完之后去完成四次挥手的最后两步。
					//
					// 实际上该程序永远没有看到四次挥手的最后两步，真实世界中的client/server交互并不总是按照RFC规范来的。
					// - 服务器按照上述逻辑继续探测clientside该连接是否有效，clientside会回一个空包，而不是回一个RST？奇怪
					// - 程序退出时，操作系统负责关闭进程打开的所有socket，如果clientside收到serverside的tcp包，此时会发个RST通知
					//   对方该连接已被重置。
					//
					// 好奇，为什么clientside明确调用了close()而非shutdown(write)后，还能收到serverside的tcp包并回应ack，不应
					// 该回应rst吗？
					// 1) 先收下探活问题：
					// - 这个serverside探活包实际上是个data长度为0的正常tcp包，
					// - 在linux中，可以查看net.ipv4.tcp_keepalive_xxx来查看探活参数设置, see: https://tldp.org/HOWTO/TCP-Keepalive-HOWTO/usingkeepalive.html
					// - 在go中，默认都开启了tcp连接探活，间隔是15s发送一个探活包，这个在tcpdump抓包中可以观察到
					// 2) 再说下探活时为什么clientside没发送RST包？
					// - 可能跟当前client、server是在一个进程内实现的有关，操作系统可能认为虽然client是希望关掉了，但是server还可能用，
					// 所以它实际上就没有真正的关闭这个tcpconn，所以serverside发起的探活包clientside还可以收到后响应。注意探活包包体
					// 是空的，并不会像其他数据包一样被上层读取到，所以即使之前发送了FIN也可以正常发探活包（tcp本身就是全双工的，其中一个
					// 方向不发送了，但是另一个反向还希望能发送，那就需要检测连接活性，响应空包就可以实现这个目的）。
					// - 如果client、server是啊不同进程上实现的，结果就应该是不同的了。
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
	buf := []byte("hello world")
	n, err := c.Write(buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("client: send bytes: %d\n", n)

	// 这个时候看下，netstat应该能看到tcp链接信息：
	//
	// $ lsof -P -i tcp| grep -i 8000
	// tcp.test  27523 root    6u  IPv6 219574953      0t0  TCP *:8000 (LISTEN)
	// tcp.test  27523 root    7u  IPv6 219574959      0t0  TCP VM-147-116-centos:35372->VM-147-116-centos:8000 (ESTABLISHED)
	// tcp.test  27523 root    8u  IPv6 219574960      0t0  TCP VM-147-116-centos:8000->VM-147-116-centos:35372 (ESTABLISHED)
	time.Sleep(time.Second * 30)
	c.Close()
	fmt.Printf("client: close tcpconn\n")

	// 阻止测试提前退出，快去看下服务端的情况
	//
	// 服务端tcpconn读取返回EOF后，并没有主动调用close(conn)来关闭连接，那么这个连接在操作系统视角是否还存在呢？
	// lsof -P -i tcp| grep -i 8000
	// tcp.test 79580 zhangjie    6u  IPv4 4934355      0t0  TCP *:8000 (LISTEN)
	//
	// lsof输出是看不到了，但是不代表四次握手正常结束了，很可能只是lsof把CLOSE_WAIT并且skbuff没数据的给剔除了。
	// 看看前面服务器read失败后写的分析。
	time.Sleep(time.Minute * 5)
}
