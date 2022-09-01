package universe_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

testcase fill_bool {
    inData =
        "
#datatype,string,long,string,string,string,dateTime:RFC3339,boolean
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m1,f1,server01,2018-12-19T22:13:30Z,false
,,0,m1,f1,server01,2018-12-19T22:13:40Z,true
,,0,m1,f1,server01,2018-12-19T22:13:50Z,false
,,0,m1,f1,server01,2018-12-19T22:14:00Z,false
,,0,m1,f1,server01,2018-12-19T22:14:10Z,
,,0,m1,f1,server01,2018-12-19T22:14:20Z,true
,,1,m1,f1,server02,2018-12-19T22:13:30Z,false
,,1,m1,f1,server02,2018-12-19T22:13:40Z,true
,,1,m1,f1,server02,2018-12-19T22:13:50Z,
,,1,m1,f1,server02,2018-12-19T22:14:00Z,true
,,1,m1,f1,server02,2018-12-19T22:14:10Z,true
,,1,m1,f1,server02,2018-12-19T22:14:20Z,
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,boolean
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:30Z,false
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:40Z,true
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:50Z,false
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:00Z,false
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:10Z,false
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:20Z,true
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:30Z,false
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:40Z,true
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:50Z,false
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:00Z,true
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:10Z,true
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:20Z,false
"

    got =
        csv.from(csv: inData)
            |> range(start: 2018-12-15T00:00:00Z)
            |> fill(value: false)
    want = csv.from(csv: outData)

    testing.diff(got, want)
}

testcase fill_int {
    inData =
        "
#datatype,string,long,string,string,string,dateTime:RFC3339,long
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m1,f1,server01,2018-12-19T22:13:30Z,
,,0,m1,f1,server01,2018-12-19T22:13:40Z,-25
,,0,m1,f1,server01,2018-12-19T22:13:50Z,46
,,0,m1,f1,server01,2018-12-19T22:14:00Z,-2
,,0,m1,f1,server01,2018-12-19T22:14:10Z,
,,0,m1,f1,server01,2018-12-19T22:14:20Z,-53
,,1,m1,f1,server02,2018-12-19T22:13:30Z,17
,,1,m1,f1,server02,2018-12-19T22:13:40Z,-44
,,1,m1,f1,server02,2018-12-19T22:13:50Z,-99
,,1,m1,f1,server02,2018-12-19T22:14:00Z,-85
,,1,m1,f1,server02,2018-12-19T22:14:10Z,
,,1,m1,f1,server02,2018-12-19T22:14:20Z,99
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,long
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:30Z,-1
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:40Z,-25
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:50Z,46
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:00Z,-2
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:10Z,-1
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:20Z,-53
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:30Z,17
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:40Z,-44
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:50Z,-99
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:00Z,-85
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:10Z,-1
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:20Z,99
"

    got =
        csv.from(csv: inData)
            |> range(start: 2018-12-15T00:00:00Z)
            |> fill(column: "_value", value: -1)
    want = csv.from(csv: outData)

    testing.diff(got, want)
}

testcase fill_time {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,true,true,false,false
#default,,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:30Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:40Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:50Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:10Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:20Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:40Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:50Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:00Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:10Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:20Z,A
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:30Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:40Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:50Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2077-12-19T22:14:00Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:10Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:20Z,A

,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2077-12-19T22:14:00Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:40Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:50Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:00Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:10Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:20Z,A
"

    got =
        csv.from(csv: inData)
            |> fill(column: "_time", value: 2077-12-19T22:14:00Z)
    want = csv.from(csv: outData)

    testing.diff(got, want)
}

testcase fill_uint {
    inData =
        "
#datatype,string,long,string,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m1,f1,server01,2018-12-19T22:13:30Z,84
,,0,m1,f1,server01,2018-12-19T22:13:40Z,52
,,0,m1,f1,server01,2018-12-19T22:13:50Z,
,,0,m1,f1,server01,2018-12-19T22:14:00Z,62
,,0,m1,f1,server01,2018-12-19T22:14:10Z,22
,,0,m1,f1,server01,2018-12-19T22:14:20Z,78
,,1,m1,f1,server02,2018-12-19T22:13:30Z,
,,1,m1,f1,server02,2018-12-19T22:13:40Z,33
,,1,m1,f1,server02,2018-12-19T22:13:50Z,97
,,1,m1,f1,server02,2018-12-19T22:14:00Z,90
,,1,m1,f1,server02,2018-12-19T22:14:10Z,96
,,1,m1,f1,server02,2018-12-19T22:14:20Z,
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:30Z,84
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:40Z,52
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:50Z,0
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:00Z,62
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:10Z,22
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:20Z,78
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:30Z,0
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:40Z,33
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:50Z,97
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:00Z,90
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:10Z,96
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:20Z,0
"
    got =
        csv.from(csv: inData)
            |> range(start: 2018-12-15T00:00:00Z)
            |> fill(column: "_value", value: uint(v: 0))
    want = csv.from(csv: outData)

    testing.diff(got, want)
}

testcase fill_float {
    inData =
        "
#datatype,string,long,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m1,f1,server01,2018-12-19T22:13:30Z,-61.68790887989735
,,0,m1,f1,server01,2018-12-19T22:13:40Z,-6.3173755351186465
,,0,m1,f1,server01,2018-12-19T22:13:50Z,
,,0,m1,f1,server01,2018-12-19T22:14:00Z,114.285955884979
,,0,m1,f1,server01,2018-12-19T22:14:10Z,16.140262630578995
,,0,m1,f1,server01,2018-12-19T22:14:20Z,29.50336437998469
,,1,m1,f1,server02,2018-12-19T22:13:30Z,
,,1,m1,f1,server02,2018-12-19T22:13:40Z,49.460104214779086
,,1,m1,f1,server02,2018-12-19T22:13:50Z,-36.564150808873954
,,1,m1,f1,server02,2018-12-19T22:14:00Z,34.319039251798635
,,1,m1,f1,server02,2018-12-19T22:14:10Z,
,,1,m1,f1,server02,2018-12-19T22:14:20Z,41.91029522104053
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:30Z,-61.68790887989735
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:40Z,-6.3173755351186465
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:50Z,0.01
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:00Z,114.285955884979
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:10Z,16.140262630578995
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:20Z,29.50336437998469
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:30Z,0.01
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:40Z,49.460104214779086
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:50Z,-36.564150808873954
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:00Z,34.319039251798635
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:10Z,0.01
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:20Z,41.91029522104053
"

    got =
        csv.from(csv: inData)
            |> range(start: 2018-12-15T00:00:00Z)
            |> fill(value: 0.01)
    want = csv.from(csv: outData)

    testing.diff(got, want)
}

testcase fill_string {
    inData =
        "
#datatype,string,long,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m1,f1,server01,2018-12-19T22:13:30Z,
,,0,m1,f1,server01,2018-12-19T22:13:40Z,
,,0,m1,f1,server01,2018-12-19T22:13:50Z,
,,0,m1,f1,server01,2018-12-19T22:14:00Z,
,,0,m1,f1,server01,2018-12-19T22:14:10Z,
,,0,m1,f1,server01,2018-12-19T22:14:20Z,
,,1,m1,f1,server02,2018-12-19T22:13:30Z,
,,1,m1,f1,server02,2018-12-19T22:13:40Z,
,,1,m1,f1,server02,2018-12-19T22:13:50Z,
,,1,m1,f1,server02,2018-12-19T22:14:00Z,
,,1,m1,f1,server02,2018-12-19T22:14:10Z,
,,1,m1,f1,server02,2018-12-19T22:14:20Z,
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:30Z,A
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:40Z,A
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:50Z,A
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:00Z,A
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:10Z,A
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:20Z,A
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:30Z,A
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:40Z,A
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:50Z,A
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:00Z,A
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:10Z,A
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:20Z,A
"

    got =
        csv.from(csv: inData)
            |> range(start: 2018-12-15T00:00:00Z)
            |> fill(column: "_value", value: "A")
    want = csv.from(csv: outData)

    testing.diff(got, want)
}

testcase fill_previous {
    inData =
        "
#datatype,string,long,string,string,string,dateTime:RFC3339,long
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m1,f1,server01,2018-12-19T22:13:30Z,
,,0,m1,f1,server01,2018-12-19T22:13:40Z,-25
,,0,m1,f1,server01,2018-12-19T22:13:50Z,46
,,0,m1,f1,server01,2018-12-19T22:14:00Z,-2
,,0,m1,f1,server01,2018-12-19T22:14:10Z,
,,0,m1,f1,server01,2018-12-19T22:14:20Z,-53
,,1,m1,f1,server02,2018-12-19T22:13:30Z,17
,,1,m1,f1,server02,2018-12-19T22:13:40Z,-44
,,1,m1,f1,server02,2018-12-19T22:13:50Z,-99
,,1,m1,f1,server02,2018-12-19T22:14:00Z,-85
,,1,m1,f1,server02,2018-12-19T22:14:10Z,
,,1,m1,f1,server02,2018-12-19T22:14:20Z,99
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,long
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:30Z,
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:40Z,-25
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:50Z,46
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:00Z,-2
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:10Z,-2
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:20Z,-53
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:30Z,17
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:40Z,-44
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:50Z,-99
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:00Z,-85
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:10Z,-85
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:20Z,99
"

    got =
        csv.from(csv: inData)
            |> range(start: 2018-12-15T00:00:00Z)
            |> fill(usePrevious: true)
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
