// Package sample provides functions that return basic sample datasets.
package sample


import "csv"

option start = 2021-01-01T00:00:00Z
option stop = 2021-01-01T00:01:00Z

_numeric = "
#group,false,false,false,true,false
#datatype,string,long,dateTime:RFC3339,string,double
#default,_result,,,,
,result,table,_time,tid,_value
,,0,2021-01-01T00:00:00Z,t1,2.18
,,0,2021-01-01T00:00:10Z,t1,10.92
,,0,2021-01-01T00:00:20Z,t1,7.35
,,0,2021-01-01T00:00:30Z,t1,17.53
,,0,2021-01-01T00:00:40Z,t1,15.23
,,0,2021-01-01T00:00:50Z,t1,4.43
,,1,2021-01-01T00:00:00Z,t2,19.85
,,1,2021-01-01T00:00:10Z,t2,4.97
,,1,2021-01-01T00:00:20Z,t2,3.75
,,1,2021-01-01T00:00:30Z,t2,19.77
,,1,2021-01-01T00:00:40Z,t2,13.86
,,1,2021-01-01T00:00:50Z,t2,1.86
"

_numericNull = "
#group,false,false,false,true,false
#datatype,string,long,dateTime:RFC3339,string,double
#default,_result,,,,
,result,table,_time,tid,_value
,,0,2021-01-01T00:00:00Z,t1,2.18
,,0,2021-01-01T00:00:10Z,t1,
,,0,2021-01-01T00:00:20Z,t1,7.35
,,0,2021-01-01T00:00:30Z,t1,
,,0,2021-01-01T00:00:40Z,t1,
,,0,2021-01-01T00:00:50Z,t1,4.43
,,1,2021-01-01T00:00:00Z,t2,
,,1,2021-01-01T00:00:10Z,t2,4.97
,,1,2021-01-01T00:00:20Z,t2,3.75
,,1,2021-01-01T00:00:30Z,t2,19.77
,,1,2021-01-01T00:00:40Z,t2,
,,1,2021-01-01T00:00:50Z,t2,1.86
"

_string = "
#group,false,false,false,true,false
#datatype,string,long,dateTime:RFC3339,string,string
#default,_result,,,,
,result,table,_time,tid,_value
,,0,2021-01-01T00:00:00Z,t1,smpl_g9qczs
,,0,2021-01-01T00:00:10Z,t1,smpl_0mgv9n
,,0,2021-01-01T00:00:20Z,t1,smpl_phw664
,,0,2021-01-01T00:00:30Z,t1,smpl_guvzy4
,,0,2021-01-01T00:00:40Z,t1,smpl_5v3cce
,,0,2021-01-01T00:00:50Z,t1,smpl_s9fmgy
,,1,2021-01-01T00:00:00Z,t2,smpl_b5eida
,,1,2021-01-01T00:00:10Z,t2,smpl_eu4oxp
,,1,2021-01-01T00:00:20Z,t2,smpl_5g7tz4
,,1,2021-01-01T00:00:30Z,t2,smpl_sox1ut
,,1,2021-01-01T00:00:40Z,t2,smpl_wfm757
,,1,2021-01-01T00:00:50Z,t2,smpl_dtn2bv
"

_stringNull = "
#group,false,false,false,true,false
#datatype,string,long,dateTime:RFC3339,string,string
#default,_result,,,,
,result,table,_time,tid,_value
,,0,2021-01-01T00:00:00Z,t1,smpl_g9qczs
,,0,2021-01-01T00:00:10Z,t1,
,,0,2021-01-01T00:00:20Z,t1,smpl_phw664
,,0,2021-01-01T00:00:30Z,t1,
,,0,2021-01-01T00:00:40Z,t1,
,,0,2021-01-01T00:00:50Z,t1,smpl_s9fmgy
,,1,2021-01-01T00:00:00Z,t2,
,,1,2021-01-01T00:00:10Z,t2,smpl_eu4oxp
,,1,2021-01-01T00:00:20Z,t2,smpl_5g7tz4
,,1,2021-01-01T00:00:30Z,t2,smpl_sox1ut
,,1,2021-01-01T00:00:40Z,t2,
,,1,2021-01-01T00:00:50Z,t2,smpl_dtn2bv
"

_bool = "#group,false,false,false,true,false
#datatype,string,long,dateTime:RFC3339,string,boolean
#default,_result,,,,
,result,table,_time,tid,_value
,,0,2021-01-01T00:00:00Z,t1,true
,,0,2021-01-01T00:00:10Z,t1,true
,,0,2021-01-01T00:00:20Z,t1,false
,,0,2021-01-01T00:00:30Z,t1,true
,,0,2021-01-01T00:00:40Z,t1,false
,,0,2021-01-01T00:00:50Z,t1,false
,,1,2021-01-01T00:00:00Z,t2,false
,,1,2021-01-01T00:00:10Z,t2,true
,,1,2021-01-01T00:00:20Z,t2,false
,,1,2021-01-01T00:00:30Z,t2,true
,,1,2021-01-01T00:00:40Z,t2,true
,,1,2021-01-01T00:00:50Z,t2,false
"

_boolNull = "#group,false,false,false,true,false
#datatype,string,long,dateTime:RFC3339,string,boolean
#default,_result,,,,
,result,table,_time,tid,_value
,,0,2021-01-01T00:00:00Z,t1,true
,,0,2021-01-01T00:00:10Z,t1,
,,0,2021-01-01T00:00:20Z,t1,false
,,0,2021-01-01T00:00:30Z,t1,
,,0,2021-01-01T00:00:40Z,t1,
,,0,2021-01-01T00:00:50Z,t1,false
,,1,2021-01-01T00:00:00Z,t2,
,,1,2021-01-01T00:00:10Z,t2,true
,,1,2021-01-01T00:00:20Z,t2,false
,,1,2021-01-01T00:00:30Z,t2,true
,,1,2021-01-01T00:00:40Z,t2,
,,1,2021-01-01T00:00:50Z,t2,false
"

// float returns a sample data set with float values.
//
// ## Parameters
//
// - `includeNull` indicates whether or not to include null values in the returned dataset.
//   Default is `false`.
//
// ## Output basic sample data with float values
//
// ```
// import "sample"
//
// sample.float()
// ```
//
float = (includeNull=false) => {
    _csvData = if not includeNull then _numeric else _numericNull

    return csv.from(csv: _csvData)
}

// int returns a sample data set with integer values.
//
// ## Parameters
//
// - `includeNull` indicates whether or not to include null values in the returned dataset.
//   Default is `false`.
//
// ## Output basic sample data with integer values
//
// ```
// import "sample"
//
// sample.int()
// ```
//
int = (includeNull=false) => {
    _csvData = if not includeNull then _numeric else _numericNull

    return csv.from(csv: _csvData) |> toInt()
}

// uint returns a sample data set with unsigned integer values.
//
// ## Parameters
//
// - `includeNull` indicates whether or not to include null values in the returned dataset.
//   Default is `false`.
//
// ## Output basic sample data with unsigned integer values
//
// ```
// import "sample"
//
// sample.uint()
// ```
//
uint = (includeNull=false) => {
    _csvData = if not includeNull then _numeric else _numericNull

    return csv.from(csv: _csvData) |> toUInt()
}

// string returns a sample data set with string values.
//
// ## Parameters
//
// - `includeNull` indicates whether or not to include null values in the returned dataset.
//   Default is `false`.
//
// ## Output basic sample data with string values
//
// ```
// import "sample"
//
// sample.string()
// ```
//
string = (includeNull=false) => {
    _csvData = if not includeNull then _string else _stringNull

    return csv.from(csv: _csvData)
}

// bool returns a sample data set with boolean values.
//
// ## Parameters
//
// - `includeNull` indicates whether or not to include null values in the returned dataset.
//   Default is `false`.
//
// ## Output basic sample data with boolean values
//
// ```
// import "sample"
//
// sample.bool()
// ```
//
bool = (includeNull=false) => {
    _csvData = if not includeNull then _bool else _boolNull

    return csv.from(csv: _csvData)
}

// numericString returns a sample data set with numeric string values.
//
// ## Parameters
//
// - `includeNull` indicates whether or not to include null values in the returned dataset.
//   Default is `false`.
//
// ## Output basic sample data with numeric string values
//
// ```
// import "sample"
//
// sample.numericString()
// ```
//
numericString = (includeNull=false) => {
    _csvData = if not includeNull then _numeric else _numericNull

    return csv.from(csv: _csvData) |> toString()
}

// numericBool returns a sample data set with numeric (integer) boolean values.
//
// ## Parameters
//
// - `includeNull` indicates whether or not to include null values in the returned dataset.
//   Default is `false`.
//
// ## Output basic sample data with numeric boolean values
//
// ```
// import "sample"
//
// sample.numericBool()
// ```
//
numericBool = (includeNull=false) => {
    _csvData = if not includeNull then _bool else _boolNull

    return csv.from(csv: _csvData) |> toInt()
}
