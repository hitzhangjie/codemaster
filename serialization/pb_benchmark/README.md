对protobuf相关的官方库、三方库进行下说明:

|name|repo|commit|remark|
|----|----|------|------|
|protobuf apiv1|https://github.com/golang/protobuf|v1.3.5|apiv1官方版本|
|gogo|https://github.com/gogo/protobuf|f67b897|对上述版本的优化|
|protobuf apiv2|https://github.com/protocolbuffers/protobuf-go|-|apiv2官方版本|
|protobuf apiv2'|https://github.com/golang/protobuf|v1.5.3|基于apiv2官方版本重写，只是没有反射接口|
|vtproto|https://github.com/planetscale/vtprotobuf|1a874c6|对上述版本apiv2'的优化|

因为go版本选择算法的问题，我们没法将多个不同的版本同时引用进来并使用，也就没法在一个benchmark测试里来对比github.com/golang/protobuf apiv1、apiv2 两个版本之间的性能差异……so，我们拆成多个module来测试，然后将结果汇总在当前README.md中。


测试用的pb序列化后size为3KB左右，此时各个pb库的性能如下:

```
apiv1   Benchmark_HelloRequest_NotEmpty-10         9745     115210 ns/op    66848 B/op     4020 allocs/op
apiv2   Benchmark_HelloRequest_NotEmpty-10        26912      43315 ns/op    17824 B/op     1387 allocs/op
gogo    Benchmark_HelloRequest_NotEmpty-10        90948      11385 ns/op     4864 B/op        1 allocs/op
vtproto Benchmark_HelloRequest_NotEmpty-10       103699      11230 ns/op     4096 B/op        1 allocs/op
```

ps: go不支持对相同module不同版本的引入，除非是大版本，否则支持很差。在我们这个测试中本来是想将多个库的测试集中到一起测试，因为这个问题只能拆开测试。将结果汇总到这里来查看。

执行下面的命令来确定编译构建中模块的版本选择算法最终确定使用的版本号，如apiv1这个用例下面：

```bash
$ cd apiv1
$ go list -m all | grep github.com/golang/protobuf
github.com/golang/protobuf v1.3.5
```
