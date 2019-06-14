package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,100,load1,system,host.local
,,0,2018-05-22T19:53:36Z,101,load1,system,host.local
,,0,2018-05-22T19:53:46Z,102,load1,system,host.local
"

outData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,100.0,load1,system,host.local
,,0,2018-05-22T19:53:36Z,101.0,load1,system,host.local
,,0,2018-05-22T19:53:46Z,102.0,load1,system,host.local
"

t_map = (table=<-) =>
	(table
		|> map(fn: (r) =>
			({r with _value: float(v:r._value)})))

test _map = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map})

