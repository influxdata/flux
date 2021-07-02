// Package array provides functions for building tables from flux arrays.
package array


// from constructs a table from an array of records.
//
// Each record in the array is converted into an output row or record. All
// records must have the same keys and data types.
//
// ## Parameters
// - `rows` is the array of records to construct a table with.
//
// ## Build an arbitrary table
//
// ```
// import "array"
//
// rows = [
//   {foo: "bar", baz: 21.2},
//   {foo: "bar", baz: 23.8}
// ]
//
// array.from(rows: rows)
// ```
//
// ## Union custom rows with query results
//
// ```
// import "influxdata/influxdb/v1"
// import "array"
//
// tags = v1.tagValues(
//   bucket: "example-bucket",
//   tag: "host"
// )
//
// wildcard_tag = array.from(rows: [{_value: "*"}])
//
// union(tables: [tags, wildcard_tag])
// ```
builtin from : (rows: [A]) => [A] where A: Record
