// Package types provides functions for working with Flux's types.
//
// introduced: 0.140.0
// tags: types
//
package types


// isType tests if a value is a specified type.
//
// ## Parameters
// - v: Value to test.
// - type: String describing the type to check against.
//
//     **Supported types**:
//     - string
//     - bytes
//     - int
//     - uint
//     - float
//     - bool
//     - time
//     - duration
//     - regexp
//
// ## Examples
//
// ### Filter by value type
// ```
// # import "csv"
// import "types"
// #
// # csvData =
// #     "
// # #datatype,string,long,dateTime:RFC3339,string,double
// # #group,false,false,false,true,false
// # #default,_result,,,,
// # ,result,table,_time,_field,_value
// # ,,0,2022-01-01T00:00:00Z,foo,12
// # ,,0,2022-01-01T00:01:00Z,foo,15
// # ,,0,2022-01-01T00:02:00Z,foo,9
// #
// # #datatype,string,long,dateTime:RFC3339,string,string
// # #group,false,false,false,true,false
// # #default,_result,,,,
// # ,result,table,_time,_field,_value
// # ,,1,2022-01-01T00:00:00Z,bar,0jCcsMYM
// # ,,1,2022-01-01T00:01:00Z,bar,jHvuDw35
// # ,,1,2022-01-01T00:02:00Z,bar,HE5uCIC2
// # "
// #
// # data = csv.from(csv: csvData)
//
// < data
// >     |> filter(fn: (r) => types.isType(v: r._value, type: "string"))
// ```
//
// tags: types, tests
//
builtin isType : (v: A, type: string) => bool where A: Basic
