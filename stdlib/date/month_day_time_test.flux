package date_test


import "csv"
import "testing"
import "date"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-01T04:53:00Z,_m,FF,1
,,0,2018-05-02T04:53:10Z,_m,FF,1
,,0,2018-05-03T04:53:20Z,_m,FF,1
,,0,2018-05-04T04:53:30Z,_m,FF,1
,,0,2018-05-05T04:53:40Z,_m,FF,1
,,0,2018-05-06T04:53:50Z,_m,FF,1
,,1,2018-05-07T04:53:00Z,_m,QQ,1
,,1,2018-05-08T04:53:10Z,_m,QQ,1
,,1,2018-05-09T04:53:20Z,_m,QQ,1
,,1,2018-05-10T04:53:30Z,_m,QQ,1
,,1,2018-05-11T04:53:40Z,_m,QQ,1
,,1,2018-05-12T04:53:50Z,_m,QQ,1
,,1,2018-05-13T04:54:00Z,_m,QQ,1
,,1,2018-05-14T04:54:10Z,_m,QQ,1
,,1,2018-05-15T04:54:20Z,_m,QQ,1
,,2,2018-05-16T04:53:00Z,_m,RR,1
,,2,2018-05-17T04:53:10Z,_m,RR,1
,,2,2018-05-18T04:53:20Z,_m,RR,1
,,2,2018-05-19T04:53:30Z,_m,RR,1
,,3,2018-05-20T04:53:40Z,_m,SR,1
,,3,2018-05-21T04:53:50Z,_m,SR,1
,,3,2018-05-22T04:54:00Z,_m,SR,1
,,3,2018-05-23T04:54:00Z,_m,SR,1
,,3,2018-05-24T04:54:00Z,_m,SR,1
,,3,2018-05-25T04:54:00Z,_m,SR,1
,,3,2018-05-26T04:54:00Z,_m,SR,1
,,3,2018-05-27T04:54:00Z,_m,SR,1
,,3,2018-05-28T04:54:00Z,_m,SR,1
,,3,2018-05-29T04:54:00Z,_m,SR,1
,,3,2018-05-30T04:54:00Z,_m,SR,1
,,3,2018-05-31T04:54:00Z,_m,SR,1
"

testcase month_day_time {
        got =
            csv.from(csv: inData)
                |> range(start: 2018-01-01T00:00:00Z)
                |> map(fn: (r) => ({r with _value: date.monthDay(t: r._time)}))

        want =
            csv.from(
                csv:
                    "
#group,false,false,true,true,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,long
#default,_result,,,,,,,
,result,table,_start,_stop,_field,_measurement,_time,_value
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-01T04:53:00Z,1
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-02T04:53:10Z,2
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-03T04:53:20Z,3
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-04T04:53:30Z,4
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-05T04:53:40Z,5
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-06T04:53:50Z,6
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-07T04:53:00Z,7
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-08T04:53:10Z,8
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-09T04:53:20Z,9
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-10T04:53:30Z,10
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-11T04:53:40Z,11
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-12T04:53:50Z,12
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-13T04:54:00Z,13
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-14T04:54:10Z,14
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-15T04:54:20Z,15
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,RR,_m,2018-05-16T04:53:00Z,16
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,RR,_m,2018-05-17T04:53:10Z,17
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,RR,_m,2018-05-18T04:53:20Z,18
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,RR,_m,2018-05-19T04:53:30Z,19
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-20T04:53:40Z,20
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-21T04:53:50Z,21
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-22T04:54:00Z,22
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-23T04:54:00Z,23
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-24T04:54:00Z,24
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-25T04:54:00Z,25
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-26T04:54:00Z,26
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-27T04:54:00Z,27
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-28T04:54:00Z,28
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-29T04:54:00Z,29
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-30T04:54:00Z,30
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-31T04:54:00Z,31
",
            )

        testing.diff(got: got, want: want)
    }

testcase month_day_time_location {
        got =
            csv.from(csv: inData)
                |> range(start: 2018-01-01T00:00:00Z)
                |> map(
                    fn: (r) =>
                        ({r with _value: date.monthDay(t: r._time, location: {zone: "America/Los_Angeles", offset: 0h}),
                        }),
                )

        want =
            csv.from(
                csv:
                    "
#group,false,false,true,true,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,long
#default,_result,,,,,,,
,result,table,_start,_stop,_field,_measurement,_time,_value
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-01T04:53:00Z,30
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-02T04:53:10Z,1
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-03T04:53:20Z,2
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-04T04:53:30Z,3
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-05T04:53:40Z,4
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-06T04:53:50Z,5
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-07T04:53:00Z,6
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-08T04:53:10Z,7
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-09T04:53:20Z,8
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-10T04:53:30Z,9
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-11T04:53:40Z,10
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-12T04:53:50Z,11
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-13T04:54:00Z,12
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-14T04:54:10Z,13
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-05-15T04:54:20Z,14
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,RR,_m,2018-05-16T04:53:00Z,15
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,RR,_m,2018-05-17T04:53:10Z,16
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,RR,_m,2018-05-18T04:53:20Z,17
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,RR,_m,2018-05-19T04:53:30Z,18
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-20T04:53:40Z,19
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-21T04:53:50Z,20
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-22T04:54:00Z,21
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-23T04:54:00Z,22
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-24T04:54:00Z,23
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-25T04:54:00Z,24
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-26T04:54:00Z,25
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-27T04:54:00Z,26
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-28T04:54:00Z,27
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-29T04:54:00Z,28
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-30T04:54:00Z,29
,,3,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SR,_m,2018-05-31T04:54:00Z,30
",
            )

        testing.diff(got: got, want: want)
    }
