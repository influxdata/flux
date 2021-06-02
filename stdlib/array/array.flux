package array


// array.from constructs a table from an array of records. each
// record in the array is converted into an output row or record.
// all records must have the same keys and data types.
//
// - `rows` is the array of records that is used to construct a table.
//
// Example:
//
//    array.from(rows:[{a:1, b: false, c: "hi"}, {a:2, b: true, c: "bye"}])
//
builtin from : (rows: [A]) => [A] where A: Record
