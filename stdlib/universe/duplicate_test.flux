package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:36Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:46Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:56Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:06Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:16Z,0,usage_guest,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:36Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:46Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:56Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:54:06Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:54:16Z,0,usage_guest_nice,cpu,cpu-total,host.local
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,false
#default,_result,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host,host_new
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,0,usage_guest,cpu,cpu-total,host.local,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,0,usage_guest,cpu,cpu-total,host.local,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,0,usage_guest,cpu,cpu-total,host.local,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,0,usage_guest,cpu,cpu-total,host.local,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,0,usage_guest,cpu,cpu-total,host.local,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,0,usage_guest,cpu,cpu-total,host.local,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,0,usage_guest_nice,cpu,cpu-total,host.local,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,0,usage_guest_nice,cpu,cpu-total,host.local,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,0,usage_guest_nice,cpu,cpu-total,host.local,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,0,usage_guest_nice,cpu,cpu-total,host.local,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,0,usage_guest_nice,cpu,cpu-total,host.local,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,0,usage_guest_nice,cpu,cpu-total,host.local,host.local
"

t_duplicate = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> duplicate(column: "host", as: "host_new"))

test _duplicate = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_duplicate})

