package fd

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestXFD(t *testing.T) {
	fd, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	fmt.Println("pid:", os.Getpid())

	ln := fd.(*net.TCPListener)
	fmt.Println("----")
	time.Sleep(time.Second * 10)

	file, _ := ln.File() // 这里就dup2了

	fmt.Println("----")
	time.Sleep(time.Second * 10)

	fmt.Println("----")
	fmt.Println("fd:", file.Fd())
	time.Sleep(time.Second * 10)

	file.Close()
	fmt.Println("----")
	time.Sleep(time.Second * 10)
}

func TestFD(t *testing.T) {
	ln, _ := net.Listen("tcp", ":8888")
	go func() {
		for {
			_, err := ln.Accept()
			if err != nil {
				return
			}
		}
	}()

	file, err := ln.(*net.TCPListener).File()
	if err != nil {
		panic(err)
	}

	fn := func(f *os.File) {
		runtime.SetFinalizer(file, func(f *os.File) {
			fmt.Println("gc occurred")
		})
	}

	runtime.SetFinalizer(file, fn)
	fd := file.Fd()
	//file.Close() // fd只在file未关闭或gc之前有效

	fmt.Println("fd1:", fd)
	err = syscall.SetNonblock(int(fd), true)
	if err != nil {
		panic(err)
	}

	f := os.NewFile(fd, "")
	fmt.Println("fd2:", f.Fd())

	fileListener, err := net.FileListener(file)
	if err != nil {
		panic(err)
	}
	lll, _ := fileListener.(*net.TCPListener).File() // 通过这里的fd拿到的总是duplicated fd，不是原来哪个，这个是非阻塞的，会影响到原来的
	lll.Fd()
	fmt.Println("new fd", lll.Fd())

	ln.Close()
	fmt.Println("listener closed")

	time.Sleep(time.Second)
	_, err = net.Dial("tcp", "localhost:8888")
	assert.Nil(t, err)

	fmt.Println(os.Getpid())
}
