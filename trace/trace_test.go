package trace_test

import (
	"context"
	"runtime/trace"
	"sync"
	"testing"
	"time"
)

func Test_trace(t *testing.T) {
	ctx, task := trace.NewTask(context.TODO(), "one day")
	defer task.End()

	getup := trace.StartRegion(ctx, "get up")
	time.Sleep(time.Microsecond * 10)
	getup.End()

	wg := sync.WaitGroup{}
	ch := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := trace.StartRegion(ctx, "washing")
		time.Sleep(time.Microsecond * 10)
		r.End()

		r = trace.StartRegion(ctx, "meat")
		time.Sleep(time.Microsecond * 10)
		r.End()

		r = trace.StartRegion(ctx, "wc")
		time.Sleep(time.Microsecond * 10)
		r.End()
		ch <- 1
		ch <- 1
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ch
		r := trace.StartRegion(ctx, "music")
		time.Sleep(time.Microsecond * 20)
		r.End()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ch
		r := trace.StartRegion(ctx, "subway")
		time.Sleep(time.Microsecond * 10)
		r.End()

		r = trace.StartRegion(ctx, "bike")
		time.Sleep(time.Microsecond * 10)
		r.End()
	}()
	wg.Wait()

	r := trace.StartRegion(ctx, "meeting")
	time.Sleep(time.Microsecond * 20)
	r.End()

	r = trace.StartRegion(ctx, "programming")
	time.Sleep(time.Microsecond * 10)
	r.End()

	trace.Logf(ctx, "msg", "this is my oneday")
}
