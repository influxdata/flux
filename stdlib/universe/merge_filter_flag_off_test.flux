package universe_test

import "testing"
import "planner"

option planner.disableLogicalRules = ["MergeFiltersRule"]

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
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,system,host.local,load4,1.77
,,0,2018-05-22T19:53:46Z,system,host.local,load4,1.77
"

merge_filter_fn = (tables=<-) => tables
    |> filter(fn: (r) => r["_value"] == 1.77)
    |> filter(fn: (r) => r["_field"] == "load4")

test merge_filter_evaluate = () =>
    ({input: testing.loadStorage(csv: input), want: testing.loadMem(csv: output), fn: merge_filter_fn})