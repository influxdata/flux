package main
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,1,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:36Z,2,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:46Z,3,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:56Z,5,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:06Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:16Z,1,usage_guest,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:36Z,4,field1,cpu,cpu-total,host.local
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,,,usage_guest,cpu,cpu-total,host.local
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,,,usage_guest_nice,cpu,cpu-total,host.local
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
"

t_difference_panic = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> filter(fn: (r) =>
			(r._field == "no_exist"))
		|> difference())

test _difference_panic = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_difference_panic})

testing.run(case: _difference_panic)