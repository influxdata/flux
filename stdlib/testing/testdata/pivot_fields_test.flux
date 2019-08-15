package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,1.83,load1,system,host.local
,,0,2018-05-22T19:53:36Z,1.7,load1,system,host.local
,,0,2018-05-22T19:53:46Z,1.74,load1,system,host.local
,,0,2018-05-22T19:53:56Z,1.63,load1,system,host.local
,,0,2018-05-22T19:54:06Z,1.91,load1,system,host.local
,,0,2018-05-22T19:54:16Z,1.84,load1,system,host.local
,,1,2018-05-22T19:53:26Z,1.98,load15,system,host.local
,,1,2018-05-22T19:53:36Z,1.97,load15,system,host.local
,,1,2018-05-22T19:53:46Z,1.97,load15,system,host.local
,,1,2018-05-22T19:53:56Z,1.96,load15,system,host.local
,,1,2018-05-22T19:54:06Z,1.98,load15,system,host.local
,,1,2018-05-22T19:54:16Z,1.97,load15,system,host.local
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,2,2018-05-22T19:53:26Z,95,load5,system,host.local
,,2,2018-05-22T19:53:36Z,92,load5,system,host.local
,,2,2018-05-22T19:53:46Z,92,load5,system,host.local
,,2,2018-05-22T19:53:56Z,89,load5,system,host.local
,,2,2018-05-22T19:54:06Z,94,load5,system,host.local
,,2,2018-05-22T19:54:16Z,93,load5,system,host.local

#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,3,2018-05-22T19:53:26Z,82,used_percent,swap,host.local
,,3,2018-05-22T19:53:36Z,83,used_percent,swap,host.local
,,3,2018-05-22T19:53:46Z,84,used_percent,swap,host.local
,,3,2018-05-22T19:53:56Z,85,used_percent,swap,host.local
,,3,2018-05-22T19:54:06Z,82,used_percent,swap,host.local
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,long
#group,false,false,true,true,false,true,true,false
#default,got,,,,,,,
,result,table,_start,_stop,_time,_measurement,host,used_percent
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,swap,host.local,82
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,swap,host.local,83
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,swap,host.local,84
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,swap,host.local,85
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,swap,host.local,82
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double,double,long
#group,false,false,true,true,false,true,true,false,false,false
#default,got,,,,,,,,,
,result,table,_start,_stop,_time,_measurement,host,load1,load15,load5
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,system,host.local,1.83,1.98,95
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,system,host.local,1.7,1.97,92
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,system,host.local,1.74,1.97,92
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,system,host.local,1.63,1.96,89
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,system,host.local,1.91,1.98,94
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,system,host.local,1.84,1.97,93
"

t_pivot_fields = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> yield(name: "0"))

test _pivot_fields = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_pivot_fields})

