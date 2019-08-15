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
#datatype,string,long,double
#group,false,false,false
#default,_result,,
,result,table,newValue
,,0,100.0
,,0,101.0
,,0,102.0
"

t_map = (table=<-) =>
	(table
		|> map(fn: (r) =>
			({newValue: float(v: r._value)})))

test _map = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map})

