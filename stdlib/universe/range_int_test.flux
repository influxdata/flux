package universe_test
 
import "testing"

option now = () => (1970-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,0,1970-01-01T00:00:03Z,0,usage_guest,cpu,cpu-total,host.local
,,0,1970-01-01T00:00:04Z,0,usage_guest,cpu,cpu-total,host.local
,,0,1970-01-01T00:00:05Z,0,usage_guest,cpu,cpu-total,host.local
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,0,1970-01-01T00:00:01Z,1970-01-01T00:01:40Z,1970-01-01T00:00:03Z,0,usage_guest,cpu,cpu-total,host.local
,,0,1970-01-01T00:00:01Z,1970-01-01T00:01:40Z,1970-01-01T00:00:04Z,0,usage_guest,cpu,cpu-total,host.local
,,0,1970-01-01T00:00:01Z,1970-01-01T00:01:40Z,1970-01-01T00:00:05Z,0,usage_guest,cpu,cpu-total,host.local
"

t_range = (table=<-) =>
	(table
		|> range(start: 1, stop: 100))

test _range = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_range})

