// Package print provides functions for displaying Flux values.
//
// ## Metadata
// introduced: NEXT
package print


import "array"

// print outputs Flux basic and composite data types in a table.
//
// ## Parameters
// - val: Value to print.
// - result_name: Result name. Default is `_result`.
//
// ## Examples
//
// ### Print a value extracted from a stream of tables
// ```
// import "contrib/anaisdg/print"
// import "import "sampledata""
//
// value = (sampledata.float() |> findRecord(fn: (key) => true,idx: 0))._value
//
// >  print.print(val: value)
// ```
print = (val, result_name="_result") =>
    array.from(rows: [{"_value": display(v: val)}]) |> yield(name: display(v: result_name))
