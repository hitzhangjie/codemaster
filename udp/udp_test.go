package udp_test

import (
	"bytes"
	"log"
	"net"
	"testing"
)

// 哈哈，有同事反映udp不可靠，可能会拆包？？？
//
// 我认为分片是可能会发生的，但是分包至少在规范层面是绝对不应该出现的。
// 如果出现了，那就是实现上的bug，而且这种bug风险比较大。
//
// 不过在实际测试过程中确实复现了，这主要是因为公司的一个虚拟网络适配器导致的，没错就是iOA。
//
// macOS、windows、linux对udp收发包大小限制不一样，比如我写一个udpclient、一个udpserver，udpclient发包给server，包大小是50KB。
// 行为：
//
// - linux->linux（都是devcloud机器），正常发送、接收；
// - windows（办公网）->linux（devcloud机器），发送方看似正常，接收方linux读到了两个udp包，一个32768，一个50KB-32768
// - macOS（办公网）-> linux（devcloud机器），发送方有限制，>=9KB发送失败，<9KB发送成功。
//
// 最后发现是公司iOA的问题

// run on host1 (start the udp server)
func Test_UDP_Server(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp4", ":8888")
	if err != nil {
		panic(err)
	}
	udpconn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 64<<10)
	for {
		n, paddr, err := udpconn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("read %d bytes from %s\n", n, paddr.String())
	}
}

// run on host2 (start the udp client)
func Test_UDP_Client(t *testing.T) {
	conn, err := net.Dial("udp4", "127.0.0.1:8888")
	if err != nil {
		panic(err)
	}
	dat := bytes.Repeat([]byte{'a'}, 50<<10)
	n, err := conn.Write(dat)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("write %d bytes", n)
}
