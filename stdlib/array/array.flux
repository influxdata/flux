package array


// from will construct a table from the input rows.
//
// This function takes the `rows` parameter. The rows
// parameter is an array of records that will be constructed.
// All of the records must have the same keys and the same types
// for the values.
//
// Example:
//
//    import "array"
//    array.from(rows:[{a:1, b: false, c: "hi"}, {a:2, b: true, c: "bye"}])
//


// from constructs a table from an array of ercords
//
// Each record in the array is converted into an output row or record. All
// records must have the same keys and data types.
//
// ## Parameters
// - `rows` is the array of records to construct a table with
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
