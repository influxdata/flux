package main
import "testing"

option now = () =>
	(2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,counter,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:26Z,1,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:36Z,2,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:46Z,3,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:56Z,5,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:06Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:16Z,1,usage_guest,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:26Z,2,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:36Z,4,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:46Z,4,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:56Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:06Z,2,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:16Z,10,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:26Z,4,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:36Z,20,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:46Z,7,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:56Z,10,usage_guest_nice,cpu,cpu-total,host.local
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,counter,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,1,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,2,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,4,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,4,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,5,usage_guest,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,2,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,2,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,4,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,12,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:26Z,12,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:36Z,28,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:46Z,28,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:56Z,31,usage_guest_nice,cpu,cpu-total,host.local
"
t_increase = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> increase(columns: ["counter"]))

test increase = {input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_increase}