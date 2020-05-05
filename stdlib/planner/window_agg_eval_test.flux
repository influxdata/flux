package planner_test

import "testing"
import "planner"

option planner.disablePhysicalRules = ["PushDownWindowAggregateRule"]

input = "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,system,host.local,load1,1.83
,,0,2018-05-22T19:53:36Z,system,host.local,load1,1.72
,,0,2018-05-22T19:53:37Z,system,host.local,load1,1.77
,,0,2018-05-22T19:53:56Z,system,host.local,load1,1.63
,,0,2018-05-22T19:54:06Z,system,host.local,load1,1.91
,,0,2018-05-22T19:54:16Z,system,host.local,load1,1.84

,,1,2018-05-22T19:53:26Z,system,host.local,load3,1.98
,,1,2018-05-22T19:53:36Z,system,host.local,load3,1.97
,,1,2018-05-22T19:53:46Z,system,host.local,load3,1.97
,,1,2018-05-22T19:53:56Z,system,host.local,load3,1.96
,,1,2018-05-22T19:54:06Z,system,host.local,load3,1.98
,,1,2018-05-22T19:54:16Z,system,host.local,load3,1.97

,,2,2018-05-22T19:53:26Z,system,host.local,load5,1.95
,,2,2018-05-22T19:53:36Z,system,host.local,load5,1.92
,,2,2018-05-22T19:53:41Z,system,host.local,load5,1.91
,,2,2018-05-22T19:53:46Z,system,host.local,load5,1.92
,,2,2018-05-22T19:53:56Z,system,host.local,load5,1.89
,,2,2018-05-22T19:54:16Z,system,host.local,load5,1.93
"

output = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,long
#group,false,false,true,true,true,true,true,false
#default,_result,,,,,,,
,result,table,_start,_stop,_measurement,host,_field,_value
,,0,2018-05-22T19:53:20Z,2018-05-22T19:53:40Z,system,host.local,load1,3
,,1,2018-05-22T19:53:40Z,2018-05-22T19:54:00Z,system,host.local,load1,1
,,2,2018-05-22T19:54:00Z,2018-05-22T19:54:20Z,system,host.local,load1,2
,,3,2018-05-22T19:53:20Z,2018-05-22T19:53:40Z,system,host.local,load3,2
,,4,2018-05-22T19:53:40Z,2018-05-22T19:54:00Z,system,host.local,load3,2
,,5,2018-05-22T19:54:00Z,2018-05-22T19:54:20Z,system,host.local,load3,2
,,6,2018-05-22T19:53:20Z,2018-05-22T19:53:40Z,system,host.local,load5,2
,,7,2018-05-22T19:53:40Z,2018-05-22T19:54:00Z,system,host.local,load5,3
,,8,2018-05-22T19:54:00Z,2018-05-22T19:54:20Z,system,host.local,load5,1
"

window_count_fn = (tables=<-) => tables
    |> range(start: 0)
    |> window(every: 20s)
    |> count()

test window_count_evaluate = () =>
    ({input: testing.loadStorage(csv: input), want: testing.loadMem(csv: output), fn: window_count_fn})
