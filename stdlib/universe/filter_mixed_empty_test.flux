package universe_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,name
,,0,2018-05-22T19:53:26Z,15204688,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:36Z,15204894,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:46Z,15205102,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:56Z,15205226,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:54:06Z,15205499,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:54:16Z,15205755,io_time,diskio,host.local,disk0
,,1,2018-05-22T19:53:26Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:36Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:46Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:56Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:06Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:16Z,648,io_time,diskio,host.local,disk2

#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,2,2018-05-22T19:53:26Z,1.83,load1,system,host.local
,,2,2018-05-22T19:53:36Z,1.7,load1,system,host.local
,,2,2018-05-22T19:53:46Z,1.74,load1,system,host.local
,,2,2018-05-22T19:53:56Z,1.63,load1,system,host.local
,,2,2018-05-22T19:54:06Z,1.91,load1,system,host.local
,,2,2018-05-22T19:54:16Z,1.84,load1,system,host.local
,,3,2018-05-22T19:53:26Z,1.98,load15,system,host.local
,,3,2018-05-22T19:53:36Z,1.97,load15,system,host.local
,,3,2018-05-22T19:53:46Z,1.97,load15,system,host.local
,,3,2018-05-22T19:53:56Z,1.96,load15,system,host.local
,,3,2018-05-22T19:54:06Z,1.98,load15,system,host.local
,,3,2018-05-22T19:54:16Z,1.97,load15,system,host.local
,,4,2018-05-22T19:53:26Z,1.95,load5,system,host.local
,,4,2018-05-22T19:53:36Z,1.92,load5,system,host.local
,,4,2018-05-22T19:53:46Z,1.92,load5,system,host.local
,,4,2018-05-22T19:53:56Z,1.89,load5,system,host.local
,,4,2018-05-22T19:54:06Z,1.94,load5,system,host.local
,,4,2018-05-22T19:54:16Z,1.93,load5,system,host.local
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,long
#group,false,false,true,true,true,true,true,true,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,host,name,_value
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,diskio,io_time,host.local,disk0,6
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,diskio,io_time,host.local,disk2,0
"

t_filter_mixed_empty = (table=<-) =>
  table
  |> range(start: 2018-05-22T19:53:26Z)
  |> filter(fn: (r) => r._measurement == "diskio")
  |> filter(fn: (r) => r["_value"] > 1000, onEmpty: "keep")
  |> count()

test _filter_mixed_empty = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_filter_mixed_empty})

