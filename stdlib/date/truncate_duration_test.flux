package date_test


import "csv"
import "testing"
import "date"
import "timezone"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,string,string,long
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:53:00.000000000Z,_m,FF,-3600000000000
,,0,2018-05-22T19:53:10.000000000Z,_m,FF,-7200000000000
,,0,2018-05-22T19:53:20.000000000Z,_m,FF,-10800000000000
,,0,2018-05-22T19:53:30.000000000Z,_m,FF,0
,,0,2018-05-22T19:53:40.000000000Z,_m,FF,3600000000000
,,0,2018-05-22T19:53:50.000000000Z,_m,FF,7200000000000
,,0,2018-05-22T19:54:00.000000000Z,_m,FF,10800000000000
"

testcase truncate_duration {
    got =
        csv.from(csv: inData)
            |> range(start: 2018-05-22T19:53:00Z)
            |> drop(columns: ["_start", "_stop"])
            |> map(fn: (r) => ({r with _value: date.truncate(t: duration(v: r._value), unit: 1s)}))

    want =
        csv.from(
            csv:
                "
#datatype,string,long,string,string,dateTime:RFC3339,dateTime:RFC3339
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_field,_measurement,_time,_value
,,0,FF,_m,2018-05-22T19:53:00.000000000Z,2029-12-31T23:00:00.000000000Z
,,0,FF,_m,2018-05-22T19:53:10.000000000Z,2029-12-31T22:00:00.000000000Z
,,0,FF,_m,2018-05-22T19:53:20.000000000Z,2029-12-31T21:00:00.000000000Z
,,0,FF,_m,2018-05-22T19:53:30.000000000Z,2030-01-01T00:00:00.000000000Z
,,0,FF,_m,2018-05-22T19:53:40.000000000Z,2030-01-01T01:00:00.000000000Z
,,0,FF,_m,2018-05-22T19:53:50.000000000Z,2030-01-01T02:00:00.000000000Z
,,0,FF,_m,2018-05-22T19:54:00.000000000Z,2030-01-01T03:00:00.000000000Z
",
        )

    testing.diff(got: got, want: want)
}

testcase truncate_duration_location {
    option location = timezone.location(name: "Europe/Madrid")

    got =
        csv.from(csv: inData)
            |> range(start: 2018-05-22T19:53:00Z)
            |> drop(columns: ["_start", "_stop"])
            |> map(fn: (r) => ({r with _value: date.truncate(t: duration(v: r._value), unit: 1mo)}))

    want =
        csv.from(
            csv:
                "
#datatype,string,long,string,string,dateTime:RFC3339,dateTime:RFC3339
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_field,_measurement,_time,_value
,,0,FF,_m,2018-05-22T19:53:00.000000000Z,2029-12-31T23:00:00.000000000Z
,,0,FF,_m,2018-05-22T19:53:10.000000000Z,2029-11-30T23:00:00.000000000Z
,,0,FF,_m,2018-05-22T19:53:20.000000000Z,2029-11-30T23:00:00.000000000Z
,,0,FF,_m,2018-05-22T19:53:30.000000000Z,2029-12-31T23:00:00.000000000Z
,,0,FF,_m,2018-05-22T19:53:40.000000000Z,2029-12-31T23:00:00.000000000Z
,,0,FF,_m,2018-05-22T19:53:50.000000000Z,2029-12-31T23:00:00.000000000Z
,,0,FF,_m,2018-05-22T19:54:00.000000000Z,2029-12-31T23:00:00.000000000Z
",
        )

    testing.diff(got: got, want: want)
}
