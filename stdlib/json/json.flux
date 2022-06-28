// Package json provides tools for working with JSON.
//
// ## Metadata
// introduced: 0.40.0
//
package json


// encode converts a value into JSON bytes.
//
// This function encodes Flux types as follows:
//
// - **time** values in [RFC3339](https://docs.influxdata.com/influxdb/cloud/reference/glossary/#rfc3339-timestamp) format
// - **duration** values in number of milliseconds since the Unix epoch
// - **regexp** values as their string representation
// - **bytes** values as base64-encoded strings
// - **function** values are not encoded and produce an error
//
// ## Parameters
// - v: Value to convert.
//
// ## Examples
//
// ### Encode a value as JSON bytes
// ```no_run
// import "json"
//
// jsonData = {foo: "bar", baz: 123, quz: [4, 5, 6]}
//
// json.encode(v: jsonData)
//
// // Returns [123 34 98 97 122 34 58 49 50 51 44 34 102 111 111 34 58 34 98 97 114 34 44 34 113 117 122 34 58 91 52 44 53 44 54 93 125]
// ```
//
// ## Metadata
// tags: type-conversions
//
builtin encode : (v: A) => bytes
