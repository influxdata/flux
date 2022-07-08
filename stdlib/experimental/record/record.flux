// Package record provides tools for working with Flux records.
//
// **Note**: The `experimental/record` package is an interim solution for
// [influxdata/flux#3461](https://github.com/influxdata/flux/issues/3461) and
// will either be removed after this issue is resolved or promoted out of the
// experimental package if other uses are found.
//
// ## Metadata
// introduced; 0.131.0
//
package record


// any is a polymorphic record value that can be used as a default record value
// when input record property types are not known.
//
builtin any : A where A: Record

// get returns a value from a record by key name or a default value if the key
// doesn’t exist in the record.
//
// This is an interim solution for the exists operator’s limited use with
// records outside of a stream of tables.
// For more information, see [influxdata/flux#4073](https://github.com/influxdata/flux/issues/4073).
//
// ## Parameters
// - r: Record to retrieve the value from.
// - key: Property key to retrieve.
// - default: Default value to return if the specified key does not exist in the record.
//
// ## Examples
// ### Dynamically return a value from a recordd
// ```no_run
// import "experimental/record"
//
// key = "foo"
// exampleRecord = {foo: 1.0, bar: "hello"}
//
// record.get(r: exampleRecord, key: key, default: "")
//
// // Returns 1.0
// ```
//
// ## Metadata
// introduced: 0.134.0
//
builtin get : (r: A, key: string, default: B) => B where A: Record
