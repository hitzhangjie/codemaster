package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var unixsockAaddr = &net.UnixAddr{
	Name: "/tmp/hot_restart3.sock",
	Net:  "unixgram",
}

func prepareUnixDomainSocket() {
	unixConn, err := net.ListenUnixgram("unixgram", unixsockAaddr)
	if err != nil {
		panic(err)
	}

	for {
		buf := make([]byte, 128, 128)
		n, peer, err := unixConn.ReadFromUnix(buf[0:])
		if err != nil {
			log.Println("read error:", err)
			continue
		}

		s := string(buf[:n])

		if s == phaseRequestFD {
			// write unixConn fd
			fd, name := globalListenerFD()

			// send back fd
			log.Println("fd:", fd, ", filename:", name)
			unixConn.WriteTo([]byte(fmt.Sprintf("%d|%s", fd, name)), peer)
		} else {
			log.Println("read data:", string(buf))
			unixConn.WriteTo(buf, peer)
		}
	}
}

func globalListenerFD() (int, string) {
	ln, ok := globalListener.(*net.TCPListener)
	if !ok {
		panic("invalid unixConn")
	}

	file, err := ln.File()
	if err != nil {
		panic(err)
	}

	fd := int(file.Fd())
	name := file.Name()
	file.Close()

	return fd, name
}

const (
	phaseRequestFD = "request listener fd"
)

var (
	globalListener net.Listener
)

func init() {
	env, _ := os.LookupEnv("RESTART")

	if env == "1" {
		log.SetPrefix("children: ")
	} else {
		log.SetPrefix("parental: ")
	}
}

func main() {

	log.Println(os.Getpid())

	env, _ := os.LookupEnv("RESTART")

	// 执行热重启逻辑
	if env == "1" {

		conn, err := net.DialUnix("unixgram", nil, unixsockAaddr)
		if err != nil {
			panic(err)
		}

		_, err = conn.Write([]byte(phaseRequestFD))
		if err != nil {
			panic(err)
		}
		log.Println("request listener fd success")

		buf := make([]byte, 128, 128)
		n, _, err := conn.ReadFromUnix(buf)
		if err != nil {
			log.Println("read listener fd error:", err)
			return
		}
		log.Println("read listener bytes:", n)

		dat := string(buf[:n])
		vals := strings.Split(dat, ":")
		fd, _ := strconv.Atoi(vals[0])
		name := vals[1]
		log.Printf("children read fd:%d, name:%s", fd, name)

		file := os.NewFile(uintptr(fd), name)
		globalListener, err = net.FileListener(file)
		if err != nil {
			log.Println("children rebuild listenfd error:", err)
		}
	}

	ch := make(chan os.Signal, 16)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	go func() {
		http.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
			fmt.Fprintf(writer, "hello from pid: %d", os.Getpid())
		})

		if globalListener == nil {
			ln, err := net.Listen("tcp", ":8888")
			if err != nil {
				panic(err)
			}
			globalListener = ln
		}

		err := http.Serve(globalListener, nil)
		if err != nil {
			panic(err)
		}
	}()

	for {
		signal := <-ch
		switch signal {
		case syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2:
			go prepareUnixDomainSocket()
			fork(os.Args)
		default:
			log.Println("parent exit, pid:", os.Getpid())
			syscall.Unlink("/tmp/hot_restart3.sock")
			os.Exit(1)
		}
	}
}

func fork(args []string) {

	dir, _ := os.Getwd()
	exec, _ := os.Executable()

	os.Setenv("RESTART", "1")

	fd, _ := globalListenerFD()

	pid, err := syscall.ForkExec(exec, args, &syscall.ProcAttr{
		Dir: dir,
		Env: os.Environ(),
		Files: []uintptr{
			os.Stdin.Fd(),
			os.Stdout.Fd(),
			os.Stderr.Fd(),
			uintptr(fd),
		},
		Sys: nil,
	})
	if err != nil {
		panic(err)
	}

	log.Println("fork succcess, pid:", pid)
}
