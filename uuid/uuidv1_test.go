package uuid

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_UUIDv1_Generate(t *testing.T) {
	v, err := uuid.NewUUID()
	assert.Nil(t, err)
	t.Logf("uuid: %v", v)
}

/*
Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^Benchmark_UUIDv1_Generate$ dsvr/common/dlock -v -count=1

goos: linux
goarch: amd64
pkg: dsvr/common/dlock
cpu: AMD EPYC 7K62 48-Core Processor
Benchmark_UUIDv1_Generate
Benchmark_UUIDv1_Generate-16    	12496284	        95.75 ns/op	       0 B/op	       0 allocs/op
*/
func Benchmark_UUIDv1_Generate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v, err := uuid.NewUUID()
		_ = v
		_ = err
		//assert.Nil(b, err)
		//assert.Len(b, v, 16)
		//b.Logf("uuid: %v", v)
	}
}

/*
Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^Benchmark_UUIDv1_ParallelGenerate$ dsvr/common/dlock -v -count=1

goos: linux
goarch: amd64
pkg: dsvr/common/dlock
cpu: AMD EPYC 7K62 48-Core Processor
Benchmark_UUIDv1_ParallelGenerate
Benchmark_UUIDv1_ParallelGenerate-16    	 6126909	       192.4 ns/op	       0 B/op	       0 allocs/op
*/
func Benchmark_UUIDv1_ParallelGenerate(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			v, err := uuid.NewUUID()
			_ = v
			_ = err
			//assert.Nil(b, err)
			//assert.Len(b, v, 16)
		}
	})
}
