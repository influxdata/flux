// Package array provides functions for manipulating arrays and for building tables from Flux arrays.
//
// **Deprecated**: This package is deprecated in favor of [`array`](https://docs.influxdata.com/flux/v0.x/stdlib/array/).
//
// ## Metadata
// introduced: 0.79.0
// tags: array,tables
//
package array


import "array"

// from constructs a table from an array of records.
//
// **Deprecated**: `from()` is deprecated in favor of [`from()`](https://docs.influxdata.com/flux/v0.x/stdlib/array/from).
// This function is available for backwards compatibility, but we recommend using the `array` package instead.
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
// ## Metadata
// deprecated: 0.103.0
from = array.from

// concat appends two arrays and returns a new array.
//
// **Deprecated**: `concat()` is deprecated in favor of [`concat()`](https://docs.influxdata.com/flux/v0.x/stdlib/array/concat).
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
// ## Metadata
// introduced: 0.155.0
// deprecated: NEXT
concat = array.concat

// map iterates over an array, applies a function to each element to produce a new element,
// and then returns a new array.
//
// **Deprecated**: `map()` is deprecated in favor of [`map()`](https://docs.influxdata.com/flux/v0.x/stdlib/array/map).
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
// ## Metadata
// introduced: 0.155.0
// deprecated: NEXT
map = array.map

// filter iterates over an array, evaluates each element with a predicate function, and then returns
// a new array with only elements that match the predicate.
//
// **Deprecated**: `filter()` is deprecated in favor of [`filter()`](https://docs.influxdata.com/flux/v0.x/stdlib/array/filter).
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
//
// ## Metadata
// introduced: 0.155.0
// deprecated: NEXT
filter = array.filter
