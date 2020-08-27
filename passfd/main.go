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
)

const envRestart = "RESTART"
const envListenFD = "LISTENFD"

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
			Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), f.Fd()},
			Sys:   nil,
		})
		if err != nil {
			panic(err)
		}
		log.Print("parent pid:", os.Getpid(), ", pass fd:", f.Fd())
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
		ff := os.NewFile(uintptr(3), "")
		fmt.Println("fd:", ff.Fd())
		if ff != nil {
			_, err := ff.Stat()
			if err != nil {
				panic(err)
			}

			// pause, ctrl+d to continue
			ioutil.ReadAll(os.Stdin)
			fmt.Println("....")
			_, err = net.FileListener(ff)
			if err != nil {
				panic(err)
			}
			ff.Close()
			// lsof -P -p $pid, 会发现有两个listenfd
			time.Sleep(time.Minute)
		}

		time.Sleep(time.Minute)
	}
}
