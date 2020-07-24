package planner_test

import "testing"
import "planner"

option now = () => (2030-01-01T00:00:00Z)

input = "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,system,host.local,load1,3.00
,,0,2018-05-22T19:53:36Z,system,host.local,load1,4.00
,,0,2018-05-22T19:53:37Z,system,host.local,load1,5.00
,,0,2018-05-22T19:53:56Z,system,host.local,load1,6.00
,,0,2018-05-22T19:54:06Z,system,host.local,load1,7.00
,,0,2018-05-22T19:54:16Z,system,host.local,load1,8.00

,,1,2018-05-22T19:53:26Z,system,host.local,load3,1.55
,,1,2018-05-22T19:53:36Z,system,host.local,load3,1.65
,,1,2018-05-22T19:53:46Z,system,host.local,load3,1.75
,,1,2018-05-22T19:53:56Z,system,host.local,load3,1.85
,,1,2018-05-22T19:54:06Z,system,host.local,load3,1.95
,,1,2018-05-22T19:54:16Z,system,host.local,load3,2.05

,,2,2018-05-22T19:53:26Z,system,host.local,load5,2.25
,,2,2018-05-22T19:53:36Z,system,host.local,load5,2.35
,,2,2018-05-22T19:53:41Z,system,host.local,load5,2.50
,,2,2018-05-22T19:53:46Z,system,host.local,load5,2.00
,,2,2018-05-22T19:53:56Z,system,host.local,load5,4.50
,,2,2018-05-22T19:54:16Z,system,host.local,load5,2.75
"

output = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double
#group,false,false,true,true,true,true,true,false
#default,_result,,,,,,,
,result,table,_start,_stop,_measurement,host,_field,_value
,,0,2018-05-22T19:53:20Z,2018-05-22T19:53:40Z,system,host.local,load1,4.0
,,1,2018-05-22T19:53:40Z,2018-05-22T19:54:00Z,system,host.local,load1,6.0
,,2,2018-05-22T19:54:00Z,2018-05-22T19:54:20Z,system,host.local,load1,7.5
,,3,2018-05-22T19:53:20Z,2018-05-22T19:53:40Z,system,host.local,load3,1.6
,,4,2018-05-22T19:53:40Z,2018-05-22T19:54:00Z,system,host.local,load3,1.8
,,5,2018-05-22T19:54:00Z,2018-05-22T19:54:20Z,system,host.local,load3,2.0
,,6,2018-05-22T19:53:20Z,2018-05-22T19:53:40Z,system,host.local,load5,2.3
,,7,2018-05-22T19:53:40Z,2018-05-22T19:54:00Z,system,host.local,load5,3.0
,,8,2018-05-22T19:54:00Z,2018-05-22T19:54:20Z,system,host.local,load5,2.75
"

window_mean_fn = (tables=<-) => tables
    |> range(start: 2018-05-22T00:00:00Z)
    |> window(every: 20s)
    |> mean()

test window_mean_evaluate = () =>
    ({input: testing.loadStorage(csv: input), want: testing.loadMem(csv: output), fn: window_mean_fn})
