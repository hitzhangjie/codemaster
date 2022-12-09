package uuid

import (
	"math/rand"
	"testing"
)

func Test_Random_Generate(t *testing.T) {
	v := rand.Uint64()
	t.Logf("rand: %v", v)
}

/*
Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^Benchmark_Random_Generate$ dsvr/common/dlock -v -count=1

goos: linux
goarch: amd64
pkg: dsvr/common/dlock
cpu: AMD EPYC 7K62 48-Core Processor
Benchmark_Random_Generate
Benchmark_Random_Generate-16    	98875276	        12.13 ns/op	       0 B/op	       0 allocs/op
*/
func Benchmark_Random_Generate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := rand.Uint64()
		_ = v
		//t.Logf("rand: %v", v)
	}
}

/*
Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^Benchmark_Random_ParallelGenerate$ dsvr/common/dlock -v -count=1

goos: linux
goarch: amd64
pkg: dsvr/common/dlock
cpu: AMD EPYC 7K62 48-Core Processor
Benchmark_Random_ParallelGenerate
Benchmark_Random_ParallelGenerate-16    	16960431	        85.23 ns/op	       0 B/op	       0 allocs/op
*/
func Benchmark_Random_ParallelGenerate(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			v := rand.Uint64()
			_ = v
			//t.Logf("rand: %v", v)
		}
	})
}
