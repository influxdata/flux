//+build gofuzz

package parser

import "github.com/influxdata/flux/ast"

// Fuzz will run the parser on the input data and return 1 on success and 0 on failure.
func Fuzz(data []byte) int {
	program := NewAST(string(data))
	if ast.Check(program) > 0 {
		return 0
	}
	return 1
}
