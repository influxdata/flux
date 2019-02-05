package main
import "testing"

option now = () =>
	(2030-01-01T00:00:00Z)

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
#datatype,string,long,string,string,string,double
#group,false,false,true,true,true,false
#default,_result,,,,,
,result,table,_field,_measurement,host,newValue
,,0,load1,system,host.local,100.0
,,0,load1,system,host.local,101.0
,,0,load1,system,host.local,102.0
"
t_map = (table=<-) =>
	(table
		|> map(fn: (r) =>
			({newValue: float(v: r._value)})))

test map = {input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map}