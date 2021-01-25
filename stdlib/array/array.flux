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
builtin from : (rows: [A]) => [A] where A: Record
