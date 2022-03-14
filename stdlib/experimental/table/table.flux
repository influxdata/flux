// Package table provides tools working with Flux tables.
//
// ## Metadata
// introduced: 0.115.0
//
package table


// fill adds a single row to empty tables in a stream of tables.
//
// Columns that are in the group key are filled with the column value defined in the group key.
// Columns not in the group key are filled with a null value.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Fill empty tables
// ```
// import "experimental/table"
// import "sampledata"
//
// data = sampledata.int()
//     |> filter(fn: (r) => r.tag != "t2", onEmpty: "keep")
//
// < data
// >     |> table.fill()
// ```
//
// ## Metadata
// tags: transformations,table
//
builtin fill : (<-tables: stream[A]) => stream[A] where A: Record
