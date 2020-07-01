package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:50:27Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:50:28Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:50:29Z,0,usage_guest,cpu,cpu-total,host.local
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:50:26Z,2030-01-01T00:00:00Z,2018-05-22T19:50:27Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:50:26Z,2030-01-01T00:00:00Z,2018-05-22T19:50:28Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:50:26Z,2030-01-01T00:00:00Z,2018-05-22T19:50:29Z,0,usage_guest,cpu,cpu-total,host.local
"

t_range = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:50:26Z, stop: 0h))

test _range = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_range})

