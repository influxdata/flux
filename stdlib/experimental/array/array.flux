// Package array provides functions for manipulating arrays and for building tables from Flux arrays.
//
// introduced: 0.79.0
// tags: array,tables
//
package array


import "array"

// from constructs a table from an array of records.
//
// The `experimental/array.from()` function was promoted to the `array` package in
// Flux 0.103.0. This function is available for backwards compatibility, but we
// recommend using the `array` package instead.
//
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
// deprecated: 0.103.0
from = array.from

// concat appends two arrays and returns a new array.
//
// ## Parameters
// - arr: First array. Default is the piped-forward array (`<-`).
// - v: Array to append to the first array.
//
// Neither input array is mutated and a new array is returned.
//
// ## Examples
// ### Merge two arrays
//
// ```
// import "experimental/array"
//
// a = [1, 2, 3]
// b = [4, 5, 6]
//
// c = a |> array.concat(v: b)
// // Returns [1, 2, 3, 4, 5, 6]
//
// // Output each value in the array as a row in a table
// > array.from(rows: c |> array.map(fn: (x) => ({_value: x})))
// ```
//
// introduced: 0.155.0
builtin concat : (<-arr: [A], v: [A]) => [A]

// map iterates over an array, applies a function to each element to produce a new element,
// and then returns a new array.
//
// ## Parameters
// - arr: Array to operate on. Defaults is the piped-forward array (`<-`).
// - fn: Function to apply to elements. The element is represented by `x` in the function.
//
// ## Examples
// ### Convert an array of integers to an array of records
//
// ```
// import "experimental/array"
//
// a = [1, 2, 3, 4, 5]
// b = a |> array.map(fn: (x) => ({_value: x}))
// // b returns [{_value: 1}, {_value: 2}, {_value: 3}, {_value: 4}, {_value: 5}]
//
// // Output the array of records as a table
// > array.from(rows: b)
// ```
//
// introduced: 0.155.0
builtin map : (<-arr: [A], fn: (x: A) => B) => [B]

// filter iterates over an array, evaluates each element with a predicate function, and then returns
// a new array with only elements that match the predicate.
//
// ## Parameters
// - arr: Array to filter. Default is the piped-forward array (`<-`).
// - fn: Predicate function to evaluate on each element.
//   The element is represented by `x` in the predicate function.
//
// ## Examples
//
// ### Filter array of integers
//
// ```
// import "experimental/array"
//
// a = [1, 2, 3, 4, 5]
// b = a |> array.filter(fn: (x) => x >= 3)
// // b returns [3, 4, 5]
//
// // Output the filtered array as a table
// > array.from(rows: b |> array.map(fn: (x) => ({_value: x})))
// ```
// introduced: 0.155.0
builtin filter : (<-arr: [A], fn: (x: A) => bool) => [A]
