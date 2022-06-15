package universe_test


import "csv"
import "testing"

option now = () => 2030-01-01T00:00:00Z

testcase sort_with_null_columns {
    got =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,1.83,load1,system,host.local
,,0,2018-05-22T19:53:36Z,1.7,load1,system,host.local
,,0,2018-05-22T19:53:46Z,1.74,load1,system,host.local
,,0,2018-05-22T19:53:56Z,1.63,load1,system,host.local
,,0,2018-05-22T19:54:06Z,1.91,load1,system,host.local
,,0,2018-05-22T19:54:16Z,1.84,load1,system,host.local
",
        )
            |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:58:00Z)
            |> aggregateWindow(every: 5m, fn: mean)
            |> group(mode: "except", columns: ["_time", "_value"])
            |> sort(columns: ["_value"])

    want =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
#group,false,false,true,true,false,false,true,true,true
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:00Z,2018-05-22T19:58:00Z,2018-05-22T19:58:00Z,,load1,system,host.local
,,0,2018-05-22T19:53:00Z,2018-05-22T19:58:00Z,2018-05-22T19:55:00Z,1.775,load1,system,host.local
",
        )

    testing.diff(got: got, want: want)
}
