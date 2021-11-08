// Package hex provides functions that perform hexadecimal conversion
// of `int`, `uint` or `bytes` values to and from `string` values.
//
// introduced: 0.131.0
package hex

// int converts a hexadecimal string to an integer.
//
// ## Parameters
//
// - v: String to convert.
builtin int : (v: string) => int

// string converts a Flux basic type to a hexadecimal string.
//
// The function is similar to `string()`, but encodes int, uint, and bytes
// types to hexadecimal lowercase characters.
//
// ## Parameters
//
// - v: Value to convert.
builtin string : (v: A) => string

// uint converts a hexadecimal string to an unsigned integer.
//
// ## Parameters
//
// - v: String to convert.
builtin uint : (v: string) => uint

// bytes converts a hexadecimal string to bytes.
//
// ## Parameters
//
// - v: String to convert.
builtin bytes : (v: string) => bytes
