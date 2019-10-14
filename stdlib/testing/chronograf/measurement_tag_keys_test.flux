package chronograf_test

import "testing"

input = "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,host,_field,_value
,,0,2018-05-22T19:53:26Z,sys,host.local,load1,1.83
,,0,2018-05-22T19:53:36Z,sys,host.local,load1,1.72
,,0,2018-05-22T19:53:46Z,sys,host.local,load1,1.74
,,0,2018-05-22T19:53:56Z,sys,host.local,load1,1.63
,,0,2018-05-22T19:54:06Z,sys,host.local,load1,1.91
,,0,2018-05-22T19:54:16Z,sys,host.local,load1,1.84

,,1,2018-05-22T19:53:26Z,sys,host.local,load3,1.98
,,1,2018-05-22T19:53:36Z,sys,host.local,load3,1.97
,,1,2018-05-22T19:53:46Z,sys,host.local,load3,1.97
,,1,2018-05-22T19:53:56Z,sys,host.local,load3,1.96
,,1,2018-05-22T19:54:06Z,sys,host.local,load3,1.98
,,1,2018-05-22T19:54:16Z,sys,host.local,load3,1.97

,,2,2018-05-22T19:53:26Z,sys,host.local,load5,1.95
,,2,2018-05-22T19:53:36Z,sys,host.local,load5,1.92
,,2,2018-05-22T19:53:46Z,sys,host.local,load5,1.92
,,2,2018-05-22T19:53:56Z,sys,host.local,load5,1.89
,,2,2018-05-22T19:54:06Z,sys,host.local,load5,1.94
,,2,2018-05-22T19:54:16Z,sys,host.local,load5,1.93

#datatype,string,long,dateTime:RFC3339,string,string,string,string,long
#group,false,false,false,true,true,true,true,false
#default,_result,,,,,,,
,result,table,_time,_measurement,reg,host,_field,_value
,,0,2018-05-22T19:53:26Z,swp,us-east,host.local,load1,10
,,0,2018-05-22T19:53:36Z,swp,us-east,host.local,load1,11
,,0,2018-05-22T19:53:46Z,swp,us-east,host.local,load1,18
,,0,2018-05-22T19:53:56Z,swp,us-east,host.local,load1,19
,,0,2018-05-22T19:54:06Z,swp,us-east,host.local,load1,17
,,0,2018-05-22T19:54:16Z,swp,us-east,host.local,load1,17

#datatype,string,long,dateTime:RFC3339,string,string,string,string,long
#group,false,false,false,true,true,true,true,false
#default,_result,,,,,,,
,result,table,_time,_measurement,region,host,_field,_value
,,0,2018-05-22T19:53:26Z,swp,us-east,host.global,load1,10
,,0,2018-05-22T19:53:36Z,swp,us-east,host.global,load1,11
,,0,2018-05-22T19:53:46Z,swp,us-east,host.global,load1,18
,,0,2018-05-22T19:53:56Z,swp,us-east,host.global,load1,19
,,0,2018-05-22T19:54:06Z,swp,us-east,host.global,load1,17
,,0,2018-05-22T19:54:16Z,swp,us-east,host.global,load1,17

,,1,2018-05-22T19:53:26Z,swp,us-east,host.global,load3,16
,,1,2018-05-22T19:53:36Z,swp,us-east,host.global,load3,16
,,1,2018-05-22T19:53:46Z,swp,us-east,host.global,load3,15
,,1,2018-05-22T19:53:56Z,swp,us-east,host.global,load3,19
,,1,2018-05-22T19:54:06Z,swp,us-east,host.global,load3,19
,,1,2018-05-22T19:54:16Z,swp,us-east,host.global,load3,19

,,2,2018-05-22T19:53:26Z,swp,us-east,host.global,load5,19
,,2,2018-05-22T19:53:36Z,swp,us-east,host.global,load5,22
,,2,2018-05-22T19:53:46Z,swp,us-east,host.global,load5,11
,,2,2018-05-22T19:53:56Z,swp,us-east,host.global,load5,12
,,2,2018-05-22T19:54:06Z,swp,us-east,host.global,load5,13
,,2,2018-05-22T19:54:16Z,swp,us-east,host.global,load5,13

#datatype,string,long,dateTime:RFC3339,string,string,string,string,double
#group,false,false,false,true,true,true,true,false
#default,_result,,,,,,,
,result,table,_time,_measurement,region,host,_field,_value
,,0,2018-05-22T19:53:26Z,swp,us-east,host.global,load2,10.003
,,0,2018-05-22T19:53:36Z,swp,us-east,host.global,load2,11.873
,,0,2018-05-22T19:53:46Z,swp,us-east,host.global,load2,18.832
,,0,2018-05-22T19:53:56Z,swp,us-east,host.global,load2,19.777
,,0,2018-05-22T19:54:06Z,swp,us-east,host.global,load2,17.190
,,0,2018-05-22T19:54:16Z,swp,us-east,host.global,load2,17.192
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

measurement_tag_keys_fn = (tables=<-) => tables
    |> range(start: 2018-01-01T00:00:00Z, stop: 2019-01-01T00:00:00Z)
    |> filter(fn: (r) => r._measurement == "swp")
    |> filter(fn: (r) => r.host == "host.global")
    |> keys()
    |> keep(columns: ["_value"])
    |> distinct()
    |> sort()

test measurement_tag_keys = () =>
    ({input: testing.loadStorage(csv: input), want: testing.loadMem(csv: output), fn: measurement_tag_keys_fn})
