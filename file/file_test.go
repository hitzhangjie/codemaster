package file

import (
	"fmt"
	"math"
	"net"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestListenerFileFD(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	tcpListener := listener.(*net.TCPListener)

	f, err := tcpListener.File()
	if err != nil {
		t.Fatalf("get file error: %v", err)
	}
	_ = f

	time.Sleep(time.Second)
	for i := 0; i < 20; i++ {
		if i >= 10 {
			tcpListener.Close()
		}
		if i >= 15 {
			f.Close()
		}
		_, err := net.Dial("tcp", "localhost:8888")
		if err != nil {
			t.Logf("%d, connect error: %v", i, err)
		} else {
			t.Logf("%d, connect ok", i)
		}
	}

	time.Sleep(time.Second * 5)

	time.Sleep(time.Minute)
}

func TestNewFile(t *testing.T) {
	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}

	tcpln := ln.(*net.TCPListener)
	f2, _ := tcpln.File() // 2 tcp listenfd now
	_ = f2
	//_ = os.NewFile(f2.Fd(), "")  // 3?
	//_ = os.NewFile(f2.Fd(), "") // 4?

	fmt.Println(os.Getpid())
	time.Sleep(time.Minute)
}

func TestFork(t *testing.T) {
	fmt.Println("parent:", os.Getpid())

	time.Sleep(time.Second)
	v := os.Getenv("restart")
	if v == "1" {
		fmt.Println("child restart now")
	} else {
		if err := syscall.Dup2(0, 10); err != nil {
			panic(err)
		}
		if err := syscall.Dup2(1, 11); err != nil {
			panic(err)
		}
		if err := syscall.Dup2(2, 12); err != nil {
			panic(err)
		}
		os.Setenv("restart", "1")
		pid, err := syscall.ForkExec(os.Args[0], os.Args[1:], &syscall.ProcAttr{
			Env: os.Environ(),
			Files: []uintptr{
				10, 11, 12,
			},
		})
		if err != nil {
			panic(err)
		}
		fmt.Println("child:", pid)
	}
	time.Sleep(time.Minute)
}

func TestNewFile2(t *testing.T) {
	for i := 0; i < 255; i++ {
		f := os.NewFile(uintptr(i), "")
		if f == nil {
			t.Fatal("nil")
		}
	}
	println(os.Getpid())
}

func TestNewFileBug(t *testing.T) {
	ln, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}

	tcpln := ln.(*net.TCPListener)
	file, err := tcpln.File()
	if err != nil {
		panic(err)
	}

	fd := file.Fd()
	fmt.Printf("fd is: %d\n", fd)

	for i := fd; i < math.MaxUint8; i++ {
		file := os.NewFile(i, "")
		if file == nil {
			fmt.Println("ok, file == nil")
			break
		}
		_, err := file.Stat()
		if err != nil {
			fmt.Printf("fd is: %d, stat error: %v\n", i, err)
			continue
		}
	}

	fmt.Println("pid is:", os.Getpid())
	time.Sleep(time.Minute)
}

func TestXXX(t *testing.T) {
	var v interface{}
	v = int(0)
	fmt.Printf("%T\n", v)
}
