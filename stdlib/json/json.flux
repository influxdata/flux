// Package json functions provide tools for working with JSON.
package json


// encode converts a value into JSON bytes
// Time values are encoded using RFC3339.
// Duration values are encoded in number of milleseconds since the epoch.
// Regexp values are encoded as their string representation.
// Bytes values are encodes as base64-encoded strings.
// Function values cannot be encoded and will produce an error.
//
// ## Parameters
// - `V` is the value to convert
//
// ## Encode all values in a column in JSON bytes
//
// ```
// import "json"
//
// from(bucket: "example-bucket")
//   |> range(start: -1h)
//   |> map(fn: (r) => ({
//       r with _value: json.encode(v: r._value)
//   }))
// ```
//
builtin encode : (v: A) => bytes

