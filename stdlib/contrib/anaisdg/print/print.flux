// Package print provides function for converting values into tables
// ## Metadata
// introduced: NEXT
package print


import "array"

// print function converts other data types to a table
// ## Parameters
// - val: The value you want to "print".
// - result_name: The name you want your table to have.
// ## Examples
//
// ### Perform a linear regression on a dataset
// ```no_run
// import "contrib/anaisdg/print"
// import "print"
//
// < value = (sampledata.float() |> findRecord(fn: (key) => true,idx: 0))._value
// >  print.print(val: value, result_name: "extracted value")
// ```
print = (val, result_name) => array.from(rows: [{"_value": display(v: val)}]) |> yield(name: display(v: result_name))
