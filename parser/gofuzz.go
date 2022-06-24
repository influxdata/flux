//go:build gofuzz
// +build gofuzz

package parser

import "github.com/mvn-trinhnguyen2-dn/flux/ast"

// Fuzz will run the parser on the input data and return 1 on success and 0 on failure.
func Fuzz(data []byte) int {
	pkg := ParseSource(string(data))
	if ast.Check(pkg) > 0 {
		return 0
	}
	return 1
}
