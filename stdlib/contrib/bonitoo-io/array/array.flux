// Package array provides functions for interacting with arrays.
package array


import ejson "experimental/json"
import "json"

// concat joins two arrays. It returns a new array.
//
// ## Parameters
// - arr: The first array.
// - v: Additional array to join.
//
//   Array may be empty. If not, value type must match the type of existing element(s).
//
// ## Concat arrays
//
// ```
// import "contrib/bonitoo-io/array"
//
// good = ["foo", "bar"]
//
// from(bucket: "my-bucket")
//   |> range(start: -1h)
//   |> keep(columns: array.concat(arr: ["_time", "_value"], v: good)
// ```
builtin concat : (arr: [A], v: [A]) => [A]

// map is a function that applies supplied function to each element. It returns a new array.
//
// ## Parameters
// - arr: The array to operate on.
// - fn: The function to be applied on items.
//
// ## Convert array of ints to array of strings
//
// ```
// import "contrib/bonitoo-io/array"
//
// ia = [1, 1]
//
// sa = array.map(arr: a, fn: (x) => string(v: x))
// ```
builtin map : (arr: [A], fn: (x: A) => B) => [B]

// emptyStr is a string representing JSON-encoded empty array
emptyStr = "[]"

// fromStr parses a string representing JSON-encoded array into an array.
//
// ## Parameters
// - arr: The string representing JSON-encoded array.
//
// ## Parse simple type array
//
// ```
// import "contrib/bonitoo-io/array"
//
// s = "[1, 2, 3]"
//
// a = array.fromStr(arr: s)
// ```
//
// ## Parse records array
//
// ```
// import "contrib/bonitoo-io/array"
//
// s = "[{\"n\": 1}, {\"n\": 2}, {\"n\": 3}]"
//
// a = array.fromStr(arr: s)
// ```
fromStr = (arr) => ejson.parse(data: bytes(v: arr))

// concatStr joins two arrays. It returns JSON-encoded array represented as a string.
// This variant of `concat` is intended to be used in `reduce` aggregation,
// because Flux table column cannot be of array type.
//
// ## Parameters
// - arr: The string representing JSON-encoded first array.
// - v: The second array to join.
//
//   Array may be empty. If not, value type must match the type of existing element(s).
//
// ## Concat arrays in reduce()
//
// ```
// import "contrib/bonitoo-io/array"
//
// from(bucket: "my-bucket")
//   |> range(start: -1h)
//   |> reduce(
//       fn: (r, accumulator) => ({
//           sarr: array.concatStr(arr: accumulator.sarr, v: [r._value])
//       }),
//       identity: {
//           sarr: array.emptyStr  // "[]"
//       }
//   )
//   |> map(fn: (r) => ({ r with status:
//       http.post(
//           url: "http://endpoint:12345/",
//           data: json.encode(v: array.fromStr(arr: r.sarr))
//       )})
//  )
// ```
concatStr = (arr=emptyStr, v) => string(v: json.encode(v: concat(arr: fromStr(arr: arr), v: v)))
