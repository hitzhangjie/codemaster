package icmp

import (
	"fmt"
	"net"
	"testing"

	"golang.org/x/net/icmp"
)

func TestICMP(t *testing.T) {
	netaddr, err := net.ResolveIPAddr("ip4", "127.0.0.1")
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenIP("ip4:icmp", netaddr)
	if err != nil {
		panic(err)
	}

	for {
		buf := make([]byte, 1024)

		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}

		msg, err := icmp.ParseMessage(1, buf[0:n])
		if err != nil {
			panic(err)
		}
		fmt.Println(n, addr, msg.Type, msg.Code, msg.Checksum)
	}
}
