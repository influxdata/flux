package date_test


import "csv"
import "testing"
import "date"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
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
outData =
    "

testcase week_duration {
    got = csv.from(csv: inData)
        |> range(start: 2018-01-01T00:00:00Z)
        |> map(fn: (r) => ({r with _value: date.week(t: duration(v: r._value))}))

    want = csv.from(
        csv: "
#group,false,false,true,true,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,long
#default,_result,,,,,,,
,result,table,_start,_stop,_field,_measurement,_time,_value
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:01:00.254819212Z,1
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:02:00.748691723Z,1
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:03:00.947182316Z,1
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:04:00.538816341Z,1
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:05:00.676423456Z,1
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:06:00.982342357Z,1
"
t_duration_week = (table=<-) =>
    table
        |> range(start: 2018-01-01T00:00:00Z)
        |> map(fn: (r) => ({r with _value: date.week(t: duration(v: r._value))}))

test _duration_week = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_duration_week})
",
    )

    testing.diff(got: got, want: want)
}

testcase week_duration_location {
    got = csv.from(csv: inData)
        |> range(start: 2018-01-01T00:00:00Z)
        |> map(fn: (r) => ({r with _value: date.week(t: duration(v: r._value), location: {zone: "Australia/Sydney", offset: -1w})}))

    want = csv.from(
        csv: "
#group,false,false,true,true,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,long
#default,_result,,,,,,,
,result,table,_start,_stop,_field,_measurement,_time,_value
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:01:00.254819212Z,52
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:02:00.748691723Z,52
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:03:00.947182316Z,52
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:04:00.538816341Z,52
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:05:00.676423456Z,52
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:06:00.982342357Z,52
",
    )

    testing.diff(got: got, want: want)
}
