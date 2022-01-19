// Package types provides functions for working with Flux's types.
//
// introduced: 0.140.0
// tags: types
package types


// isType returns true if `v` is an object matching `type`.
// This function can only check against the basic types:
// `string`, `bytes`, `int`, `uint`, `float`, `bool`, `time`, `duration`, `regexp`.
//
// ## Parameters
// - v: The value which to check the type of.
// - type: A string describing the the type to check against.
builtin isType : (v: A, type: string) => bool where A: Basic
