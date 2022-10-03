// Package dynamic provides tools for working with values of unknown types.
//
// ## Metadata
// introduced: 0.185.0
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

// asArray converts a dynamic value into an array of dynamic elements.
//
// The dynamic input value must be an array. If it is not an array, `dynamic.asArray()` returns an error.
//
// ## Parameters
// - v: Dynamic value to convert. Default is the piped-forward value (`<-`).
//
// ## Metadata
// tags: type-conversions
builtin asArray : (<-v: dynamic) => [dynamic]

// _isNotDistinct returns true if both values print the same way, indicating
// they are essentially equivalent.
_isNotDistinct = (a, b) => display(v: a) == display(v: b)
