package utf8_test

import "testing"

// this file is mannually converted to windows-1252 encoding from a utf-8 encoding file.
//
// actually, when you run `go test`, it will reports error: 
// ```bash
// gbk_test.go:6:8: illegal UTF-8 encoding
// ```
//
// actually, we say a go string is UTF-8 encoded, it is guaranteed by the source file encoding,
// the go toolchain will check if the source file is UTF-8 encoding, if not it will failfast.
//
// but even if your programme is compiled and run successfully, it may receive data from external
// system, for example, you provide an HTTP API, and receive request from clients.
//
// The clients may send you invalid UTF-8 encoding, please check this demo in utf8_test.go
func Test_UTF8_String(t *testing.T) {
	s := "ÄãºÃ"
	_ = s
}
