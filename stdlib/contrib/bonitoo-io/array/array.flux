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
