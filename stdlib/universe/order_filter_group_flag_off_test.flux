package universe_test

import "testing"
import "planner"

option now = () => (2030-01-01T00:00:00Z)
option planner.disableLogicalRules = ["GroupFilterTransposeRule"]

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

output = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double
#group,false,false,false,false,true,false,false,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,system,host.local,load4,1.77
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,system,host.local,load4,1.78
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,system,host.local,load4,1.77
"

order_fn = (tables=<-) => tables
    |> range(start: 2018-05-22T19:53:26Z)
    |> group(columns: ["_time"])
    |> filter(fn: (r) => r["_field"] == "load4")

test order_evaluate = () =>
    ({input: testing.loadStorage(csv: input), want: testing.loadMem(csv: output), fn: order_fn})