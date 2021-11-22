// Package rows provides additional functions for remapping values in rows.
//
// introduced: 0.77.0
package rows


// map is an alternate implementation of `map()` that is faster, but more limited.
//
// `rows.map()` cannot modify groups keys and, therefore, does not need to regroup tables.
// **Attempts to change columns in the group key are ignored.**
//
// ## Parameters
//
// - fn: A single argument function to apply to each record.
//   The return value must be a record.
//
//   _(Use the `with` operator to preserve columns not in the group key
//   and not explicitly mapped in the operation.)_
// 
// - tables: Input data. Default is piped-forward data (`<-`).
//
//
// ## Examples
// ### Perform mathemtical operations on column values
// The following example returns the square of each value in the `_value` column:
//
// ```
// import "contrib/jsternberg/rows"
// import "sampledata"
//
// sampledata.int()
//     |> rows.map(fn: (r) => ({ _value: r._value * r._value }))
// ```
//
// **Important**: The `_time` column is dropped because:
// 
// - It’s not in the group key.
// - It’s not explicitly mapped in the operation.
// - The `with` operator was not used to include existing columns.
//
// ### Preserve all columns in the operation
//
// Use the with operator in your mapping operation to preserve all columns,
// including those not in the group key, without explicitly remapping them.
//
// ```
// import "contrib/jsternberg/rows"
// import "sampledata"
//
// sampledata.int()
//     |> rows.map(fn: (r) => ({ r with _value: r._value * r._value }))
// ```
//
// **Important**: The mapping operation remaps the `_value` column.
// The `with` operator preserves all other columns not in the group key (`_time`).
//
// ### Attempt to remap columns in the group key
// ```
// import "contrib/jsternberg/rows"
// import "sampledata"
//
// sampledata.int()
//     |> rows.map(fn: (r) => ({ r with tag: "3" }))
// ```
//
// **Important**: Remapping the `tag` column to `"t3"` is ignored because `tag` is part of the group key.
// The `with` operator preserves columns not in the group key (`_time` and `_value`).
//
// tags: transformations
//
builtin map : (<-tables: [A], fn: (r: A) => B) => [B] where A: Record, B: Record
