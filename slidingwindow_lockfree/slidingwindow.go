package slidingwindow

import (
	"sync/atomic"
	"time"

	lockfree "github.com/hitzhangjie/codemaster/queue"
)

// SlidingWindow sliding window consists two windows `curr` and `prev`,
// the window is advanced when recording events.
type SlidingWindow struct {
	size  time.Duration
	count int64
	last  time.Time

	curr *window
	prev *window

	q lockfree.IQueue[event]
}

// NewSlidingWindow creates a new slidingwindow
func NewSlidingWindow(size time.Duration) *SlidingWindow {
	currWin := newLocalWindow()

	// The previous window is static (i.e. no add changes will happen within it),
	// so we always create it as an instance of window.
	//
	// In this way, the whole limiter, despite containing two windows, now only
	// consumes at most one goroutine for the possible sync behaviour within
	// the current window.
	prevWin := newLocalWindow()

	sw := &SlidingWindow{
		size:  size,
		count: 0,
		last:  time.Now(),
		curr:  currWin,
		prev:  prevWin,
		q:     lockfree.NewLockfreeQueue[event](),
	}
	go sw.start()

	return sw
}

type event struct {
	t time.Time
	v int64
}

func (sw *SlidingWindow) start() {
	for {
		event, ok := sw.q.Dequeue()
		if !ok {
			continue
		}
		sw.record(event.t, event.v)
	}
}

// Size returns the time duration of one window size. Note that the size
// is defined to be read-only, if you need to change the size,
// create a new limiter with a new size instead.
func (sw *SlidingWindow) Size() time.Duration {
	return sw.size
}

// Allow is shorthand for AllowN(time.Now(), 1).
func (sw *SlidingWindow) Record() {
	evt := event{
		t: time.Now(),
		v: 1,
	}
	sw.q.Enqueue(evt)
}

// AllowN reports whether n events may happen at time now.
func (sw *SlidingWindow) record(now time.Time, n int64) {
	sw.advance(now)
	sw.curr.AddCount(n)

	// here we uses `now` instead of time.Now()
	elapsed := now.Sub(sw.curr.Start())
	weight := float64(sw.size-elapsed) / float64(sw.size)
	count := int64(weight*float64(sw.prev.Count())) + sw.curr.Count()
	atomic.StoreInt64(&sw.count, count)

	//sync btw cache and ram may cause cache bouncing, so we sync later
	//if now.Sub(sw.last) > sw.size {
	//	atomic.StoreInt64(&sw.count, sw.count)
	//	sw.last = now
	//}
}

func (sw *SlidingWindow) Count() int64 {
	return atomic.LoadInt64(&sw.count)
}

// advance updates the current/previous windows resulting from the passage of time.
func (sw *SlidingWindow) advance(now time.Time) {
	// Calculate the start boundary of the expected current-window.
	newCurrStart := now.Truncate(sw.size)

	diffSize := newCurrStart.Sub(sw.curr.Start()) / sw.size

	// Fast path, the same window
	if diffSize == 0 {
		return
	}

	// Slow path, the current-window is at least one-window-size behind the expected one.

	// The new current-window always has zero count.
	sw.curr.Reset(newCurrStart, 0)

	// reset previous window
	newPrevCount := int64(0)
	if diffSize == 1 {
		// The new previous-window will overlap with the old current-window,
		// so it inherits the count.
		//
		// Note that the count here may be not accurate, since it is only a
		// SNAPSHOT of the current-window's count, which in itself tends to
		// be inaccurate due to the asynchronous nature of the sync behaviour.
		newPrevCount = sw.curr.Count()
	}
	sw.prev.Reset(newCurrStart.Add(-sw.size), newPrevCount)
}
