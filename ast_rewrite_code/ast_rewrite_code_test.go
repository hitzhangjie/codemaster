package ast_rewrite_code

import (
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

	printer.Fprint(os.Stdout, fset, f)
}
