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

// asArray attempts to convert a dynamic value into an array of dynamic elements.
// If the value is not an array, `asArray` will result in a fatal error.
//
// ## Parameters
// - v: Dynamic value to convert. Defaults is the piped-forward value (`<-`).
//
// ## Metadata
// tags: type-conversions
builtin asArray : (<-v: dynamic) => [dynamic]
