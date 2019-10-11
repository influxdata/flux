package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:36Z,4,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:46Z,4,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:56Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:06Z,2,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:16Z,10,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:26Z,4,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:36Z,20,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:46Z,7,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:56Z,10,usage_guest_nice,cpu,cpu-total,host.local
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,0.02,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,0.08,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:26Z,,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:36Z,0.16,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:46Z,,usage_guest_nice,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:56Z,0.03,usage_guest_nice,cpu,cpu-total,host.local
"
derivative_nonnegative = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> derivative(unit: 100ms, nonNegative: true))

test _derivative_nonnegative = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: derivative_nonnegative})

