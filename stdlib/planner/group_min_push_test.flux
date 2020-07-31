package planner_test

import "testing"


option now = () => (2030-01-01T00:00:00Z)

input = "
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

output = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,double
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_start,_stop,host,_field,_value
,,0,2018-05-22T19:00:00Z,2030-01-01T00:00:00Z,hostA,load1,1.63
,,1,2018-05-22T19:00:00Z,2030-01-01T00:00:00Z,hostB,load3,1.96
,,2,2018-05-22T19:00:00Z,2030-01-01T00:00:00Z,hostC,load1,1.89
"

group_min_fn = (tables=<-) => tables
    |> range(start: 2018-05-22T19:00:00Z)
    |> group(columns:["_start", "_stop","host"])
    |> min()
    |> drop(columns: ["_measurement", "_time"])

test group_min_pushdown = () =>
	({
		input: testing.loadStorage(csv: input),
		want: testing.loadMem(csv: output),
		fn: group_min_fn
	})
