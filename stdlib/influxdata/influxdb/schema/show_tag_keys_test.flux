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

#datatype,string,long,dateTime:RFC3339,string,string,string,string,long
#group,false,false,false,true,true,true,true,false
#default,_result,,,,,,,
,result,table,_time,_measurement,host,region,_field,_value
,,0,2018-05-22T19:53:26Z,system,us-east,host.local,load1,10
,,0,2018-05-22T19:53:36Z,system,us-east,host.local,load1,11
,,0,2018-05-22T19:53:46Z,system,us-east,host.local,load1,18
,,0,2018-05-22T19:53:56Z,system,us-east,host.local,load1,19
,,0,2018-05-22T19:54:06Z,system,us-east,host.local,load1,17
,,0,2018-05-22T19:54:16Z,system,us-east,host.local,load1,17

,,1,2018-05-22T19:53:26Z,system,us-east,host.local,load3,16
,,1,2018-05-22T19:53:36Z,system,us-east,host.local,load3,16
,,1,2018-05-22T19:53:46Z,system,us-east,host.local,load3,15
,,1,2018-05-22T19:53:56Z,system,us-east,host.local,load3,19
,,1,2018-05-22T19:54:06Z,system,us-east,host.local,load3,19
,,1,2018-05-22T19:54:16Z,system,us-east,host.local,load3,19

,,2,2018-05-22T19:53:26Z,system,us-west,host.local,load5,19
,,2,2018-05-22T19:53:36Z,system,us-west,host.local,load5,22
,,2,2018-05-22T19:53:46Z,system,us-west,host.local,load5,11
,,2,2018-05-22T19:53:56Z,system,us-west,host.local,load5,12
,,2,2018-05-22T19:54:06Z,system,us-west,host.local,load5,13
,,2,2018-05-22T19:54:16Z,system,us-west,host.local,load5,13
"

output = "
#datatype,string,long,string
#group,false,false,false
#default,0,,
,result,table,_value
,,0,_field
,,0,_measurement
,,0,_start
,,0,_stop
,,0,host
,,0,region
"

show_tag_keys_fn = (tables=<-) => tables
    |> range(start: 2018-01-01T00:00:00Z, stop: 2019-01-01T00:00:00Z)
    |> filter(fn: (r) => true)
    |> keys()
    |> keep(columns: ["_value"])
    |> distinct()
    |> sort()

test show_tag_keys = () =>
    ({input: testing.loadStorage(csv: input), want: testing.loadMem(csv: output), fn: show_tag_keys_fn})