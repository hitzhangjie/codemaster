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

// GoFuncMap stores registered Go functions
var GoFuncMap = map[string]interface{}{
	"Add": Add,
}

func Add(a, b int) int {
	fmt.Println("Hey! I'm a Go function!")
	return a + b
}

// callGoFunc is a Starlark function that calls registered Go functions
func callGoFunc(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("call_gofunc requires at least one argument (function name)")
	}

	funcName, ok := args[0].(starlark.String)
	if !ok {
		return nil, fmt.Errorf("first argument must be a string (function name)")
	}

	goFunc, ok := GoFuncMap[string(funcName)]
	if !ok {
		return nil, fmt.Errorf("function %s not found", funcName)
	}

	// Convert Starlark arguments to Go values
	goArgs := make([]interface{}, len(args)-1)
	for i, arg := range args[1:] {
		switch v := arg.(type) {
		case starlark.Int:
			if v, ok := v.Int64(); ok {
				goArgs[i] = int(v)
			} else {
				return nil, fmt.Errorf("integer too large")
			}
		case starlark.Float:
			goArgs[i] = float64(v)
		case starlark.String:
			goArgs[i] = string(v)
		case starlark.Bool:
			goArgs[i] = bool(v)
		default:
			return nil, fmt.Errorf("unsupported argument type: %T", arg)
		}
	}

	// Call the Go function
	switch f := goFunc.(type) {
	case func(int, int) int:
		if len(goArgs) != 2 {
			return nil, fmt.Errorf("Add function requires exactly 2 arguments")
		}
		a, ok1 := goArgs[0].(int)
		b, ok2 := goArgs[1].(int)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("Add function requires integer arguments")
		}
		result := f(a, b)
		return starlark.MakeInt(result), nil
	default:
		return nil, fmt.Errorf("unsupported function type: %T", goFunc)
	}
}

func main() {
	go func() {
		// Create a new Starlark thread
		thread := &starlark.Thread{
			Name: "repl",
			Print: func(thread *starlark.Thread, msg string) {
				fmt.Println(msg)
			},
		}

		// Create a new global environment with call_gofunc
		globals := starlark.StringDict{
			"call_gofunc": starlark.NewBuiltin("call_gofunc", callGoFunc),
		}

		// Create a scanner for reading input
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Starlark REPL (type 'exit' to quit)")
		fmt.Println("Example1: starlark exprs and stmts")
		fmt.Println("Example2: call_gofunc('Add', 1, 2)")

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
	}()

	select {}
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
