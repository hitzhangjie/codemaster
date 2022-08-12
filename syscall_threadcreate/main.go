package main

import (
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}

	for {
		go func() {

		}()
		time.Sleep(time.Millisecond * 10)
	}
}
