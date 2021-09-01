// Package array provides functions for interacting with arrays.
package array


import ejson "experimental/json"
import "json"

// append is a function that appends value to array.
//
// ## Parameters
// - `arr` is the array to operate on.
// - `v` is the value to append to the array.
//
//   Array may be empty. If not, value type must match the type of existing element(s).
//
// ## Append to array
//
// ```
// import "contrib/bonitoo-io/array"
//
// good = ["foo", "bar"]
//
// from(bucket: "my-bucket")
//   |> range(start: -1h)
//   |> keep(columns: array.append(arr: ["_time", "_value"], v: good)
// ```
builtin append : (arr: [A], v: [A]) => [A]

// map is a function that applies supplied function to each element and returns a new array.
//
// ## Parameters
// - `arr` is the array to operate on.
// - `fn` is the function to convert items.
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

// empty JSON-encoded array as string
emptyStr = "[]"

// fromStr parses JSON-encoded (as string) array into array.
//
// ## Parameters
// - `arr` is the JSON-encoded (as string) array to operate on.
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

// appendStr is a function that appends value to JSON-encoded array.
//
// ## Parameters
// - `arr` is the JSON-encoded array to operate on.
// - `v` is the value to append to the array.
//
//   Array may be empty. If not, value type must match the type of existing element(s).
//   This variant of append is useful in transformations, because flux table column cannot be of array type.
//
// ## Append to array
//
// ```
// import "contrib/bonitoo-io/array"
//
// from(bucket: "my-bucket")
//   |> range(start: -1h)
//   |> reduce(
//       fn: (r, accumulator) => ({
//           sarr: array.appendStr(arr: accumulator.sarr, v: [r._value])
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
appendStr = (arr = emptyStr, v) => string(v: json.encode(v: append(arr: fromStr(arr: arr), v: v)))
