package planner_test

import "testing"
import "planner"

option now = () => (2030-01-01T00:00:00Z)
option planner.disablePhysicalRules = ["PushDownBareAggregateRule"]

input = "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,system,host.local,load1,2.00
,,0,2018-05-22T19:53:36Z,system,host.local,load1,3.00
,,0,2018-05-22T19:53:37Z,system,host.local,load1,2.00
,,0,2018-05-22T19:53:56Z,system,host.local,load1,3.00
,,0,2018-05-22T19:54:06Z,system,host.local,load1,2.00
,,0,2018-05-22T19:54:16Z,system,host.local,load1,3.00

,,1,2018-05-22T19:53:26Z,system,host.local,load3,3.00
,,1,2018-05-22T19:53:36Z,system,host.local,load3,4.00
,,1,2018-05-22T19:53:46Z,system,host.local,load3,3.00
,,1,2018-05-22T19:53:56Z,system,host.local,load3,4.00
,,1,2018-05-22T19:54:06Z,system,host.local,load3,3.00
,,1,2018-05-22T19:54:16Z,system,host.local,load3,4.00

,,2,2018-05-22T19:53:26Z,system,host.local,load5,4.00
,,2,2018-05-22T19:53:36Z,system,host.local,load5,5.00
,,2,2018-05-22T19:53:41Z,system,host.local,load5,4.00
,,2,2018-05-22T19:53:46Z,system,host.local,load5,5.00
,,2,2018-05-22T19:53:56Z,system,host.local,load5,4.00
,,2,2018-05-22T19:54:16Z,system,host.local,load5,5.00
"

output = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double
#group,false,false,true,true,true,true,true,false
#default,_result,,,,,,,
,result,table,_start,_stop,_measurement,host,_field,_value
,,0,2018-05-01T00:00:00Z,2030-01-01T00:00:00Z,system,host.local,load1,2.5
,,1,2018-05-01T00:00:00Z,2030-01-01T00:00:00Z,system,host.local,load3,3.5
,,2,2018-05-01T00:00:00Z,2030-01-01T00:00:00Z,system,host.local,load5,4.5
"

bare_mean_fn = (tables=<-) => tables
    |> range(start: 2018-05-01T00:00:00Z)
    |> mean()

test bare_mean_evaluate = () =>
    ({input: testing.loadStorage(csv: input), want: testing.loadMem(csv: output), fn: bare_mean_fn})
