// Package array provides functions for building tables from Flux arrays.
// 
// The `experimental/array` package was promoted to the `array` package in
// Flux 0.103.0. This package is avaiable for backwards compatibility, but we
// recommend using the `array` package instead.
//
// introduced: 0.79.0
// deprecated: 0.103.0
// tags: array,tables
// 
package array


import "array"

// from constructs a table from an array of records.
//
// Each record in the array is converted into an output row or record. All
// records must have the same keys and data types.
//
// ## Parameters
// - rows: Array of records to construct a table with.
//
// ## Examples
//
// ### Build an arbitrary table
// ```
// import "experimental/array"
//
// rows = [
//     {foo: "bar", baz: 21.2},
//     {foo: "bar", baz: 23.8},
// ]
//
// > array.from(rows: rows)
// ```
//
// ### Union custom rows with query results
// ```no_run
// import "influxdata/influxdb/v1"
// import "experimental/array"
//
// tags = v1.tagValues(
//     bucket: "example-bucket",
//     tag: "host",
// )
//
// wildcard_tag = array.from(rows: [{_value: "*"}])
//
// union(tables: [tags, wildcard_tag])
// ```
//
from = array.from
