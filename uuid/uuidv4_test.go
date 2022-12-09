package uuid

import (
	"testing"

	"github.com/google/uuid"
)

func Test_UUIDv4_Generate(t *testing.T) {
	v := uuid.New()
	t.Logf("uuid: %v", v)
}

/*
Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^Benchmark_UUIDv4_Generate$ dsvr/common/dlock -v -count=1

goos: linux
goarch: amd64
pkg: dsvr/common/dlock
cpu: AMD EPYC 7K62 48-Core Processor
Benchmark_UUIDv4_Generate
Benchmark_UUIDv4_Generate-16    	  279673	      4304 ns/op	      16 B/op	       1 allocs/op
*/
func Benchmark_UUIDv4_Generate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := uuid.New()
		_ = v
	}
}

/*
Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^Benchmark_UUIDv4_ParallelGenerate$ dsvr/common/dlock -v -count=1

goos: linux
goarch: amd64
pkg: dsvr/common/dlock
cpu: AMD EPYC 7K62 48-Core Processor
Benchmark_UUIDv4_ParallelGenerate
Benchmark_UUIDv4_ParallelGenerate-16    	  263300	      4648 ns/op	      16 B/op	       1 allocs/op
*/
func Benchmark_UUIDv4_ParallelGenerate(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			v := uuid.New()
			_ = v
		}
	})
}
