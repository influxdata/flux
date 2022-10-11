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

// jsonParse takes JSON data as bytes and returns dynamic values.
//
// JSON input is converted to dynamic-typed values which may be converted to
// a statically typed value with `dynamic.asArray()` or casting functions in the `dynamic` package.
//
// ## Parameters
// - data: JSON data (as bytes) to parse.
//
// ## Metadata
// tags: type-conversions
// introduced: 0.186.0
builtin jsonParse : (data: bytes) => dynamic

// jsonEncode converts a dynamic value into JSON bytes.
//
// ## Parameters
// - v: Value to encode into JSON.
//
// ## Metadata
// tags: type-conversions
// introduced: 0.186.0
builtin jsonEncode : (v: dynamic) => bytes

// isType tests if a dynamic type holds a value of a specified type.
//
// ## Parameters
// - v: Value to test.
// - type: String describing the type to check against.
//
//     **Supported types**:
//     - string
//     - bytes
//     - int
//     - uint
//     - float
//     - bool
//     - time
//     - duration
//     - regexp
//     - array
//     - object
//     - function
//     - dictionary
//     - vector
//
// ## Metadata
// tags: types, tests
// introduced: 0.186.0
builtin isType : (v: dynamic, type: string) => bool
