// Package hex provides functions that perform hexadecimal conversion
// of `int`, `uint` or `bytes` values to and from `string` values.
package hex

// int converts a hexadecimal string to an integer.
//
// ## Parameters
//
// - v: string to convert.
builtin int : (v: string) => int

// string converts an integer string to a hexadecimal string.
//
// ## Parameters
//
// - v: string to convert.
builtin string : (v: A) => string

// uint converts a hexadecimal string to an unsigned integer.
//
// ## Parameters
//
// - v: string to convert.
builtin uint : (v: string) => uint

// bytes converts a hexadecimal string to bytes.
//
// ## Parameters
//
// - v: string to convert.
builtin bytes : (v: string) => bytes
