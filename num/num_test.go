package num

import (
	"fmt"
	"math"
	"sync/atomic"
	"testing"
	"time"
)

func TestAddToOverflow(t *testing.T) {
	var a uint64

	s := time.Now()
	for {
		v := atomic.AddUint64(&a, 1)
		if v == math.MaxUint64 {
			break
		}
	}
	d := time.Since(s)
	fmt.Println(d)
}
