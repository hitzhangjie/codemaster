package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	passfd "github.com/ftrvxmtrx/fd"
)

const envRestart = "RESTART"
const envListenFD = "LISTENFD"
const unixsockname = "/tmp/xxxxxxxxxxxxxxxxx.sock"

func main() {

	v := os.Getenv(envRestart)

	if v != "1" {

		ln, err := net.Listen("tcp", "localhost:8888")
		if err != nil {
			panic(err)
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				ln.Accept()
			}
		}()

		tcpln := ln.(*net.TCPListener)
		f, err := tcpln.File()
		if err != nil {
			panic(err)
		}

		os.Setenv(envRestart, "1")
		os.Setenv(envListenFD, fmt.Sprintf("%d", f.Fd()))

		_, err = syscall.ForkExec(os.Args[0], os.Args, &syscall.ProcAttr{
			Env:   os.Environ(),
			Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), /*f.Fd()*/}, // comment this when test unixsock
			Sys:   nil,
		})
		if err != nil {
			panic(err)
		}
		log.Print("parent pid:", os.Getpid(), ", pass fd:", f.Fd())

		os.Remove(unixsockname)
		unix, err := net.Listen("unix", unixsockname)
		if err != nil {
			panic(err)
		}
		unixconn, err := unix.Accept()
		if err != nil {
			panic(err)
		}
		err = passfd.Put(unixconn.(*net.UnixConn), f)
		if err != nil {
			panic(err)
		}

		f.Close()
		wg.Wait()

	} else {

		v := os.Getenv(envListenFD)
		fd, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			panic(err)
		}
		log.Print("child pid:", os.Getpid(), ", recv fd:", fd)

		// test1

		// 通过环境变量肯定是不行的，fd根本不对应子进程中的fd
		//ff := os.NewFile(uintptr(fd), "")
		//if ff != nil {
		//	_, err := ff.Stat()
		//	if err != nil {
		//		log.Println(err)
		//	}
		//}

		// test2

		// 假定我们知道fd是多少，比如fd=3
		//ff := os.NewFile(uintptr(3), "")
		//if ff != nil {
		//	_, err := ff.Stat()
		//	if err != nil {
		//		panic(err)
		//	}
		//
		//	// pause, ctrl+d to continue
		//	ioutil.ReadAll(os.Stdin)
		//	fmt.Println("....")
		//	_, err = net.FileListener(ff) //会dup一个fd出来，有多个listener
		//	if err != nil {
		//		panic(err)
		//	}
		//	// lsof -P -p $pid, 会发现有两个listenfd
		//	time.Sleep(time.Minute)
		//}

		// test 3

		ioutil.ReadAll(os.Stdin)
		fmt.Println(".....") // lsof -P -p $pid，检查下

		unixconn, err := net.Dial("unix", unixsockname)
		if err != nil {
			panic(err)
		}

		files, err := passfd.Get(unixconn.(*net.UnixConn), 1, nil)
		if err != nil {
			panic(err)
		}

		// lsof -P -p $pid再检查下

		f := files[0]
		f.Stat()

		time.Sleep(time.Minute)
	}
}
