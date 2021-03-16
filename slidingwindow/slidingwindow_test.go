package slidingwindow

import (
	"fmt"
	"testing"
	"time"
)

type mockClock struct {
	now time.Time
}

func (m *mockClock) Now() time.Time {
	return m.now
}

func (m *mockClock) Add(d time.Duration) {
	m.now = m.now.Add(d)
}

func (m *mockClock) Set(t time.Time) {
	m.now = t
}

func TestSlidingWindow(t *testing.T) {
	c := &mockClock{}
	startTime, _ := time.Parse(time.RFC3339, "2017-01-14T12:00:00Z")
	c.Set(startTime)

	sw := NewSlidingWindow(WithStep(24*time.Hour), WithDuration(7*24*time.Hour), WithClock(c))

	sw.Insert(1.0)
	sw.Insert(2.0)
	fmt.Println(sw.Max())

	c.Add(24 * time.Hour)
	sw.Insert(1.2)
	fmt.Println(sw.Max())

	c.Add(24 * time.Hour)
	sw.Insert(1.3)
	fmt.Println(sw.Max())

	c.Add(24 * time.Hour)
	sw.Insert(1.4)
	fmt.Println(sw.Max())

	c.Add(24 * time.Hour)
	sw.Insert(1.5)
	fmt.Println(sw.Max())

	c.Add(24 * time.Hour)
	sw.Insert(1.6)
	fmt.Println(sw.Max())

	c.Add(24 * time.Hour)
	sw.Insert(1.7)
	fmt.Println(sw.Max())

	c.Add(24 * time.Hour)
	sw.Insert(1.8)
	fmt.Println(sw.Max())

	c.Add(24 * time.Hour)
	sw.Insert(1.9)
	fmt.Println(sw.Max())
}
