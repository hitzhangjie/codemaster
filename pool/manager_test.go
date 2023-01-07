package pool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var serverAddr = getFreeAddr("tcp")

func TestMain(m *testing.M) {
	// setup tcp server
	l, err := net.Listen("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	fmt.Println("server listening:", serverAddr)

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				fmt.Println("accept err:", err)
			}
			buf := make([]byte, 1024)
			go func() {
				defer c.Close()
				for {
					_, err := c.Read(buf)
					if err == io.EOF {
						return
					}
				}
			}()
		}
	}()

	// run testcases
	os.Exit(m.Run())
}

func TestManager(t *testing.T) {
	// 最多1个链接，池子中允许预备最少0个链接
	t.Run("idle=0, max=1", func(t *testing.T) {
		pm := New(0, 0, 1, defaultCheckInterval, defaultIdleDuration)
		assert.NotNil(t, pm)

		conn, err := pm.Get(context.TODO(), "tcp", serverAddr)
		//nolint:staticcheck
		defer conn.Close()

		assert.Nil(t, err)
		assert.NotNil(t, conn)
	})

	t.Run("idle=0, max=1", func(t *testing.T) {
		pm := New(0, 0, 1, defaultCheckInterval, defaultIdleDuration)
		assert.NotNil(t, pm)

		c1, err := pm.Get(context.TODO(), "tcp", serverAddr)
		assert.Nil(t, err)
		assert.NotNil(t, c1)

		c2, err := pm.Get(context.TODO(), "tcp", serverAddr)
		assert.NotNil(t, err)
		assert.Equal(t, ErrConnTooMany, err)
		assert.Nil(t, c2)
	})
}

func BenchmarkXxx(b *testing.B) {
	pm := New(1, 1, 4, defaultCheckInterval, defaultIdleDuration)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn, err := pm.Get(context.TODO(), "tcp", serverAddr)
		if err != nil {
			b.Fatalf("get conn fail: %v", err)
		}
		conn.Close()
	}
}

func getFreeAddr(network string) string {
	p, err := getFreePort(network)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(":%d", p)
}

func getFreePort(network string) (int, error) {
	if network == "tcp" || network == "tcp4" || network == "tcp6" {
		addr, err := net.ResolveTCPAddr(network, "localhost:0")
		if err != nil {
			return -1, err
		}
		l, err := net.ListenTCP(network, addr)
		if err != nil {
			return -1, err
		}
		defer l.Close()
		return l.Addr().(*net.TCPAddr).Port, nil
	}
	if network == "udp" || network == "udp4" || network == "udp6" {
		addr, err := net.ResolveUDPAddr(network, "localhost:0")
		if err != nil {
			return -1, err
		}
		l, err := net.ListenUDP(network, addr)
		if err != nil {
			return -1, err
		}
		defer l.Close()
		return l.LocalAddr().(*net.UDPAddr).Port, nil
	}
	return -1, errors.New("invalid network")
}
