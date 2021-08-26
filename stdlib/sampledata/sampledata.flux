// Package sampledata provides functions that return basic sample datasets.
package sampledata


import "csv"

option start = 2021-01-01T00:00:00Z
option stop = 2021-01-01T00:01:00Z

_numeric = (includeNull=false) => "
#group,false,false,false,true,false
#datatype,string,long,dateTime:RFC3339,string,double
#default,_result,,,,
,result,table,_time,tag,_value
,,0,2021-01-01T00:00:00Z,t1,-2.18
,,0,2021-01-01T00:00:10Z,t1," + (if not includeNull then "10.92" else "") + "
,,0,2021-01-01T00:00:20Z,t1,7.35
,,0,2021-01-01T00:00:30Z,t1," + (if not includeNull then "17.53" else "") + "
,,0,2021-01-01T00:00:40Z,t1," + (if not includeNull then "15.23" else "") + "
,,0,2021-01-01T00:00:50Z,t1,4.43
,,1,2021-01-01T00:00:00Z,t2," + (if not includeNull then "19.85" else "") + "
,,1,2021-01-01T00:00:10Z,t2,4.97
,,1,2021-01-01T00:00:20Z,t2,-3.75
,,1,2021-01-01T00:00:30Z,t2,19.77
,,1,2021-01-01T00:00:40Z,t2," + (if not includeNull then "13.86" else "") + "
,,1,2021-01-01T00:00:50Z,t2,1.86
"

_string = (includeNull=false) => "
#group,false,false,false,true,false
#datatype,string,long,dateTime:RFC3339,string,string
#default,_result,,,,
,result,table,_time,tag,_value
,,0,2021-01-01T00:00:00Z,t1,smpl_g9qczs
,,0,2021-01-01T00:00:10Z,t1," + (if not includeNull then "smpl_0mgv9n" else "") + "
,,0,2021-01-01T00:00:20Z,t1,smpl_phw664
,,0,2021-01-01T00:00:30Z,t1," + (if not includeNull then "smpl_guvzy4" else "") + "
,,0,2021-01-01T00:00:40Z,t1," + (if not includeNull then "smpl_5v3cce" else "") + "
,,0,2021-01-01T00:00:50Z,t1,smpl_s9fmgy
,,1,2021-01-01T00:00:00Z,t2," + (if not includeNull then "smpl_b5eida" else "") + "
,,1,2021-01-01T00:00:10Z,t2,smpl_eu4oxp
,,1,2021-01-01T00:00:20Z,t2,smpl_5g7tz4
,,1,2021-01-01T00:00:30Z,t2,smpl_sox1ut
,,1,2021-01-01T00:00:40Z,t2," + (if not includeNull then "smpl_wfm757" else "") + "
,,1,2021-01-01T00:00:50Z,t2,smpl_dtn2bv
"

_bool = (includeNull=false) => "#group,false,false,false,true,false
#datatype,string,long,dateTime:RFC3339,string,boolean
#default,_result,,,,
,result,table,_time,tag,_value
,,0,2021-01-01T00:00:00Z,t1,true
,,0,2021-01-01T00:00:10Z,t1," + (if not includeNull then "true" else "") + "
,,0,2021-01-01T00:00:20Z,t1,false
,,0,2021-01-01T00:00:30Z,t1," + (if not includeNull then "true" else "") + "
,,0,2021-01-01T00:00:40Z,t1," + (if not includeNull then "false" else "") + "
,,0,2021-01-01T00:00:50Z,t1,false
,,1,2021-01-01T00:00:00Z,t2," + (if not includeNull then "false" else "") + "
,,1,2021-01-01T00:00:10Z,t2,true
,,1,2021-01-01T00:00:20Z,t2,false
,,1,2021-01-01T00:00:30Z,t2,true
,,1,2021-01-01T00:00:40Z,t2," + (if not includeNull then "true" else "") + "
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
// import "sampledata"
//
// sampledata.float()
// ```
// 
// ## Output data
// 
// | tag | _time                | _value |
// | :-: | :------------------- | -----: |
// | t1  | 2021-01-01T00:00:00Z |  -2.18 |
// | t1  | 2021-01-01T00:00:10Z |  10.92 |
// | t1  | 2021-01-01T00:00:20Z |   7.35 |
// | t1  | 2021-01-01T00:00:30Z |  17.53 |
// | t1  | 2021-01-01T00:00:40Z |  15.23 |
// | t1  | 2021-01-01T00:00:50Z |   4.43 |

// | tag | _time                | _value |
// | :-: | :------------------- | -----: |
// | t2  | 2021-01-01T00:00:00Z |  19.85 |
// | t2  | 2021-01-01T00:00:10Z |   4.97 |
// | t2  | 2021-01-01T00:00:20Z |  -3.75 |
// | t2  | 2021-01-01T00:00:30Z |  19.77 |
// | t2  | 2021-01-01T00:00:40Z |  13.86 |
// | t2  | 2021-01-01T00:00:50Z |   1.86 |
// 
float = (includeNull=false) => {
    _csvData = _numeric(includeNull:includeNull)

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
// import "sampledata"
//
// sampledata.int()
// ```
// 
// ## Output data
// 
// | tag | _time                | _value |
// | :-: | :------------------- | -----: |
// | t1  | 2021-01-01T00:00:00Z |     -2 |
// | t1  | 2021-01-01T00:00:10Z |     10 |
// | t1  | 2021-01-01T00:00:20Z |      7 |
// | t1  | 2021-01-01T00:00:30Z |     17 |
// | t1  | 2021-01-01T00:00:40Z |     15 |
// | t1  | 2021-01-01T00:00:50Z |      4 |

// | tag | _time                | _value |
// | :-: | :------------------- | -----: |
// | t2  | 2021-01-01T00:00:00Z |     19 |
// | t2  | 2021-01-01T00:00:10Z |      4 |
// | t2  | 2021-01-01T00:00:20Z |     -3 |
// | t2  | 2021-01-01T00:00:30Z |     19 |
// | t2  | 2021-01-01T00:00:40Z |     13 |
// | t2  | 2021-01-01T00:00:50Z |      1 |
//
int = (includeNull=false) => {
    _csvData = _numeric(includeNull:includeNull)

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
// import "sampledata"
//
// sampledata.uint()
// ```
// 
// ## Output data
// 
// | tag | _time                |               _value |
// | :-: | :------------------- | -------------------: |
// | t1  | 2021-01-01T00:00:00Z | 18446744073709551614 |
// | t1  | 2021-01-01T00:00:10Z |                   10 |
// | t1  | 2021-01-01T00:00:20Z |                    7 |
// | t1  | 2021-01-01T00:00:30Z |                   17 |
// | t1  | 2021-01-01T00:00:40Z |                   15 |
// | t1  | 2021-01-01T00:00:50Z |                    4 |
// 
// | tag | _time                |               _value |
// | :-: | :------------------- | -------------------: |
// | t2  | 2021-01-01T00:00:00Z |                   19 |
// | t2  | 2021-01-01T00:00:10Z |                    4 |
// | t2  | 2021-01-01T00:00:20Z | 18446744073709551613 |
// | t2  | 2021-01-01T00:00:30Z |                   19 |
// | t2  | 2021-01-01T00:00:40Z |                   13 |
// | t2  | 2021-01-01T00:00:50Z |                    1 |
//
uint = (includeNull=false) => {
    _csvData = _numeric(includeNull:includeNull)

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
// import "sampledata"
//
// sampledata.string()
// ```
//
// ## Output data
// 
// | tag | _time                |      _value |
// | :-- | :------------------- | ----------: |
// | t1  | 2021-01-01T00:00:00Z | smpl_g9qczs |
// | t1  | 2021-01-01T00:00:10Z | smpl_0mgv9n |
// | t1  | 2021-01-01T00:00:20Z | smpl_phw664 |
// | t1  | 2021-01-01T00:00:30Z | smpl_guvzy4 |
// | t1  | 2021-01-01T00:00:40Z | smpl_5v3cce |
// | t1  | 2021-01-01T00:00:50Z | smpl_s9fmgy |
// 
// | tag | _time                |      _value |
// | :-- | :------------------- | ----------: |
// | t2  | 2021-01-01T00:00:00Z | smpl_b5eida |
// | t2  | 2021-01-01T00:00:10Z | smpl_eu4oxp |
// | t2  | 2021-01-01T00:00:20Z | smpl_5g7tz4 |
// | t2  | 2021-01-01T00:00:30Z | smpl_sox1ut |
// | t2  | 2021-01-01T00:00:40Z | smpl_wfm757 |
// | t2  | 2021-01-01T00:00:50Z | smpl_dtn2bv |
// 
string = (includeNull=false) => {
    _csvData = _string(includeNull:includeNull)

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
// import "sampledata"
//
// sampledata.bool()
// ```
//
// ## Output data
// 
// | tag | _time                | _value |
// | :-- | :------------------- | -----: |
// | t1  | 2021-01-01T00:00:00Z |   true |
// | t1  | 2021-01-01T00:00:10Z |   true |
// | t1  | 2021-01-01T00:00:20Z |  false |
// | t1  | 2021-01-01T00:00:30Z |   true |
// | t1  | 2021-01-01T00:00:40Z |  false |
// | t1  | 2021-01-01T00:00:50Z |  false |
// 
// | tag | _time                | _value |
// | :-- | :------------------- | -----: |
// | t2  | 2021-01-01T00:00:00Z |  false |
// | t2  | 2021-01-01T00:00:10Z |   true |
// | t2  | 2021-01-01T00:00:20Z |  false |
// | t2  | 2021-01-01T00:00:30Z |   true |
// | t2  | 2021-01-01T00:00:40Z |   true |
// | t2  | 2021-01-01T00:00:50Z |  false |
// 
bool = (includeNull=false) => {
    _csvData = _bool(includeNull:includeNull)

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
// import "sampledata"
//
// sampledata.numericString()
// ```
// 
// ## Output data
// 
// | tag | _time                | _value (string) |
// | :-: | :------------------- | --------------: |
// | t1  | 2021-01-01T00:00:00Z |              -2 |
// | t1  | 2021-01-01T00:00:10Z |              10 |
// | t1  | 2021-01-01T00:00:20Z |               7 |
// | t1  | 2021-01-01T00:00:30Z |              17 |
// | t1  | 2021-01-01T00:00:40Z |              15 |
// | t1  | 2021-01-01T00:00:50Z |               4 |
// 
// | tag | _time                | _value (string) |
// | :-: | :------------------- | --------------: |
// | t2  | 2021-01-01T00:00:00Z |              19 |
// | t2  | 2021-01-01T00:00:10Z |               4 |
// | t2  | 2021-01-01T00:00:20Z |              -3 |
// | t2  | 2021-01-01T00:00:30Z |              19 |
// | t2  | 2021-01-01T00:00:40Z |              13 |
// | t2  | 2021-01-01T00:00:50Z |               1 |
//
numericString = (includeNull=false) => {
    _csvData = _numeric(includeNull:includeNull)

    return csv.from(csv: _csvData) |> toInt() |> toString()
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
// import "sampledata"
//
// sampledata.numericBool()
// ```
//
// ## Output data
// 
// | tag | _time                | _value |
// | :-- | :------------------- | -----: |
// | t1  | 2021-01-01T00:00:00Z |      1 |
// | t1  | 2021-01-01T00:00:10Z |      1 |
// | t1  | 2021-01-01T00:00:20Z |      0 |
// | t1  | 2021-01-01T00:00:30Z |      1 |
// | t1  | 2021-01-01T00:00:40Z |      0 |
// | t1  | 2021-01-01T00:00:50Z |      0 |
// 
// | tag | _time                | _value |
// | :-- | :------------------- | -----: |
// | t2  | 2021-01-01T00:00:00Z |      0 |
// | t2  | 2021-01-01T00:00:10Z |      1 |
// | t2  | 2021-01-01T00:00:20Z |      0 |
// | t2  | 2021-01-01T00:00:30Z |      1 |
// | t2  | 2021-01-01T00:00:40Z |      1 |
// | t2  | 2021-01-01T00:00:50Z |      0 |
// 
numericBool = (includeNull=false) => {
    _csvData = _bool(includeNull:includeNull)

    return csv.from(csv: _csvData) |> toInt()
}
