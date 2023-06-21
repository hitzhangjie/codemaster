package timer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimer(t *testing.T) {
	wait := time.Second * 2

	t.Run("wait expired", func(t *testing.T) {
		now := time.Now()
		timer := time.NewTimer(wait)
		<-timer.C
		assert.GreaterOrEqual(t, time.Since(now), wait)
	})

	t.Run("reset alive timer", func(t *testing.T) {
		timer := time.NewTimer(wait)
		time.Sleep(wait / 2)
		assert.True(t, timer.Reset(time.Second))
	})

	t.Run("reset expired timer", func(t *testing.T) {
		timer := time.NewTimer(wait)
		time.Sleep(wait)
		assert.False(t, timer.Reset(time.Second))
	})
}

func TestTimerReset(t *testing.T) {

}
