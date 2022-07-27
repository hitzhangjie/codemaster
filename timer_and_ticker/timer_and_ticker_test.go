package timer_and_ticker_test

import (
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	now := time.Now()
	tm := time.NewTimer(time.Second * 2)
	<-tm.C
	println("ok")
	println(time.Since(now).Seconds())
}

func TestTicker(t *testing.T) {
	tk := time.NewTicker(time.Second)
	for range tk.C {
		println("ok")
	}
}
