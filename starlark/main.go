package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func main() {
	// Create a new Starlark thread
	thread := &starlark.Thread{
		Name: "repl",
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Println(msg)
		},
	}

	// Create a new global environment
	globals := starlark.StringDict{}

	// Create a scanner for reading input
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Starlark REPL (type 'exit' to quit)")

	errExit := errors.New("exit")

	for {
		// Print prompt
		fmt.Print(">>> ")

		// Read input
		readline := func() ([]byte, error) {
			if !scanner.Scan() {
				return nil, io.EOF
			}
			line := strings.TrimSpace(scanner.Text())

			if line == "exit" {
				return nil, errExit
			}
			if line == "" {
				return nil, nil
			}
			return []byte(line + "\n"), nil
		}

		// Execute the input
		if err := rep(readline, thread, globals); err != nil {
			if err == io.EOF {
				break
			}
			if err == errExit {
				os.Exit(0)
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
}

// rep reads, evaluates, and prints one item.
//
// It returns an error (possibly readline.ErrInterrupt)
// only if readline failed. Starlark errors are printed.
func rep(readline func() ([]byte, error), thread *starlark.Thread, globals starlark.StringDict) error {
	eof := false

	f, err := syntax.ParseCompoundStmt("<stdin>", readline)
	if err != nil {
		if eof {
			return io.EOF
		}
		printError(err)
		return nil
	}

	if expr := soleExpr(f); expr != nil {
		//TODO: check for 'exit'
		// eval
		v, err := evalExprOptions(nil, thread, expr, globals)
		if err != nil {
			printError(err)
			return nil
		}

		// print
		if v != starlark.None {
			fmt.Println(v)
		}
	} else {
		// compile
		prog, err := starlark.FileProgram(f, globals.Has)
		if err != nil {
			printError(err)
			return nil
		}

		// execute (but do not freeze)
		res, err := prog.Init(thread, globals)
		if err != nil {
			printError(err)
		}

		// The global names from the previous call become
		// the predeclared names of this call.
		// If execution failed, some globals may be undefined.
		for k, v := range res {
			globals[k] = v
		}
	}

	return nil
}

var defaultSyntaxFileOpts = &syntax.FileOptions{
	Set:             true,
	While:           true,
	TopLevelControl: true,
	GlobalReassign:  true,
	Recursion:       true,
}

// evalExprOptions is a wrapper around starlark.EvalExprOptions.
// If no options are provided, it uses default options.
func evalExprOptions(opts *syntax.FileOptions, thread *starlark.Thread, expr syntax.Expr, globals starlark.StringDict) (starlark.Value, error) {
	if opts == nil {
		opts = defaultSyntaxFileOpts
	}
	return starlark.EvalExprOptions(opts, thread, expr, globals)
}

func soleExpr(f *syntax.File) syntax.Expr {
	if len(f.Stmts) == 1 {
		if stmt, ok := f.Stmts[0].(*syntax.ExprStmt); ok {
			return stmt.X
		}
	}
	return nil
}

// printError prints the error to stderr,
// or its backtrace if it is a Starlark evaluation error.
func printError(err error) {
	if evalErr, ok := err.(*starlark.EvalError); ok {
		fmt.Fprintln(os.Stderr, evalErr.Backtrace())
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
}
