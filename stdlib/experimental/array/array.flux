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
// **Deprecated**: Experimental `array.from()` is deprecated in favor of
// [`array.from()`](https://docs.influxdata.com/flux/v0.x/stdlib/array/from).
// This function is available for backwards compatibility, but we recommend using the `array` package instead.
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
// **Deprecated**: Experimetnal `array.concat()` is deprecated in favor of
// [`array.concat()`](https://docs.influxdata.com/flux/v0.x/stdlib/array/concat).
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
// deprecated: 0.173.0
concat = array.concat

// map iterates over an array, applies a function to each element to produce a new element,
// and then returns a new array.
//
// **Deprecated**: Experimental `array.map()` is deprecated in favor of
// [`array.map()`](https://docs.influxdata.com/flux/v0.x/stdlib/array/map).
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
// deprecated: 0.173.0
map = array.map

// filter iterates over an array, evaluates each element with a predicate function, and then returns
// a new array with only elements that match the predicate.
//
// **Deprecated**: Experimental `array.filter()` is deprecated in favor of
// [`array.filter()`](https://docs.influxdata.com/flux/v0.x/stdlib/array/filter).
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
// deprecated: 0.173.0
filter = array.filter

// toBool converts all values in an array to booleans.
//
// #### Supported array types
//
// - `[string]` with values `true` or `false`
// - `[int]` with values `1` or `0`
// - `[uint]` with values `1` or `0`
// - `[float]` with values `1.0` or `0.0`
//
// ## Parameters
// - arr: Array of values to convert. Default is the piped-forward array (`<-`).
//
// ## Examples
//
// ### Convert an array of integers to booleans
// ```no_run
// import "experimental/array"
//
// arr = [1, 1, 0, 1, 0]
//
// array.toBool(arr: arr)
// // Returns [true, true, false, true, false]
// ```
//
// ## Metadata
// introduced: 0.184.0
// tags: type-conversions
//
toBool = (arr=<-) => array.map(arr: arr, fn: (x) => bool(v: x))

// toDuration converts all values in an array to durations.
//
// #### Supported array types and behaviors
//
// - `[int]` (parsed as nanosecond epoch timestamps)
// - `[string]` with values that use [duration literal](https://docs.influxdata.com/flux/v0.x/data-types/basic/duration/#duration-syntax) representation.
// - `[uint]` (parsed as nanosecond epoch timestamps)
//
// ## Parameters
// - arr: Array of values to convert. Default is the piped-forward array (`<-`).
//
// ## Examples
//
// ### Convert an array of integers to durations
// ```no_run
// import "experimental/array"
//
// arr = [80000000000, 56000000000, 132000000000]
//
// array.toDuration(arr: arr)
// // Returns [1m20s, 56s, 2m12s]
// ```
//
// ## Metadata
// introduced: 0.184.0
// tags: type-conversions
//
toDuration = (arr=<-) => array.map(arr: arr, fn: (x) => duration(v: x))

// toFloat converts all values in an array to floats.
//
// #### Supported array types
//
// - `[string]` (numeric, scientific notation, ±Inf, or NaN)
// - `[bool]`
// - `[int]`
// - `[uint]`
//
// ## Parameters
// - arr: Array of values to convert. Default is the piped-forward array (`<-`).
//
// ## Examples
//
// ### Convert an array of integers to floats
// ```no_run
// import "experimental/array"
//
// arr = [12, 24, 36, 48]
//
// array.toFloat(arr: arr)
// // Returns [12.0, 24.0, 36.0, 48.0]
// ```
//
// ### Convert an array of strings to floats
// ```no_run
// import "experimental/array"
//
// arr = ["12", "1.23e+4", "NaN", "24.2"]
//
// array.toFloat(arr: arr)
// // Returns [12.0, 1.2300, NaN, 24.2]
// ```
//
// ## Metadata
// introduced: 0.184.0
// tags: type-conversions
//
toFloat = (arr=<-) => array.map(arr: arr, fn: (x) => float(v: x))

// toInt converts all values in an array to integers.
//
// #### Supported array types and behaviors
//
// | Array type   | Returned array values                      |
// | :----------- | :----------------------------------------- |
// | `[bool]`     | 1 (true) or 0 (false)                      |
// | `[duration]` | Number of nanoseconds in the duration      |
// | `[float]`    | Value truncated at the decimal             |
// | `[string]`   | Integer equivalent of the numeric string   |
// | `[time]`     | Equivalent nanosecond epoch timestamp      |
// | `[uint]`     | Integer equivalent of the unsigned integer |
//
// ## Parameters
// - arr: Array of values to convert. Default is the piped-forward array (`<-`).
//
// ## Examples
//
// ### Convert an array of floats to integers
// ```no_run
// import "experimental/array"
//
// arr = [12.1, 24.2, 36.3, 48.4]
//
// array.toInt(arr: arr)
// // Returns [12, 24, 36, 48]
// ```
//
// ## Metadata
// introduced: 0.184.0
// tags: type-conversions
//
toInt = (arr=<-) => array.map(arr: arr, fn: (x) => int(v: x))

// toString converts all values in an array to strings.
//
// #### Supported array types
//
// - `[bool]`
// - `[duration]`
// - `[float]`
// - `[int]`
// - `[time]`
// - `[uint]`
//
// ## Parameters
// - arr: Array of values to convert. Default is the piped-forward array (`<-`).
//
// ## Examples
//
// ### Convert an array of floats to strings
// ```no_run
// import "experimental/array"
//
// arr = [12.0, 1.2300, NaN, 24.2]
//
// array.toString(arr: arr)
// // Returns ["12.0", "1.2300", "NaN", "24.2"]
// ```
//
// ## Metadata
// introduced: 0.184.0
// tags: type-conversions
//
toString = (arr=<-) => array.map(arr: arr, fn: (x) => string(v: x))

// toTime converts all values in an array to times.
//
// #### Supported array types
//
// - `[int]` (parsed as nanosecond epoch timestamps)
// - `[string]` with values that use [time literal](https://docs.influxdata.com/flux/v0.x/data-types/basic/time/#time-syntax)
//    representation (RFC3339 timestamps).
// - `[uint]` (parsed as nanosecond epoch timestamps)
//
// ## Parameters
// - arr: Array of values to convert. Default is the piped-forward array (`<-`).
//
// ## Examples
//
// ### Convert an array of integers to time values
// ```no_run
// import "experimental/array"
//
// arr = [1640995200000000000, 1643673600000000000, 1646092800000000000]
//
// array.toTime(arr: arr)
// // Returns [2022-01-01T00:00:00Z, 2022-02-01T00:00:00Z, 2022-03-01T00:00:00Z]
// ```
//
// ## Metadata
// introduced: 0.184.0
// tags: type-conversions
//
toTime = (arr=<-) => array.map(arr: arr, fn: (x) => time(v: x))

// toUInt converts all values in an array to unsigned integers.
//
// #### Supported array types and behaviors
//
// | Array type   | Returned array values                      |
// | :----------- | :----------------------------------------- |
// | `[bool]`     | 1 (true) or 0 (false)                      |
// | `[duration]` | Number of nanoseconds in the  duration     |
// | `[float]`    | Value truncated at the decimal             |
// | `[int]`      | Unsigned integer equivalent of the integer |
// | `[string]`   | Integer equivalent of the numeric string   |
// | `[time]`     | Equivalent nanosecond epoch timestamp      |
//
// ## Parameters
// - arr: Array of values to convert. Default is the piped-forward array (`<-`).
//
// ## Examples
//
// ### Convert an array of floats to usigned integers
// ```no_run
// import "experimental/array"
//
// arr = [-12.1, 24.2, -36.3, 48.4]
//
// array.toInt(arr: arr)
// // Returns [18446744073709551604, 24, 18446744073709551580, 48]
// ```
//
// ## Metadata
// introduced: 0.184.0
// tags: type-conversions
//
toUInt = (arr=<-) => array.map(arr: arr, fn: (x) => uint(v: x))
