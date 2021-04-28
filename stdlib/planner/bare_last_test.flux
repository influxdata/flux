package planner_test


import "testing"
import "planner"

input = "
#datatype,string,long,dateTime:RFC3339,string,string,long
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2020-10-30T00:00:01Z,m,f,1
,,0,2020-10-30T00:00:09Z,m,f,9
"
output = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string
#group,false,false,true,true,false,false,true,true
#default,_result,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement
,,0,2020-10-30T00:00:01Z,2020-10-30T00:00:09Z,2020-10-30T00:00:01Z,1,f,m
"
bare_last_fn = (tables=<-) => tables
    |> range(start: 2020-10-30T00:00:01Z, stop: 2020-10-30T00:00:09Z)
    |> last()

test bare_last = () => ({input: testing.loadStorage(csv: input), want: testing.loadMem(csv: output), fn: bare_last_fn})
