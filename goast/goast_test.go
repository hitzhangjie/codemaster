package goast

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

var src = `
// Package main provides a ast example.
package main

import (
	"fmt"
)

type Student struct {
	Name string
	Age int
}

// programme entry point,
// print helloworld
func main() {
	fmt.Println("helloworld")

	{
		fmt.Println("xxxxxxxxx")
	}

	student := &Student {
		Name: "jie",
		Age: 30,
	}
	fmt.Println(student)
}
`

func TestGoAst(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "main.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parser.ParseFile: %v", err)
	}

	fmt.Println("------------------------- package ----------------------------")
	fmt.Println(file.Name)

	fmt.Println("----------------------- declaration --------------------------")
	for _, decl := range file.Decls {
		fmt.Println(src[decl.Pos()-1 : decl.End()])
	}

	fmt.Println("------------------------- comment -----------------------------")
	for _, c := range file.Comments {
		fmt.Println(strings.ReplaceAll(c.Text(), "\n", " "))
	}

	fmt.Println("-------------------------- scope -----------------------------")
	buf, _ := json.MarshalIndent(file.Scope, "", "\t")
	fmt.Println(string(buf))

	fmt.Println("--------------------------- doc ------------------------------")
	fmt.Println(file.Doc.Text())

}
