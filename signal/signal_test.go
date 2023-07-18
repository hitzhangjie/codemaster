package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

// press Ctrl+C to generate SIGINT
func Test_SignalDispatch(t *testing.T) {
	println(os.Getpid())
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		ch := make(chan os.Signal, 10)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		sig := <-ch
		fmt.Println("B: recv signal:", sig)
	}()

	sig := <-ch
	fmt.Println("A: recv signal:", sig)
	time.Sleep(time.Second)
}
