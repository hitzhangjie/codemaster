package ast_rewrite_code

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var src = `// Package main just outputs hello world
package main

// comment 1

// comment 2

// comment 3
// comment 4

import (
	"fmt"
)

func add(a, b int) int {
	return a + b
}

func main() {
	fmt.Println("hello world")
}`

func TestASTRewriteCode(t *testing.T) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	assert.Nil(t, err)

	//for _, g := range f.Comments {
	//	fmt.Println(g.Text())
	//}

	// clear comments
	f.Comments = nil
	f.Doc = nil

	//ast.Print(fset, f)

	for _, f := range f.Decls {
		fn, ok := f.(*ast.FuncDecl)
		if !ok {
			continue
		}

		println(fn.Name.Name)

		if fn.Name.String() != "add" {
			continue
		}
		fn.Doc = &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Slash: 0,
					Text:  "// " + fn.Name.String() + " do something ...\n",
				},
			},
		}
	}

	printer.Fprint(os.Stdout, fset, f)
}
