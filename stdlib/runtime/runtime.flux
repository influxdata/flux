// Package runtime provides information about the current Flux runtime.
//
// introduce: 0.38.0
//
package runtime


// version returns the current Flux version.
//
// ## Examples
// ### Return the Flux version in a stream of tables
// ```
// import "array"
// import "runtime"
//
// > array.from(rows: [{version: runtime.version()}])
// ```
//
builtin version : () => string
