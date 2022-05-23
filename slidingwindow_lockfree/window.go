package slidingwindow

import (
	"time"
)

// window represents a window that ignores sync behavior entirely
// and only stores counters in memory.
type window struct {
	// The start boundary (timestamp in nanoseconds) of the window.
	// [start, start + size)
	start int64

	// The total count of events happened in the window.
	count int64
}

func newLocalWindow() *window {
	return &window{}
}

func (w *window) Start() time.Time {
	return time.Unix(0, w.start)
}

func (w *window) Count() int64 {
	return w.count
}

func (w *window) AddCount(n int64) {
	w.count += n
}

func (w *window) Reset(s time.Time, c int64) {
	w.start = s.UnixNano()
	w.count = c
}
