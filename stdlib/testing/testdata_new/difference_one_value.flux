package main
// 
import "testing"

option now = () =>
	(2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,1,usage_guest,cpu,cpu-total,host.local
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,,,usage_guest,cpu,cpu-total,host.local
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
"
difference_one_value = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> difference(nonNegative: true))

test difference_one_value = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: difference_one_value})

testing.run(case: difference_one_value)