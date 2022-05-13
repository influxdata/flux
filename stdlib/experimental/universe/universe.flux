// Package universe provides equivalent functions to the standard universe package
// but with more precise type signatures.
//
// ## Metadata
// introduced: v0.166.0
//
package universe


// columns returns the column labels in each input table.
//
// **Note:** `universe.columns()` is an experimental function with a more precise type signature.
//
// For each input table, `columns` outputs a table with the same group key
// columns and a new column containing the column labels in the input table.
// Each row in an output table contains the group key value and the label of one
//  column of the input table.
// Each output table has the same number of rows as the number of columns of the input table.
//
// ## Parameters
// - column: Name of the output column to store column labels in.
//   Default is "_value".
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### List all columns per input table
// ```
// import "experimental/universe"
// import "sampledata"
//
// sampledata.string()
//     |> universe.columns(column: "labels")
// ```
//
// ## Metadata
// introduced: v0.166.0
// tags: transformations
//
builtin columns : (<-tables: stream[A], ?column: C = "_value") => stream[{C: string}] where A: Record, C: Label

// fill replaces all null values in input tables with a non-null value.
//
// **Note:** `universe.fill()` is an experimental function with a more precise type signature.
//
// Output tables are the same as the input tables with all null values replaced
// in the specified column.
//
// ## Parameters
// - column: Column to replace null values in. Default is `_value`.
// - value: Constant value to replace null values with.
//
//   Value type must match the type of the specified column.
//
// - usePrevious: Replace null values with the previous non-null value.
//   Default is `false`.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Fill null values with a specified non-null value
// ```
// import "experimental/universe"
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> universe.fill(value: 0)
// ```
//
// ### Fill null values with the previous non-null value
// ```
// import "experimental/universe"
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> universe.fill(usePrevious: true)
// ```
//
// ## Metadata
// introduced: v0.166.0
// tags: transformations
//
builtin fill : (
        <-tables: stream[{A with C: B}],
        ?column: C = "_value",
        ?value: B,
        ?usePrevious: bool,
    ) => stream[{A with C: B}]
    where
    A: Record,
    C: Label

// mean returns the average of non-null values in a specified column from each
// input table.
//
// **Note:** `universe.mean()` is an experimental function with a more precise type signature.
//
// ## Parameters
// - column: Column to use to compute means. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the average of values in each input table
// ```
// import "experimental/universe"
// import "sampledata"
//
// < sampledata.int()
// >     |> universe.mean()
// ```
//
// ## Metadata
// introduced: v0.166.0
// tags: transformations, aggregates
//
builtin mean : (<-tables: stream[{A with C: B}], ?column: C = "_value") => stream[{C: B}]
    where
    A: Record,
    B: Numeric,
    C: Label
