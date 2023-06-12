package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/hitzhangjie/codemaster/compress/tar"
	"github.com/iancoleman/strcase"
)

var (
	input  = flag.String("input", "", "read data from input, which could be a regular file or directory")
	output = flag.String("output", "", "write transformed data to named *.go, which could be linked with binary")
	gopkg  = flag.String("gopkg", "gobin", "write transformed data to *.go, whose package is $package")
)

var tpl = `package {{.GoPackage}}
var {{.Variable}} = []uint8{
{{ range $idx, $val := .Data }}{{$val}},{{ end }}
}`

func init() {
	flag.Parse()
}

func main() {

	// 输入输出参数校验
	if len(*input) == 0 || len(*gopkg) == 0 {
		fmt.Println("invalid argument: invalid input")
		os.Exit(1)
	}

	// 读取输入内容
	buf, err := ReadFromInputSource(*input)
	if err != nil {
		fmt.Errorf("read data error: %v\n", err)
		os.Exit(1)
	}

	// 将内容转换成go文件写出
	inputBaseName := filepath.Base(*input)
	if len(*output) == 0 {
		*output = fmt.Sprintf("%s_bindata.go", inputBaseName)
	}

	outputDir, outputBaseName := filepath.Split(*output)
	tplInstance, err := template.New(outputBaseName).Parse(tpl)
	if err != nil {
		fmt.Printf("parse template error: %v\n", err)
		os.Exit(1)
	}
	_ = os.MkdirAll(outputDir, 0777)

	fout, err := os.OpenFile(*output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("open input error: %v", err)
		os.Exit(1)
	}

	err = tplInstance.Execute(fout, &struct {
		GoPackage string
		Variable  string
		Data      []uint8
	}{
		GoPackage: *gopkg,
		Variable:  strcase.ToCamel(outputBaseName),
		Data:      buf,
	})
	if err != nil {
		panic(fmt.Errorf("template execute error: %v", err))
	}

	fmt.Printf("ok, filedata stored to %s\n", *output)
}

// ReadFromInputSource 从输入读取内容，可以是一个文件，也可以是一个目录（会先gzip压缩然后再返回内容）
func ReadFromInputSource(inputSource string) (data []byte, err error) {

	_, err = os.Lstat(inputSource)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	err = tar.Tar(inputSource, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
