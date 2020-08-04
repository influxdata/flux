// Package function implements a utility for defining
// transformations using reflection on structs.
//
// This package is experimental and important features
// have not yet been implemented. Be cautioned to test
// well any code that uses this package to build procedure
// specs. If the struct itself is invalid, the program can
// panic.
//
// The heart of this package is in the ReadArgs function
// and the associated RegisterTransformation. Look at the
// documentation to those functions to get a better idea
// of how to use this package.
package function
