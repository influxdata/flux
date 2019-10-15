package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-04-17T00:00:00Z,m1,f1,42
,,0,2018-04-17T00:00:01Z,m1,f1,43
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,true,false,false
#default,_result,,,,,,,
,result,table,_start,_stop,_measurement,_field,_time,_value
,,0,2018-04-17T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,2018-04-17T00:00:01Z,43
"
simple_max = (table=<-) =>
	table
		|> range(start: 2018-04-17T00:00:00Z)
		|> max(column: "_value")

test _simple_max = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: simple_max})

