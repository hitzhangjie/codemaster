package main

import (
	"log"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// 这是之前遇到的一个奇怪的bug的复现用例：
// - udpconn工作在非阻塞模式，一直在forloop收取udp请求；
// - 有一个操作尝试获取udpconn.File().Fd()，然后通过这个fd获取udp socket recvbuf大小
// - 上述udpconn.File()底层是是通过dup新建了一个指向socket的fd，然后包装了一个新的os.File
// - 然后尝试获取file.Fd()，也就是dup出来的fd，
// - 再然后通过该fd获取对应的socket的recvbuf大小，fd虽然不同，但是指向的是同一个socket
//
// 注意：这里的file.Fd()操作历史上会将返回的fd设置为阻塞模式的，此时就会影响到原来的udpconn,
// udpconn以后将工作在阻塞模式，除非有后续操作将其设置为非阻塞。
//
// 执行这个测试用例最后你将看到尝试Close的goroutine一直在等锁:
/*
goroutine 6 [semacquire]:
internal/poll.runtime_Semacquire(0x7f9e449ea8a0?)
	/usr/local/go/src/runtime/sema.go:67 +0x25
internal/poll.(*FD).Close(0xc0000fc380)
	/usr/local/go/src/internal/poll/fd_unix.go:113 +0x65
net.(*netFD).Close(0xc0000fc380)
	/usr/local/go/src/net/fd_posix.go:37 +0x32
net.(*conn).Close(0xc000098090)
	/usr/local/go/src/net/net.go:203 +0x36
github.com/hitzhangjie/codemaster/weired.Test_NonblockToBlock(0xc0000f49c0)
	/root/github/codemaster/weired/nonblock_to_block_test.go:90 +0x171
testing.tRunner(0xc0000f49c0, 0x5f52c0)
	/usr/local/go/src/testing/testing.go:1689 +0xfb
created by testing.(*T).Run in goroutine 1
	/usr/local/go/src/testing/testing.go:1742 +0x390
*/
func Test_NonblockToBlock(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp4", ":8888")
	require.Nil(t, err)

	udpconn, err := net.ListenUDP("udp4", addr)
	require.Nil(t, err)

	ch := make(chan struct{})

	// before another goroutine calls file.Fd(), this goroutine keeps reading
	// from udpconn, then timeout (for no incoming requests), then again and again.
	//
	// until another goroutine calls file.Fd(), then the udpconn is set to blocking mode.
	// then it will block on the read operation.
	//
	// so no more "times-%d read data: ...." output.
	go func() {
		var times int
		for {
			times++
			if err := udpconn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				panic(err)
			}
			buf := make([]byte, 64<<10)
			n, err := udpconn.Read(buf)
			log.Printf("times-%d read data: %d, err: %v", times, n, err)
		}
	}()

	// This goroutine calls file.Fd(), then the udpconn is set to blocking mode.
	// It first waits for 5 seconds, to let the read goroutine keep reading from
	// udpconn and print reading timeout logs.
	//
	// Then it calls file.Fd(), then the udpconn is set to blocking mode.
	// This will block the read goroutine, so no more "times-%d read data: ...." output.
	go func() {
		time.Sleep(time.Second * 5)

		f, err := udpconn.File()
		if err != nil {
			panic(err)
		}
		fd := f.Fd()
		log.Printf("fd: %d", fd)

		val, err := syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF)
		if err != nil {
			panic(err)
		}
		log.Printf("SO_RCVBUF: %d", val)

		close(ch)
	}()

	// OK, let the read goroutine keep reading from udpconn and print reading
	// timeout logs, then another goroutine calls file.Fd(), then the udpconn
	// is set to blocking mode.
	//
	// After that, we sleep 5s first, we want udpconn.Read happens first, so
	// udpconn.Read will block and never return, then udpconn.netFd.fdmu is
	// locked.
	//
	// Then udpconn.Close() will be serialized and keep waiting for the lock.
	// The weired bug happens here, udpconn.Close() blocked forever.
	<-ch
	time.Sleep(time.Second * 5)
	err = udpconn.Close()
	require.Nil(t, err)
}
