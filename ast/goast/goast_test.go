package goast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
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

func TestGoAst_ParseFile(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "pb_test.go", src, parser.ParseComments)
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

func TestGoAst_ParseDir(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	for _, pkg := range pkgs {
		for fname, file := range pkg.Files {
			fmt.Printf("working on file %v\n", fname)
			ast.Inspect(file, func(n ast.Node) bool {
				// perform analysis here
				// ...
				fn, ok := n.(*ast.FuncDecl)
				if !ok {
					return true
				}

				buf := bytes.Buffer{}
				err := format.Node(&buf, fset, fn)
				if err != nil {
					panic(err)
				}
				fmt.Printf("%s\n", buf.String())

				return true
			})
			//buf := new(bytes.Buffer)
			//err := format.Node(buf, fset, file)
			//if err != nil {
			//	fmt.Printf("error: %v\n", err)
			//} else {
			//	//if fname[len(fname)-8:] != "_test.go" {
			//	//	ioutil.WriteFile(fname, buf.Bytes(), 0664)
			//	//}
			//	fmt.Printf("%s\n", buf.String())
			//}
		}
	}
}

func TestGoAst_ParseFile_AppendFunc(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "pb_test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parser.ParseFile: %v", err)
	}

	decl := &ast.FuncDecl{
		Doc:  nil,
		Recv: nil,
		Name: ast.NewIdent("helloworld"),
		Type: &ast.FuncType{
			Func:    0,
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			Lbrace: 0,
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Return:  0,
					Results: nil,
				},
			},
			Rbrace: 0,
		},
	}

	file.Decls = append(file.Decls, decl)

	buf := &bytes.Buffer{}
	err = format.Node(buf, fset, file)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", buf.String())
}
