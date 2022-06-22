// Package hex provides functions that perform hexadecimal conversion
// of `int`, `uint` or `bytes` values to and from `string` values.
//
// ## Metadata
// introduced: 0.131.0
// contributors: **GitHub**: [@sranka](https://github.com/sranka), [@bonitoo-io](https://github.com/bonitoo-io) | **InfluxDB Slack**: [@sranka](https://influxdata.com/slack)
//
package hex


// int converts a hexadecimal string to an integer.
//
// ## Parameters
//
// - v: String to convert.
//
// ## Examples
// ### Convert hexadecimal string to integer
// ```no_run
// import "contrib/bonitoo-io/hex"
//
// hex.int(v: "4d2")
//
// // Returns 1234
// ```
//
// ## Metadata
// tag: type-conversion
builtin int : (v: string) => int

// string converts a Flux basic type to a hexadecimal string.
//
// The function is similar to `string()`, but encodes int, uint, and bytes
// types to hexadecimal lowercase characters.
//
// ## Parameters
//
// - v: Value to convert.
//
// ## Examples
// ### Convert integer to hexadecimal string
// ```no_run
// import "contrib/bonitoo-io/hex"
//
// hex.string(v: 1234)
//
// // Returns 4d2
// ```
//
// ### Convert a boolean to a hexadecimal string value
// ```no_run
// import "contrib/bonitoo-io/hex"
//
// hex.string(v: true)
//
// // Returns "true"
// ```
//
// ### Convert a duration to a hexadecimal string value
// ```no_run
// import "contrib/bonitoo-io/hex"
//
// hex.string(v: 1m)
//
// // Returns "1m"
// ```
//
// ### Convert a time to a hexadecimal string value
// ```no_run
// import "contrib/bonitoo-io/hex"
//
// hex.string(v: 2021-01-01T00:00:00Z)
//
// // Returns "2021-01-01T00:00:00Z"
// ```
//
// ### Convert an integer to a hexadecimal string value
// ```no_run
// import "contrib/bonitoo-io/hex"
//
// hex.string(v: 1234)
//
// // Returns "4d2"
// ```
//
// ### Convert a uinteger to a hexadecimal string value
// ```no_run
// import "contrib/bonitoo-io/hex"
//
// hex.string(v: uint(v: 5678))
//
// // Returns "162e"
// ```
//
// ### Convert a float to a hexadecimal string value
// ```no_run
// import "contrib/bonitoo-io/hex"
//
// hex.string(v: 10.12)
//
// // Returns "10.12"
// ```
//
// ### Convert bytes to a hexadecimal string value
// ```no_run
// import "contrib/bonitoo-io/hex"
//
// hex.string(v: bytes(v: "Hello world!"))
//
// // Returns "48656c6c6f20776f726c6421"
// ```
//
// ### Convert all values in a column to hexadecimal string values
// Use `map()` to iterate over and update all input rows.
// Use `hex.string()` to update the value of a column.
// The following example uses data provided by the sampledata package.
//
// ```
// import "sampledata"
// import "contrib/bonitoo-io/hex"
//
// data = sampledata.int()
//     |> map(fn: (r) => ({ r with _value: r._value * 1000 }))
//
// < data
// >     |> map(fn:(r) => ({ r with _value: hex.string(v: r.foo) }))
// ```
//
// ## Metadata
// tag: type-conversion
builtin string : (v: A) => string

// uint converts a hexadecimal string to an unsigned integer.
//
// ## Parameters
//
// - v: String to convert.
//
// ## Examples
// ### Convert a hexadecimal string to an unsigned integer
// ```no_run
// import "contrib/bonitoo-io/hex"
//
// hex.uint(v: "4d2")
//
// // Returns 1234
// ```
//
// ## Metadata
// tag: type-conversion
builtin uint : (v: string) => uint

// bytes converts a hexadecimal string to bytes.
//
// ## Parameters
//
// - v: String to convert.
//
// ## Examples
// ### Convert a hexadecimal string into bytes
// ```no_run
// import "contrib/bonitoo-io/hex"
//
// hex.bytes(v: "FF5733")
//
// // Returns [255 87 51] (bytes)
// ```
//
// ## Metadata
// tag: type-conversion
builtin bytes : (v: string) => bytes
