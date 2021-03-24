package planner_test

import "testing"
import "planner"

option now = () => (2030-01-01T00:00:00Z)

input = "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,system,host.local,load1,1.83
,,0,2018-05-22T19:53:36Z,system,host.local,load1,1.63
,,1,2018-05-22T19:53:26Z,system,host.local,load3,1.72
,,2,2018-05-22T19:53:26Z,system,host.local,load4,1.77
,,2,2018-05-22T19:53:36Z,system,host.local,load4,1.78
,,2,2018-05-22T19:53:46Z,system,host.local,load4,1.77
"

testcase group_min_bare {
// todo(faith): remove drop() call once storage doesnt force _start and _stop columns to be in group key
    result = testing.loadStorage(csv: input)
        |> range(start: 2018-05-22T19:53:26Z)
        |> filter(fn: (r) => r["_value"] == 1.77)
        |> group(columns: ["_field"])
        |> min()
        |> drop(columns: ["_start", "_stop"])

out_min_bare = "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,false,false,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,system,host.local,load4,1.77
"

    testing.diff(got: result, want: testing.loadMem(csv: out_min_bare)) |> yield()
}

testcase group_min_window {
    result = testing.loadStorage(csv: input)
        |> range(start: 2018-05-22T19:53:26Z)
        |> filter(fn: (r) => r["_value"] == 1.77)
        |> group(columns: ["_field"])
        |> window(every: 1d)
        |> min()

out_min_window = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double
#group,false,false,true,true,false,false,false,true,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,2018-05-23T00:00:00Z,2018-05-22T19:53:26Z,system,host.local,load4,1.77
"

    testing.diff(got: result, want: testing.loadMem(csv: out_min_window)) |> yield()
}

testcase group_min_agg_window {
    result = testing.loadStorage(csv: input)
        |> range(start: 2018-05-22T19:53:26Z)
        |> group(columns: ["host"])
        |> aggregateWindow(fn: min, every: 1d)

out_min_agg_window = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double
#group,false,false,true,true,false,false,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-23T00:00:00Z,system,host.local,load1,1.63
"

    testing.diff(got: result, want: testing.loadMem(csv: out_min_agg_window)) |> yield()
}

testcase group_min_agg_window_empty {
    result = testing.loadStorage(csv: input)
    |> range(start: 2018-05-22T19:53:26Z)
    |> group(columns: ["_field"])
    |> aggregateWindow(fn: min, every: 1d, createEmpty: true)

    out_min_agg_window_empty = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double
#group,false,false,true,true,false,false,false,true,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-23T00:00:00Z,system,host.local,load1,1.63
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-23T00:00:00Z,system,host.local,load3,1.72
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-23T00:00:00Z,system,host.local,load4,1.77
"

    testing.diff(got: result, want: testing.loadMem(csv: out_min_agg_window_empty)) |> yield()
}
