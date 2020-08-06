package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hitzhangjie/codemaster/hot_restart2/logger"
	"github.com/hitzhangjie/codemaster/hot_restart2/server"
)

func main() {
	log := logger.New("Server")

	var s *server.Server
	var err error
	if os.Getenv("_GRACEFUL_RESTART") == "true" {
		s, err = server.NewFromFD(log, 3) // 3, is the listenerfd passed in &ProcAttr{...}
	} else {
		s, err = server.New(log, 12345)
	}
	if err != nil {
		log.Fatalln("fail to init server:", err)
	}
	log.Println("Listen on", s.Addr())

	go s.StartAcceptLoop()

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGTERM)

	for sig := range signals {
		if sig == syscall.SIGTERM {
			// Stop accepting new connections
			s.Stop()
			// Wait a maximum of 10 seconds for existing connections to finish
			err := s.WaitWithTimeout(10 * time.Second)
			if err == server.WaitTimeoutError {
				log.Printf("Timeout when stopping server, %d active connections will be cut.\n", s.ConnectionsCounter())
				os.Exit(-127)
			}
			// Then the program exists
			log.Println("Server shutdown successful")
			os.Exit(0)
		} else if sig == syscall.SIGHUP {
			// Stop accepting requests
			s.Stop()
			// Get socket file descriptor to pass it to fork
			listenerFD, err := s.ListenerFD()
			if err != nil {
				log.Fatalln("Fail to get socket file descriptor:", err)
			}
			// Set a flag for the new process start process
			os.Setenv("_GRACEFUL_RESTART", "true")
			execSpec := &syscall.ProcAttr{
				Env:   os.Environ(),
				Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), listenerFD},
			}
			// Fork exec the new version of your server
			fork, err := syscall.ForkExec(os.Args[0], os.Args, execSpec)
			if err != nil {
				log.Fatalln("Fail to fork", err)
			}
			log.Println("SIGHUP received: fork-exec to", fork)
			// Wait for all conections to be finished
			s.Wait()
			log.Println(os.Getpid(), "Server gracefully shutdown")

			// Stop the old server, all the connections have been closed and the new one is running
			os.Exit(0)
		}
	}
}
