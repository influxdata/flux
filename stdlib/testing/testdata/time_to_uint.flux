package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,string,dateTime:RFC3339,dateTime:RFC3339
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,5,m1,f5,server01,2018-12-19T22:14:10Z,2018-12-19T22:14:10Z
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f5,server01,2018-12-19T22:14:10Z,1545257650000000000
"

t_to_int = (table=<-) =>
	(table
		|> range(start: 2018-12-15T00:00:00Z)
		|> drop(columns:["_value"])
		|> duplicate(column:"_time", as: "_value")
		|> toUInt())

test _to = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_to_int})

