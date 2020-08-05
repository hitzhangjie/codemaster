package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/libp2p/go-reuseport"
)

func startHttpServer() net.Listener {
	http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello from %d\n", os.Getpid())
	})

	//l, err := net.Listen("tcp", ":8080")
	l, err := reuseport.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	err = http.Serve(l, nil)
	if err != nil {
		panic(err)
	}

	return l
}

func forkWithFile(listener net.Listener) {

	exec, err := os.Executable()
	if err != nil {
		panic(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var listenerFile *os.File
	if l, ok := listener.(*net.TCPListener); ok && l != nil {
		f, err := l.File()
		if err != nil {
			panic(err)
		}
		listenerFile = f
	}

	files := []*os.File{
		os.Stdin, os.Stdout, os.Stderr, listenerFile,
	}

	p, err := os.StartProcess(exec, nil, &os.ProcAttr{
		Dir:   wd,
		Env:   nil,
		Files: files,
		Sys:   nil,
	})
	if err != nil {
		fmt.Println("fork error: ", err)
		os.Exit(1)
	}

	fmt.Println("fork succ, pid: %d", p.Pid)
}

func handleSignal(signal os.Signal) {
	switch signal {
	case syscall.SIGUSR1, syscall.SIGUSR2:
		forkWithFile(listener)
	case syscall.SIGTERM, syscall.SIGQUIT:
		os.Exit(0)
	}
}

var listener net.Listener

func main() {

	ch := make(chan os.Signal, 16)
	signal.Notify(ch, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGTERM, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		listener = startHttpServer()
	}()

	fmt.Println("server started")

	for {
		select {
		case s := <-ch:
			fmt.Printf("Recv signal: %v\n", s.String())
			handleSignal(s)
		default:
			time.Sleep(time.Second)
		}
	}
}
