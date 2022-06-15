package planner_test


import "csv"
import "testing"
import "planner"

input =
    "
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
    result =
        csv.from(csv: input)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z, stop: 2018-05-24T00:00:00Z)
            |> filter(fn: (r) => r["_value"] == 1.77)
            |> group(columns: ["_field"])
            |> min()
            |> drop(columns: ["_start", "_stop"])
    out_min_bare =
        "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,false,false,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,system,host.local,load4,1.77
"

    testing.diff(got: result, want: csv.from(csv: out_min_bare)) |> yield()
}

input_host =
    "
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
,,3,2018-05-22T19:53:36Z,system,host.remote,load4,1.78
,,3,2018-05-22T19:53:46Z,system,host.remote,load4,1.77
"

testcase group_min_bare_host {
    // todo(faith): remove drop() call once storage doesnt force _start and _stop columns to be in group key
    result =
        csv.from(csv: input_host)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z, stop: 2018-05-24T00:00:00Z)
            |> filter(fn: (r) => r["host"] == "host.local")
            |> group(columns: ["_field"])
            |> min()
            |> drop(columns: ["_start", "_stop"])
    out_min_bare =
        "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,false,false,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:36Z,system,host.local,load1,1.63
,,1,2018-05-22T19:53:26Z,system,host.local,load3,1.72
,,2,2018-05-22T19:53:26Z,system,host.local,load4,1.77
"

    testing.diff(got: result, want: csv.from(csv: out_min_bare)) |> yield()
}

input_field =
    "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,system,hostA,load1,1.83
,,0,2018-05-22T19:53:36Z,system,hostA,load1,1.72
,,0,2018-05-22T19:53:37Z,system,hostA,load1,1.77
,,0,2018-05-22T19:53:56Z,system,hostA,load1,1.63
,,0,2018-05-22T19:54:06Z,system,hostA,load1,1.91
,,0,2018-05-22T19:54:16Z,system,hostA,load1,1.84

,,1,2018-05-22T19:53:26Z,system,hostB,load3,1.98
,,1,2018-05-22T19:53:36Z,system,hostB,load3,1.97
,,1,2018-05-22T19:53:46Z,system,hostB,load3,1.97
,,1,2018-05-22T19:53:56Z,system,hostB,load3,1.96
,,1,2018-05-22T19:54:06Z,system,hostB,load3,1.98
,,1,2018-05-22T19:54:16Z,system,hostB,load3,1.97

,,2,2018-05-22T19:53:26Z,system,hostC,load5,1.95
,,2,2018-05-22T19:53:36Z,system,hostC,load5,1.92
,,2,2018-05-22T19:53:41Z,system,hostC,load5,1.91

,,3,2018-05-22T19:53:46Z,system,hostC,load1,1.92
,,3,2018-05-22T19:53:56Z,system,hostC,load1,1.89
,,3,2018-05-22T19:54:16Z,system,hostC,load1,1.93
"

testcase group_min_bare_field {
    // todo(faith): remove drop() call once storage doesnt force _start and _stop columns to be in group key
    result =
        csv.from(csv: input_field)
            |> testing.load()
            |> range(start: 2018-05-22T19:00:00Z, stop: 2018-05-24T00:00:00Z)
            |> group(columns: ["_start", "_stop", "host"])
            |> min()
            |> drop(columns: ["_measurement", "_time"])
    out_min_bare =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,double
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_start,_stop,host,_field,_value
,,0,2018-05-22T19:00:00Z,2018-05-24T00:00:00Z,hostA,load1,1.63
,,1,2018-05-22T19:00:00Z,2018-05-24T00:00:00Z,hostB,load3,1.96
,,2,2018-05-22T19:00:00Z,2018-05-24T00:00:00Z,hostC,load1,1.89
"

    testing.diff(got: result, want: csv.from(csv: out_min_bare)) |> yield()
}
testcase group_min_window {
    result =
        csv.from(csv: input)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z, stop: 2018-05-24T00:00:00Z)
            |> filter(fn: (r) => r["_value"] == 1.77)
            |> group(columns: ["_field"])
            |> window(every: 1d)
            |> min()
    out_min_window =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double
#group,false,false,true,true,false,false,false,true,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,2018-05-23T00:00:00Z,2018-05-22T19:53:26Z,system,host.local,load4,1.77
"

    testing.diff(got: result, want: csv.from(csv: out_min_window)) |> yield()
}
testcase group_min_agg_window {
    result =
        csv.from(csv: input)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z, stop: 2018-05-24T00:00:00Z)
            |> group(columns: ["host"])
            |> aggregateWindow(fn: min, every: 1d, createEmpty: false)
    out_min_agg_window =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double
#group,false,false,true,true,false,false,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,2018-05-24T00:00:00Z,2018-05-23T00:00:00Z,system,host.local,load1,1.63
"

    testing.diff(got: result, want: csv.from(csv: out_min_agg_window)) |> yield()
}
testcase group_min_agg_window_empty {
    result =
        csv.from(csv: input)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z, stop: 2018-05-24T00:00:00Z)
            |> group(columns: ["_field"])
            |> aggregateWindow(fn: min, every: 1d, createEmpty: true)
            |> drop(columns: ["_measurement", "host"])
    out_min_agg_window_empty =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,double
#group,false,false,true,true,false,true,false
#default,_result,,,,,,
,result,table,_start,_stop,_time,_field,_value
,,0,2018-05-22T19:53:26Z,2018-05-24T00:00:00Z,2018-05-23T00:00:00Z,load1,1.63
,,0,2018-05-22T19:53:26Z,2018-05-24T00:00:00Z,2018-05-24T00:00:00Z,load1,
,,1,2018-05-22T19:53:26Z,2018-05-24T00:00:00Z,2018-05-23T00:00:00Z,load3,1.72
,,1,2018-05-22T19:53:26Z,2018-05-24T00:00:00Z,2018-05-24T00:00:00Z,load3,
,,2,2018-05-22T19:53:26Z,2018-05-24T00:00:00Z,2018-05-23T00:00:00Z,load4,1.77
,,2,2018-05-22T19:53:26Z,2018-05-24T00:00:00Z,2018-05-24T00:00:00Z,load4,
"

    testing.diff(got: result, want: csv.from(csv: out_min_agg_window_empty)) |> yield()
}
