// Package benchmark_test 这里整理了一些常见的json库的序列化、反序列化、get、set性能对比，
// 以及大致整理下json在这些方面面临的挑战，以及各个json库实现做的好的地方。
//
// 主要对比下面这些：
// - stdlib, encoding/json
// - sonic, https://www.libhunt.com/r/bytedance/sonic
// - fastjson, https://www.libhunt.com/r/valyala/fastjson
// - jsoniter, https://www.libhunt.com/r/jsoniter
// - encoding, https://www.libhunt.com/r/encoding
// - simdjson, https://www.libhunt.com/r/simdjson
// - simdjson-go, https://www.libhunt.com/r/simdjson-go
// - rapidjson, https://www.libhunt.com/r/rapidjson
package benchmark_test
