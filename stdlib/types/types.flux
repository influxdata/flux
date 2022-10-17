// Package types provides functions for working with Flux's types.
//
// ## Metadata
// introduced: 0.141.0
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
// ### Aggregate or select data based on type
// ```
// # import "csv"
// # import "sampledata"
// import "types"
// #
// # data = csv.from(
// #     csv: "
// # #group,false,false,false,true,false
// # #datatype,string,long,dateTime:RFC3339,string,double
// # #default,_result,,,,
// # ,result,table,_time,type,_value
// # ,,0,2021-01-01T00:00:00Z,float,-2.18
// # ,,0,2021-01-01T00:00:10Z,float,10.92
// # ,,0,2021-01-01T00:00:20Z,float,7.35
// # ,,0,2021-01-01T00:00:30Z,float,17.53
// # ,,0,2021-01-01T00:00:40Z,float,15.23
// # ,,0,2021-01-01T00:00:50Z,float,4.43
// #
// # #group,false,false,false,true,false
// # #datatype,string,long,dateTime:RFC3339,string,boolean
// # #default,_result,,,,
// # ,result,table,_time,type,_value
// # ,,0,2021-01-01T00:00:00Z,bool,true
// # ,,0,2021-01-01T00:00:10Z,bool,true
// # ,,0,2021-01-01T00:00:20Z,bool,false
// # ,,0,2021-01-01T00:00:30Z,bool,true
// # ,,0,2021-01-01T00:00:40Z,bool,false
// # ,,0,2021-01-01T00:00:50Z,bool,false
// #
// # #group,false,false,false,true,false
// # #datatype,string,long,dateTime:RFC3339,string,string
// # #default,_result,,,,
// # ,result,table,_time,type,_value
// # ,,0,2021-01-01T00:00:00Z,string,smpl_g9qczs
// # ,,0,2021-01-01T00:00:10Z,string,smpl_0mgv9n
// # ,,0,2021-01-01T00:00:20Z,string,smpl_phw664
// # ,,0,2021-01-01T00:00:30Z,string,smpl_guvzy4
// # ,,0,2021-01-01T00:00:40Z,string,smpl_5v3cce
// # ,,0,2021-01-01T00:00:50Z,string,smpl_s9fmgy
// #
// # #group,false,false,false,false,true
// # #datatype,string,long,dateTime:RFC3339,long,string
// # #default,_result,,,,
// # ,result,table,_time,_value,type
// # ,,0,2021-01-01T00:00:00Z,-2,int
// # ,,0,2021-01-01T00:00:10Z,10,int
// # ,,0,2021-01-01T00:00:20Z,7,int
// # ,,0,2021-01-01T00:00:30Z,17,int
// # ,,0,2021-01-01T00:00:40Z,15,int
// # ,,0,2021-01-01T00:00:50Z,4,int
// # ",
// # )
// #     |> range(start: sampledata.start, stop: sampledata.stop)
//
// < nonNumericData = data
//     |> filter(fn: (r) => types.isType(v: r._value, type: "string") or types.isType(v: r._value, type: "bool"))
//     |> aggregateWindow(every: 30s, fn: last)
//
// numericData = data
//     |> filter(fn: (r) => types.isType(v: r._value, type: "int") or types.isType(v: r._value, type: "float"))
//     |> aggregateWindow(every: 30s, fn: mean)
//
// > union(tables: [nonNumericData, numericData])
// ```
//
// ## Metadata
// tags: types, tests
//
builtin isType : (v: A, type: string) => bool where A: Basic

// isNumeric tests if a value is a numeric type (int, uint, or float).
//
// This is a helper function to test or filter for values that can be used in
// arithmatic operations or aggregations.
//
// ## Parameters
// - v: Value to test.
//
// ## Examples
//
// ### Filter by numeric values
// ```
// # import "csv"
// import "types"
//
// # data =
// #     csv.from(
// #         csv: "
// # #group,false,false,false,true,false
// # #datatype,string,long,dateTime:RFC3339,string,double
// # #default,_result,,,,
// # ,result,table,_time,type,_value
// # ,,0,2021-01-01T00:00:00Z,float,-2.18
// # ,,0,2021-01-01T00:00:10Z,float,10.92
// # ,,0,2021-01-01T00:00:20Z,float,7.35
// #
// # #group,false,false,false,true,false
// # #datatype,string,long,dateTime:RFC3339,string,boolean
// # #default,_result,,,,
// # ,result,table,_time,type,_value
// # ,,0,2021-01-01T00:00:00Z,bool,true
// # ,,0,2021-01-01T00:00:10Z,bool,true
// # ,,0,2021-01-01T00:00:20Z,bool,false
// #
// # #group,false,false,false,true,false
// # #datatype,string,long,dateTime:RFC3339,string,string
// # #default,_result,,,,
// # ,result,table,_time,type,_value
// # ,,0,2021-01-01T00:00:00Z,string,smpl_g9qczs
// # ,,0,2021-01-01T00:00:10Z,string,smpl_0mgv9n
// # ,,0,2021-01-01T00:00:20Z,string,smpl_phw664
// #
// # #group,false,false,false,true,false
// # #datatype,string,long,dateTime:RFC3339,string,long
// # #default,_result,,,,
// # ,result,table,_time,type,_value
// # ,,0,2021-01-01T00:00:00Z,int,-2
// # ,,0,2021-01-01T00:00:10Z,int,10
// # ,,0,2021-01-01T00:00:20Z,int,7
// # ",
// #     )
// #
// < data
// >     |> filter(fn: (r) => types.isNumeric(v: r._value))
// ```
//
// ## Metadata
// introduced: 0.187.0
// tags: types, tests
isNumeric = (v) =>
    isType(v: v, type: "int") or isType(v: v, type: "uint") or isType(v: v, type: "float")
