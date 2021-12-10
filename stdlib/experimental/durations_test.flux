package experimental_test


import "csv"
import "testing"
import "date"
import "experimental"

option now = () => 2018-05-22T20:00:00Z

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,long
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:01:00.254819212Z,_m,FF,-3
,,0,2018-05-22T19:02:00.748691723Z,_m,FF,-2
,,0,2018-05-22T19:03:00.947182316Z,_m,FF,-1
,,0,2018-05-22T19:04:00.538816341Z,_m,FF,0
,,0,2018-05-22T19:05:00.676423456Z,_m,FF,1
,,0,2018-05-22T19:06:00.982342357Z,_m,FF,2
"

testcase add_durations {
    got = csv.from(csv: inData)
        |> range(start: 2018-01-01T00:00:00Z)
        |> map(fn: (r) => ({r with _value: date.minute(t: experimental.addDuration(d: -1m, to: r._time))}))

    want = csv.from(
        csv: "
#group,false,false,true,true,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,long
#default,_result,,,,,,,
,result,table,_start,_stop,_field,_measurement,_time,_value
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:01:00.254819212Z,0
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:02:00.748691723Z,1
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:03:00.947182316Z,2
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:04:00.538816341Z,3
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:05:00.676423456Z,4
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:06:00.982342357Z,5
",
    )

    testing.diff(got: got, want: want)
}

testcase add_durations_location {
    got = csv.from(csv: inData)
        |> range(start: 2018-01-01T00:00:00Z)
        |> map(fn: (r) => ({r with _value: date.minute(t: experimental.addDuration(d: -1m, to: r._time, location: {zone: "Asia/Kolkata", offset: 0h}))}))

    want = csv.from(
        csv: "
#group,false,false,true,true,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,long
#default,_result,,,,,,,
,result,table,_start,_stop,_field,_measurement,_time,_value
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:01:00.254819212Z,30
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:02:00.748691723Z,31
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:03:00.947182316Z,32
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:04:00.538816341Z,33
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:05:00.676423456Z,34
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:06:00.982342357Z,35
",
    )

    testing.diff(got: got, want: want)
}

testcase sub_durations {
    got = csv.from(csv: inData)
        |> range(start: 2018-01-01T00:00:00Z)
        |> map(fn: (r) => ({r with _value: date.minute(t: experimental.subDuration(d: -1m, from: r._time))}))

    want = csv.from(
        csv: "
#group,false,false,true,true,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,long
#default,_result,,,,,,,
,result,table,_start,_stop,_field,_measurement,_time,_value
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:01:00.254819212Z,2
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:02:00.748691723Z,3
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:03:00.947182316Z,4
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:04:00.538816341Z,5
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:05:00.676423456Z,6
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:06:00.982342357Z,7
",
    )

    testing.diff(got: got, want: want)
}

testcase sub_durations_location {
    got = csv.from(csv: inData)
        |> range(start: 2018-01-01T00:00:00Z)
        |> map(fn: (r) => ({r with _value: date.minute(t: experimental.subDuration(d: -1m, from: r._time, location: {zone: "Asia/Kolkata", offset: 0h}))}))

    want = csv.from(
        csv: "
#group,false,false,true,true,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,long
#default,_result,,,,,,,
,result,table,_start,_stop,_field,_measurement,_time,_value
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:01:00.254819212Z,32
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:02:00.748691723Z,33
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:03:00.947182316Z,34
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:04:00.538816341Z,35
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:05:00.676423456Z,36
,,0,2018-01-01T00:00:00Z,2018-05-22T20:00:00Z,FF,_m,2018-05-22T19:06:00.982342357Z,37
",
    )

    testing.diff(got: got, want: want)
}