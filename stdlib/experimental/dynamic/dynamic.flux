// Package dynamic provides tools for working with values of unknown types.
//
// ## Metadata
// introduced: NEXT
//
package dynamic


// dynamic wraps a value so it can be used as a `dynamic` value.
//
// ## Parameters
// - v: Value to wrap as dynamic.
//
// ## Metadata
// tags: type-conversions
builtin dynamic : (v: A) => dynamic
