package tcp_test

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// 这个测试用例用来测试，tcp连接关闭时的一些细节，我们偏偏不走四次挥手的正常流程，来看看可能的一些问题，了解下协议栈是如何处理tcp连接关闭的。
//
// 最开始是想测试，skbuff里的数据还没读，就收到了FIN包，此时skbuff里的数据还能否接着读的问题 …… 能读，只是想测试下系统调用、go标准库返回值；
// 因为用例设计的问题，观察到其他现象：
// - client发送完数据调用了close()，但是还可以正常收到server端发送来的tcp probes，并且还能正常响应？FIN_WAIT2的处理
// - server端读取到eof后没有调用close()完成四次挥手，但是读取完数据后，lsof看不到连接了，从CLOSE_WAIT状态直接变CLOSED状态了？skbuff的处理，RST的处理
// - tcp探活逻辑，我并没有显示设置tcp探活，那为什么会有tcp probes呢？go标准库默认设置问题
// - tcp probes是有哪些参数控制什么时候发送probes，发送间隔是什么，超时时间是什么，连续多少个probes失败才算连接dead？Linux内核参数设置问题
//
// ok, 就这些，都搞清楚了。详细的现象、分析，可以看下面的注释。
//
// ps: 过程中怀疑跟client、server运行在一个进程中、不同进程中、不同机器上对上述问题是否有影响，
// 所以将当前测试用例改成了一个可执行程序接受`-mode=[server|client] -addr=ip:port`，也就是是tcp/tcp_whathappens_when_closing/main.go
// 结论，跟运行在相同进程、不同进程、不同机器没关系，跟网络协议栈对FIN_WAIT2、CLOSE_WAIT、skbuff data、RST这些的处理有关系。多看下协议栈代码吧。
func Test_WhatHappensWhenClosing(t *testing.T) {
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

					// Q1: server端并没有显示调用close，为什么server读取完数据后lsof看不到连接了？
					//
					// 收完包后，内核才会真正关闭连接，但是关闭也要主动关闭才行，需要调用close来完成，或者 …… 在close之前收到了对端发送来的RST
					// $ lsof -P -i tcp| grep -i 8000
					// tcp.test  27523 root    6u  IPv6 219574953      0t0  TCP *:8000 (LISTEN)
					//
					// 从现象上，执行完Read后，lsof命令输出的连接列表就把之前的连接给剔除掉了，why?
					// 但是，服务器并没有完成tcp连接关闭的四次挥手的最后两步，tcp连接状态还处在连接状态还应该处在CLOSE_WAIT状态才对，
					// 正常编码时服务器是需要显示关闭该tcp连接的。我们这里注释掉了`defer func() {c.Close()}`，所以没有按照正常四次
					// 挥手来终止连接。
					//
					// anyway，这里当连接处于 CLOSE_WAIT + skbuff有数据时，lsof是可以显示出该连接的；
					// 但是，当skbuff中数据被读取后，lsof就不显示了，发生了什么呢？
					// 另外，这个连接不是正常关闭的，因为tcpdump抓包，没有看到read完之后去完成四次挥手的最后两步。
					//
					// 实际上我们这个用例client、server是这样交互的：
					// - client发送完数据后调用close，发送了FIN、收到了server ACK，完成了四次挥手的前两步；
					// - 正常接下来是server也主动关闭掉client，但是我们不是正常流程，且听下面分析：
					//   - server此时还会发tcp probes给client，client此时在FIN_WAIT2最多停留30s (sysctl net.ipv4.tcp_fin_timeout=30s)，tcp probe 15s1个
					//   - 在FIN_WAIT2超时前，client可以响应server tcp probes，但是超时后，就直接跳过TIME_WAIT进入CLOSED，client再收到tcp probes就直接回RST
					//   - server收到RST之后，内核层面直接就认为这个连接已经完蛋了，netstat还能看到CLOSE_WAIT状态，仅仅是因为skbuff里还有数据未读走，读完立马进入CLOSED
					//
					// 问题解决！
					//
					// ----------------------------------------------------------------------------------------------
					//
					// Q2: 从这里的tcp probes，说说tcp的fullduplex机制？
					//
					// 有意思的是，尽管clientside已经发送FIN给serverside表示没有数据发送了，serverside在未调用close关闭连接 or 收到RST感知到连接断开之前，
					// serverside还每隔一段时间就发一个空消息体给clientside，这可能是检测clientside连接是否还ok？为什么这么做呢？
					//
					// 因为TCP本身就是fullduplex设计，clientside结束数据发送，serverside还可以继续发送数据给clientside，有些场景下这是有可能的，
					// 所以serverside read eof之后是否关闭连接，应该根据serverside的处理场景来定。比如，如果serverside是一个一发一收的RPC服务端，
					// 那么关闭就没有任何问题，但是如果是一个支持streaming的流式服务，那么关闭就可能有问题，因为serverside完全可能收到1个请求后会发送多个响应回去。
					// 因为client发送FIN，服务器在回完所有数据前就直接关闭，可能导致响应数据不能完全发给客户端。
					//
					// Q3: 好奇，为什么clientside明确调用了close()而非shutdown(write)后，还能收到serverside的tcp包并回应ack，不应该回应rst吗？
					//
					// 1) 先说下探活问题：
					// - 这个serverside探活包实际上是个data长度为0的正常tcp包
					//   ps: 如果A给B发送了FIN包，A还能发探活包给B吗？毕竟它data长度为0？
					//   - TODO 不应该发送，理由是发送了FIN包表示没有更多数据可以发送，不应该继续发送了
					//          ps: 那么回tcp probe探活包算不算呢？看上去响应还是可以发的？探活包是server发过来给client的
					//			    如果client给server发探活包呢，毕竟data==0? 测试下
					//   - TODO 可以发送
					//
					// - 在linux中，可以查看net.ipv4.tcp_keepalive_xxx来查看探活参数设置, see: https://tldp.org/HOWTO/TCP-Keepalive-HOWTO/usingkeepalive.html
					//   简单总结，未开启探活时，linux内核2h后开始发送探活包，每个探活包间隔(探活超时时间是75s)，连续n个探活失败则认为连接dead。
					//   如果显示指定了开启探活以及对应的参数，那么就按业务层指定的来。
					// - 在go中，默认都开启了tcp连接探活，间隔是15s发送一个探活包，这个在tcpdump抓包中可以观察到
					//
					// 2) 再说下clientside调用close关闭连接之后，serverside发起探活请求，为什么clientside还能正常响应，而非直接发送RST？结合tcp state diagram来思考这个问题。
					//
					// - (已验证）调用close时，客户端应该销毁掉这个连接？如果client、server在不同机器上，确实可能如此，那后续server发起探活时就直接回RST了
					//    观察了一下：
					//      > t+0s client发送FIN包后，收到了server的响应
					//      > t+15s server发送了探活包，收到了client的响应
					//      > t+30s server再次发送探活包，收到了client的RST
					//
					//      这个是符合预期的，client主动close()，发送FIN收到ACK后进入FIN_WAIT2，对该阶段的处理操作系统协议栈实现时会引入一些自己的决策。
					//      1) 如果client没收到server的FIN，那么client会在该阶段停留一段时间，此由于客户端没有完全关闭，所以还是可以收server数据；
					//      2) 那client最迟在该阶段停留多久呢？如果服务器崩溃了呢？sysctl net.ipv4.tcp_fin_timeout=30s, Linux为此引入了一个超时时间，
					//         我测试的机器上默认是30s，所以我们观察到了t+30s时，client tcp连接从FIN_WAIT2直接进入了状态ClOSED，就不能响应server探活包了。
					//         ps: 正常转换过程是FIN_WAIT2->TIME_WAIT->CLOSED，但是FIN_WAIT2超时时时直接进入CLOSED状态。
					//
					// - (已验证）但是如果client、server运行在同一台机器上呢？也是这个道理，
					//    - 只不过当时不同机器测试时用的devcloud机器，FIN_WAIT2超时时间是30s；
					//    - 最开始在一台机器上测试时事是用的我的个人电脑wsl，wsl里配置的FIN_WAIT2超时时间是60s，
					//      所以看上去探活包成功的多一点，误导我多一点，实际上也是4次探活成功后后面1次就失败了，上面的分析也能解释通。
					//
					// - （已验证）如果close时，客户端、服务器运行在同一个进程里面？也是一样的道理
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
