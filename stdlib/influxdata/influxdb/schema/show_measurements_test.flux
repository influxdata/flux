package schema_test

import "testing"

input = "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,system,host.local,load1,1.83
,,0,2018-05-22T19:53:36Z,system,host.local,load1,1.72
,,0,2018-05-22T19:53:46Z,system,host.local,load1,1.74
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
,,2,2018-05-22T19:53:46Z,system,host.local,load5,1.92
,,2,2018-05-22T19:53:56Z,system,host.local,load5,1.89
,,2,2018-05-22T19:54:06Z,system,host.local,load5,1.94
,,2,2018-05-22T19:54:16Z,system,host.local,load5,1.93

,,3,2018-05-22T19:53:26Z,swap,host.local,used_percent,82.98
,,3,2018-05-22T19:53:36Z,swap,host.local,used_percent,82.59
,,3,2018-05-22T19:53:46Z,swap,host.local,used_percent,82.59
,,3,2018-05-22T19:53:56Z,swap,host.local,used_percent,82.59
,,3,2018-05-22T19:54:06Z,swap,host.local,used_percent,82.59
,,3,2018-05-22T19:54:16Z,swap,host.local,used_percent,82.64
"

output = "
#datatype,string,long,string
#group,false,false,false
#default,0,,
,result,table,_value
,,0,swap
,,0,system
"

show_measurements_fn = (tables=<-) => tables
    |> range(start: 2018-01-01T00:00:00Z, stop: 2019-01-01T00:00:00Z)
    |> filter(fn: (r) => true)
    |> keep(columns: ["_measurement"])
    |> group()
    |> distinct(column: "_measurement")
    |> sort()

test show_measurements = () =>
    ({input: testing.loadStorage(csv: input), want: testing.loadMem(csv: output), fn: show_measurements_fn})