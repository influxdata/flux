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
,,0,2018-05-22T19:53:00.000000000Z,_m,FF,1
,,0,2018-05-22T19:53:10.000000000Z,_m,FF,1
,,0,2018-05-22T19:53:20.000000000Z,_m,FF,1
,,0,2018-05-22T19:53:30.000000000Z,_m,FF,1
,,0,2018-05-22T19:53:40.000000000Z,_m,FF,1
,,0,2018-05-22T19:53:50.000000000Z,_m,FF,1
,,0,2018-04-22T19:53:50.000000000Z,_m,FF,1
,,0,2018-03-22T19:53:50.000000000Z,_m,FF,1
,,0,2018-02-22T19:53:50.000000000Z,_m,FF,1
,,0,2018-01-22T19:53:50.000000000Z,_m,FF,1
,,0,2017-12-22T19:53:50.000000000Z,_m,FF,1
"

testcase truncate_time {
    got =
        csv.from(csv: inData)
            |> range(start: 2017-12-22T19:53:00Z)
            |> drop(columns: ["_start", "_stop"])
            |> map(fn: (r) => ({r with _time: date.truncate(t: r._time, unit: 1h)}))

    want =
        csv.from(
            csv:
                "
#datatype,string,long,string,string,dateTime:RFC3339,long
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_field,_measurement,_time,_value
,,0,FF,_m,2018-05-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-05-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-05-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-05-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-05-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-05-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-04-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-03-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-02-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-01-22T19:00:00.000000000Z,1
,,0,FF,_m,2017-12-22T19:00:00.000000000Z,1
",
        )

    testing.diff(got: got, want: want)
}

testcase truncate_time_location {
    option location = timezone.location(name: "Europe/Madrid")

    got =
        csv.from(csv: inData)
            |> range(start: 2017-12-22T19:53:00Z)
            |> drop(columns: ["_start", "_stop"])
            |> map(fn: (r) => ({r with _time: date.truncate(t: r._time, unit: 1mo)}))

    want =
        csv.from(
            csv:
                "
#datatype,string,long,string,string,dateTime:RFC3339,long
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_field,_measurement,_time,_value
,,0,FF,_m,2018-04-30T22:00:00.000000000Z,1
,,0,FF,_m,2018-04-30T22:00:00.000000000Z,1
,,0,FF,_m,2018-04-30T22:00:00.000000000Z,1
,,0,FF,_m,2018-04-30T22:00:00.000000000Z,1
,,0,FF,_m,2018-04-30T22:00:00.000000000Z,1
,,0,FF,_m,2018-04-30T22:00:00.000000000Z,1
,,0,FF,_m,2018-03-31T22:00:00.000000000Z,1
,,0,FF,_m,2018-02-28T23:00:00.000000000Z,1
,,0,FF,_m,2018-01-31T23:00:00.000000000Z,1
,,0,FF,_m,2017-12-31T23:00:00.000000000Z,1
,,0,FF,_m,2017-11-30T23:00:00.000000000Z,1
",
        )

    testing.diff(got: got, want: want)
}
