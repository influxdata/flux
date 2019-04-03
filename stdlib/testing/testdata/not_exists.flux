package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,true,true,true,false
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,name
,,1,2018-05-22T19:53:26Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:36Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:46Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:56Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:06Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:16Z,648,io_time,diskio,host.local,
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,1,2018-05-22T19:53:26Z,648,io_time,diskio,host.local2
,,1,2018-05-22T19:53:36Z,648,io_time,diskio,host.local2
,,1,2018-05-22T19:53:46Z,648,io_time,diskio,host.local2
,,1,2018-05-22T19:53:56Z,648,io_time,diskio,host.local2
,,1,2018-05-22T19:54:06Z,648,io_time,diskio,host.local2
,,1,2018-05-22T19:54:16Z,648,io_time,diskio,host.local2
"

outData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,true,true,true,false
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,name
,,1,2018-05-22T19:54:16Z,648,io_time,diskio,host.local,
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,1,2018-05-22T19:53:26Z,648,io_time,diskio,host.local2
,,1,2018-05-22T19:53:36Z,648,io_time,diskio,host.local2
,,1,2018-05-22T19:53:46Z,648,io_time,diskio,host.local2
,,1,2018-05-22T19:53:56Z,648,io_time,diskio,host.local2
,,1,2018-05-22T19:54:06Z,648,io_time,diskio,host.local2
,,1,2018-05-22T19:54:16Z,648,io_time,diskio,host.local2
"

t_filter_by_tags = (table=<-) =>
  table
  |> range(start: 2018-05-22T19:53:26Z)
  |> filter(fn: (r) => not exists(r:r, key: "name"))
  |> drop(columns:["_start", "_stop"])


test _filter_by_tags = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_filter_by_tags})

