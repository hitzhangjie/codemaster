package main

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var addr = ":8888"

func main() {
	log.Printf("pid: %d", os.Getpid())
	if val := os.Getenv("hot_restart"); val != "1" {
		if err := startTCPServer(); err != nil {
			panic(err)
		}
		log.Printf("tcp server started")
	} else {
		log.Printf("recv tcpconn and read")
		recvTCPConnFdAndRead()
	}
	select {}
}

func startTCPServer() error {
	l, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}

	go func() {
		defer l.Close()
		for {
			c, err := l.Accept()
			if err != nil {
				log.Println("accept:", err)
				log.Println("exit")
				return
			}

			go func() {
				for {
					buf := make([]byte, 8)
					n, err := c.Read(buf)
					if err != nil {
						if errors.Is(err, io.EOF) {
							log.Println("conn read eof")
							log.Println("break")
							return
						}
					}

					log.Printf("pid: %d, conn read %d bytes, val: %v", os.Getpid(), n, buf)
				}
			}()

			var ch = make(chan os.Signal, 16)
			signal.Notify(ch, syscall.SIGINT, syscall.SIGUSR2)

			go func() {
				sig := <-ch
				switch sig {
				case syscall.SIGINT:
					log.Printf("recv SIGINT, quit")
					os.Exit(0)
				case syscall.SIGUSR2:
					log.Printf("ready to restart and pass connfd")

					tcpconn := c.(*net.TCPConn)
					f, err := tcpconn.File()
					if err != nil {
						panic(err)
					}
					forkWithFile(f)
					f.Close()
				}
			}()

		}
	}()

	return nil
}

func forkWithFile(f *os.File) {
	env := os.Environ()
	env = append(env, "hot_restart=1")
	pid, err := syscall.ForkExec(os.Args[0], nil, &syscall.ProcAttr{
		Env:   env,
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), f.Fd()},
		Sys:   nil,
	})
	if err != nil {
		panic(err)
	}
	log.Printf("forked pid: %d", pid)
}

func recvTCPConnFdAndRead() {
	f := os.NewFile(uintptr(3), "passed-tcpconn-fd")
	conn, err := net.FileConn(f)
	if err != nil {
		panic(err)
	}
	f.Close()
	c, ok := conn.(*net.TCPConn)
	if !ok {
		panic("not tcpconn")
	}
	for {
		buf := make([]byte, 8)
		n, err := c.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("conn read eof")
				log.Println("break")
				return
			}
		}

		log.Printf("pid: %d, conn read %d bytes, val: %v", os.Getpid(), n, buf)
	}
}
