package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/tdigest"
)

func Test_tdigest_quantile(t *testing.T) {
	td := tdigest.NewWithCompression(100)
	iterations := 100000
	for i := 0; i < iterations; i++ {
		td.Add(float64(time.Microsecond*time.Duration(15)), 1)
		td.Add(float64(time.Microsecond*time.Duration(15)), 1)
		td.Add(float64(time.Microsecond*time.Duration(15)), 1)
		td.Add(float64(time.Millisecond*time.Duration(88)), 1)
		td.Add(float64(time.Millisecond*time.Duration(90)), 1)
	}
	fmt.Printf("iteration-%d, p90-%.2f,p99-%.2f\n", iterations, td.Quantile(0.90), td.Quantile(0.99))
}
