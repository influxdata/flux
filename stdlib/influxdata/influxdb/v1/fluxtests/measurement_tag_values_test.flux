package v1_test

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

,,1,2018-05-22T19:53:26Z,sys,host.local,load3,1.98
,,1,2018-05-22T19:53:36Z,sys,host.local,load3,1.97
,,1,2018-05-22T19:53:46Z,sys,host.local,load3,1.97
,,1,2018-05-22T19:53:56Z,sys,host.local,load3,1.96
,,1,2018-05-22T19:54:06Z,sys,host.local,load3,1.98
,,1,2018-05-22T19:54:16Z,sys,host.local,load3,1.97

,,2,2018-05-22T19:53:26Z,system,host.local,load5,1.95
,,2,2018-05-22T19:53:36Z,system,host.local,load5,1.92
,,2,2018-05-22T19:53:46Z,system,host.local,load5,1.92
,,2,2018-05-22T19:53:56Z,system,host.local,load5,1.89
,,2,2018-05-22T19:54:06Z,system,host.local,load5,1.94
,,2,2018-05-22T19:54:16Z,system,host.local,load5,1.93

,,3,2018-05-22T19:53:26Z,swap,host.global,used_percent,82.98
,,3,2018-05-22T19:53:36Z,swap,host.global,used_percent,82.59
,,3,2018-05-22T19:53:46Z,swap,host.global,used_percent,82.59
,,3,2018-05-22T19:53:56Z,swap,host.global,used_percent,82.59
,,3,2018-05-22T19:54:06Z,swap,host.global,used_percent,82.59
,,3,2018-05-22T19:54:16Z,swap,host.global,used_percent,82.64

#datatype,string,long,dateTime:RFC3339,string,string,string,long
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,sys,host.global,load7,183
,,0,2018-05-22T19:53:36Z,sys,host.global,load7,172
,,0,2018-05-22T19:53:46Z,sys,host.global,load7,174
,,0,2018-05-22T19:53:56Z,sys,host.global,load7,163
,,0,2018-05-22T19:54:06Z,sys,host.global,load7,191
,,0,2018-05-22T19:54:16Z,sys,host.global,load7,184

,,1,2018-05-22T19:53:26Z,sys,host.local,load8,198
,,1,2018-05-22T19:53:36Z,sys,host.local,load8,197
,,1,2018-05-22T19:53:46Z,sys,host.local,load8,197
,,1,2018-05-22T19:53:56Z,sys,host.local,load8,196
,,1,2018-05-22T19:54:06Z,sys,host.local,load8,198
,,1,2018-05-22T19:54:16Z,sys,host.local,load8,197

,,2,2018-05-22T19:53:26Z,sys,host.global,load9,195
,,2,2018-05-22T19:53:36Z,sys,host.global,load9,192
,,2,2018-05-22T19:53:46Z,sys,host.global,load9,192
,,2,2018-05-22T19:53:56Z,sys,host.global,load9,189
,,2,2018-05-22T19:54:06Z,sys,host.global,load9,194
,,2,2018-05-22T19:54:16Z,sys,host.global,load9,193

,,3,2018-05-22T19:53:26Z,swp,host.global,used_percent,8298
,,3,2018-05-22T19:53:36Z,swp,host.global,used_percent,8259
,,3,2018-05-22T19:53:46Z,swp,host.global,used_percent,8259
,,3,2018-05-22T19:53:56Z,swp,host.global,used_percent,8259
,,3,2018-05-22T19:54:06Z,swp,host.global,used_percent,8259
,,3,2018-05-22T19:54:16Z,swp,host.global,used_percent,8264
"

output = "
#datatype,string,long,string
#group,false,false,false
#default,0,,
,result,table,_value
,,0,load3
,,0,load8
"

measurement_tag_values_fn = (tables=<-) => tables
    |> range(start: 2018-01-01T00:00:00Z, stop: 2019-01-01T00:00:00Z)
    |> filter(fn: (r) => r._measurement == "sys")
    |> filter(fn: (r) => r.host == "host.local")
    |> keep(columns: ["_field"])
    |> group()
    |> distinct(column: "_field")
    |> sort()

test measurement_tag_values = () =>
    ({input: testing.loadStorage(csv: input), want: testing.loadMem(csv: output), fn: measurement_tag_values_fn})