// Package sampledata provides functions that return basic sample datasets.
//
// ## Metadata
// introduced: 0.128.0
// tags: sample data
//
package sampledata


import "csv"

// start represents the earliest time included in sample datasets.
//
// start should be used used with `range` when `_start` and `_stop` columns are
// required to demonstrate a transformation.
//
start = 2021-01-01T00:00:00Z

// stop represents the latest time included in sample datasets.
//
// stop should be used used with `range` when `_start` and `_stop` columns are
// required to demonstrate a transformation.
//
stop = 2021-01-01T00:01:00Z

// Base numeric dataset used as a boilerplate for other numeric types
_numeric =
    (includeNull=false) =>
        "
#group,false,false,false,true,false
#datatype,string,long,dateTime:RFC3339,string,double
#default,_result,,,,
,result,table,_time,tag,_value
,,0,2021-01-01T00:00:00Z,t1,-2.18
,,0,2021-01-01T00:00:10Z,t1,"
            +
            (if includeNull then "" else "10.92")
            +
            "
,,0,2021-01-01T00:00:20Z,t1,7.35
,,0,2021-01-01T00:00:30Z,t1," + (if includeNull then
                ""
            else
                "17.53") + "
,,0,2021-01-01T00:00:40Z,t1," + (if includeNull then
                ""
            else
                "15.23") + "
,,0,2021-01-01T00:00:50Z,t1,4.43
,,1,2021-01-01T00:00:00Z,t2,"
            +
            (if includeNull then
                ""
            else
                "19.85")
            +
            "
,,1,2021-01-01T00:00:10Z,t2,4.97
,,1,2021-01-01T00:00:20Z,t2,-3.75
,,1,2021-01-01T00:00:30Z,t2,19.77
,,1,2021-01-01T00:00:40Z,t2,"
            +
            (if includeNull then "" else "13.86") + "
,,1,2021-01-01T00:00:50Z,t2,1.86
"

// String sample dataset.
_string =
    (includeNull=false) =>
        "
#group,false,false,false,true,false
#datatype,string,long,dateTime:RFC3339,string,string
#default,_result,,,,
,result,table,_time,tag,_value
,,0,2021-01-01T00:00:00Z,t1,smpl_g9qczs
,,0,2021-01-01T00:00:10Z,t1,"
            +
            (if includeNull then "" else "smpl_0mgv9n")
            +
            "
,,0,2021-01-01T00:00:20Z,t1,smpl_phw664
,,0,2021-01-01T00:00:30Z,t1,"
            +
            (if includeNull then
                ""
            else
                "smpl_guvzy4") + "
,,0,2021-01-01T00:00:40Z,t1," + (if includeNull then
                ""
            else
                "smpl_5v3cce")
            +
            "
,,0,2021-01-01T00:00:50Z,t1,smpl_s9fmgy
,,1,2021-01-01T00:00:00Z,t2,"
            +
            (if includeNull then
                ""
            else
                "smpl_b5eida")
            +
            "
,,1,2021-01-01T00:00:10Z,t2,smpl_eu4oxp
,,1,2021-01-01T00:00:20Z,t2,smpl_5g7tz4
,,1,2021-01-01T00:00:30Z,t2,smpl_sox1ut
,,1,2021-01-01T00:00:40Z,t2,"
            +
            (if includeNull then "" else "smpl_wfm757")
            +
            "
,,1,2021-01-01T00:00:50Z,t2,smpl_dtn2bv
"

// Boolean sample dataset.
_bool =
    (includeNull=false) =>
        "#group,false,false,false,true,false
#datatype,string,long,dateTime:RFC3339,string,boolean
#default,_result,,,,
,result,table,_time,tag,_value
,,0,2021-01-01T00:00:00Z,t1,true
,,0,2021-01-01T00:00:10Z,t1,"
            +
            (if includeNull then "" else "true")
            +
            "
,,0,2021-01-01T00:00:20Z,t1,false
,,0,2021-01-01T00:00:30Z,t1," + (if includeNull then
                ""
            else
                "true") + "
,,0,2021-01-01T00:00:40Z,t1," + (if includeNull then
                ""
            else
                "false") + "
,,0,2021-01-01T00:00:50Z,t1,false
,,1,2021-01-01T00:00:00Z,t2,"
            +
            (if includeNull then
                ""
            else
                "false")
            +
            "
,,1,2021-01-01T00:00:10Z,t2,true
,,1,2021-01-01T00:00:20Z,t2,false
,,1,2021-01-01T00:00:30Z,t2,true
,,1,2021-01-01T00:00:40Z,t2,"
            +
            (if includeNull then "" else "true") + "
,,1,2021-01-01T00:00:50Z,t2,false
"

// float returns a sample data set with float values.
//
// ## Parameters
//
// - includeNull: Include null values in the returned dataset.
//   Default is `false`.
//
// ## Examples
//
// ### Output basic sample data with float values
// ```
// import "sampledata"
//
// > sampledata.float()
// ```
//
float = (includeNull=false) => {
    _csvData = _numeric(includeNull: includeNull)

    return csv.from(csv: _csvData)
}

// int returns a sample data set with integer values.
//
// ## Parameters
//
// - includeNull: Include null values in the returned dataset.
//   Default is `false`.
//
// ## Examples
//
// ### Output basic sample data with integer values
// ```
// import "sampledata"
//
// > sampledata.int()
// ```
//
int = (includeNull=false) => {
    _csvData = _numeric(includeNull: includeNull)

    return csv.from(csv: _csvData) |> toInt()
}

// uint returns a sample data set with unsigned integer values.
//
// ## Parameters
//
// - includeNull: Include null values in the returned dataset.
//   Default is `false`.
//
// ## Examples
//
// ### Output basic sample data with unsigned integer values
// ```
// import "sampledata"
//
// > sampledata.uint()
// ```
//
uint = (includeNull=false) => {
    _csvData = _numeric(includeNull: includeNull)

    return csv.from(csv: _csvData) |> toUInt()
}

// string returns a sample data set with string values.
//
// ## Parameters
//
// - includeNull: Include null values in the returned dataset.
//   Default is `false`.
//
// ## Examples
//
// ### Output basic sample data with string values
// ```
// import "sampledata"
//
// > sampledata.string()
// ```
//
string = (includeNull=false) => {
    _csvData = _string(includeNull: includeNull)

    return csv.from(csv: _csvData)
}

// bool returns a sample data set with boolean values.
//
// ## Parameters
//
// - includeNull: Include null values in the returned dataset.
//   Default is `false`.
//
// ## Examples
//
// ### Output basic sample data with boolean values
// ```
// import "sampledata"
//
// > sampledata.bool()
// ```
//
bool = (includeNull=false) => {
    _csvData = _bool(includeNull: includeNull)

    return csv.from(csv: _csvData)
}

// numericBool returns a sample data set with numeric (integer) boolean values.
//
// ## Parameters
//
// - includeNull: Include null values in the returned dataset.
//   Default is `false`.
//
// ## Examples
//
// ### Output basic sample data with numeric boolean values
// ```
// import "sampledata"
//
// > sampledata.numericBool()
// ```
//
numericBool = (includeNull=false) => {
    _csvData = _bool(includeNull: includeNull)

    return csv.from(csv: _csvData) |> toInt()
}
