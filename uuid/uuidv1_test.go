package uuid

import (
	"fmt"
	"sync"
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

// UUIDv1 isn't safe for generating unique id in parallel.
// If want to generate unique id in parallel, use UUIDv4 instead.
func Test_UUIDv1_Generate_Dup(t *testing.T) {
	var existed sync.Map
	for i := 0; i < 10; i++ {
		go func() {
			for {
				v, err := uuid.NewUUID()
				if err != nil {
					panic(err)
				}
				_, loaded := existed.LoadOrStore(v, v)
				if loaded {
					panic(fmt.Sprintf("found dup uuid: %v", v.String()))
				}
			}
		}()
	}

	select {}
}
