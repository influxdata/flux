// Package print provides function for converting values into tables
// ## Metadata
// introduced: NEXT
package print


import "array"

print = (val, result_name) => array.from(rows: [{"_value": display(v: val)}]) |> yield(name: display(v: result_name))
