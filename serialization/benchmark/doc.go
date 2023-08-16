package benchmark

/*
case: Benchmark_Unmarshal_Slice_HasNoSchema

goos: linux
goarch: amd64
pkg: github.com/hitzhangjie/codemaster/serialization/benchmark
cpu: AMD EPYC 7K62 48-Core Processor
Benchmark_Unmarshal_Slice_HasNoSchema
Benchmark_Unmarshal_Slice_HasNoSchema/4-KB.gen.json
Benchmark_Unmarshal_Slice_HasNoSchema/4-KB.gen.json/Go/encoding/json-16         	   13708	     88161 ns/op	   29472 B/op	     772 allocs/op
Benchmark_Unmarshal_Slice_HasNoSchema/4-KB.gen.json/Bytedance/sonic-16          	   28861	     41247 ns/op	   32063 B/op	     337 allocs/op

Benchmark_Unmarshal_Slice_HasNoSchema/9-KB.gen.json/Go/encoding/json-16         	    6218	    177443 ns/op	   58800 B/op	    1536 allocs/op
Benchmark_Unmarshal_Slice_HasNoSchema/9-KB.gen.json/Bytedance/sonic-16          	   14572	     80466 ns/op	   63397 B/op	     668 allocs/op

Benchmark_Unmarshal_Slice_HasNoSchema/24-KB.gen.json/Go/encoding/json-16        	    2776	    430432 ns/op	  146256 B/op	    3817 allocs/op
Benchmark_Unmarshal_Slice_HasNoSchema/24-KB.gen.json/Bytedance/sonic-16         	    5584	    208300 ns/op	  159452 B/op	    1659 allocs/op

Benchmark_Unmarshal_Slice_HasNoSchema/48-KB.gen.json/Go/encoding/json-16        	    1374	    871859 ns/op	  292874 B/op	    7614 allocs/op
Benchmark_Unmarshal_Slice_HasNoSchema/48-KB.gen.json/Bytedance/sonic-16         	    2899	    418713 ns/op	  321376 B/op	    3310 allocs/op

Benchmark_Unmarshal_Slice_HasNoSchema/95-KB.gen.json/Go/encoding/json-16        	     681	   1762727 ns/op	  581275 B/op	   15204 allocs/op
Benchmark_Unmarshal_Slice_HasNoSchema/95-KB.gen.json/Bytedance/sonic-16         	    1340	    892073 ns/op	  628717 B/op	    6611 allocs/op

Benchmark_Unmarshal_Slice_HasNoSchema/240-KB.gen.json/Go/encoding/json-16       	     261	   4549906 ns/op	 1464852 B/op	   37976 allocs/op
Benchmark_Unmarshal_Slice_HasNoSchema/240-KB.gen.json/Bytedance/sonic-16        	     511	   2344420 ns/op	 1570166 B/op	   16513 allocs/op

Benchmark_Unmarshal_Slice_HasNoSchema/480-KB.gen.json/Go/encoding/json-16       	     126	   9455119 ns/op	 2910188 B/op	   75938 allocs/op
Benchmark_Unmarshal_Slice_HasNoSchema/480-KB.gen.json/Bytedance/sonic-16        	     249	   4770760 ns/op	 3142408 B/op	   33015 allocs/op
*/
