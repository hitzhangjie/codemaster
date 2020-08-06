package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/libp2p/go-reuseport"
)

const (
	hotRestartFlag = "HOT_RESTART_FLAG"

	network = "tcp"
	address = ":8888"
)

func getListener(network, address string) (net.Listener, error) {

	flag, _ := os.LookupEnv(hotRestartFlag)
	hotRestart, _ := strconv.ParseBool(flag)

	if !hotRestart {
		fmt.Println("server started")
		os.Setenv(hotRestartFlag, "1")

		return reuseport.Listen(network, address)
	}

	// BUG: https://github.com/libp2p/go-reuseport/issues/80
	//
	// on darwin or iOS, reuseport again doesn't guarantee that the two sockets share the same queue,
	// so no balancing work.
	//
	// while, on linux, it works.
	//
	// on darwin, we must use the file descriptor to rebuild the listener.
	// the kernel guarantees that the files passed by &ProcAttr{....,listenerFD} are shared, we can use the listenerFD
	// to rebuild the listener.
	//
	// so, how to get the listenerFD? Is it exactly 3? Yes!
	// When ForkExec runs, it first fork the process in which only the files in &ProcAttr{...} are shared!
	// 0:stdin, 1:stdout, 2:stderr, listenerFD maybe 3 or 10 or other integer, but PID are allocated from the unused
	// minimum pid space, so the listener must be 3!
	fmt.Println("hot restart: server started")
	os.Setenv(hotRestartFlag, "")

	// MUST: use fd=3 to rebuild the listener
	nf := os.NewFile(3, "")
	ln, err := net.FileListener(nf)

	return ln, err
}

func startHttpServer(listener net.Listener) error {

	http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello from %d\n", os.Getpid())
	})

	return http.Serve(listener, nil)
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

	// syscall.ForkExec is not just fork(), it includes exec(), so we cannot use `if pid == 0`  to judge
	// whether current process is parental process or not, use environment instead.
	os.Setenv(hotRestartFlag, "1")

	pid, err := syscall.ForkExec(exec, nil, &syscall.ProcAttr{
		Dir:   wd,
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), listenerFile.Fd()},
	})
	if err != nil {
		fmt.Println("fork error: ", err)
		os.Exit(1)
	}

	fmt.Printf("fork succ, pid: %d\n", pid)
}

func handleSignal(signal os.Signal, listener net.Listener) {
	switch signal {
	case syscall.SIGUSR1, syscall.SIGUSR2:
		forkWithFile(listener)
	case syscall.SIGTERM, syscall.SIGQUIT:
		os.Exit(0)
	}
}

func main() {

	ch := make(chan os.Signal, 16)
	signal.Notify(ch, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGTERM, syscall.SIGTERM, syscall.SIGQUIT)

	ln, err := getListener(network, address)
	if err != nil {
		panic(err)
	}

	go func() {
		err := startHttpServer(ln)
		if err != nil {
			panic(err)
		}
	}()

	for {
		select {
		case s := <-ch:
			fmt.Printf("Recv signal: %v\n", s.String())
			handleSignal(s, ln)
		default:
			time.Sleep(time.Second)
		}
	}
}
