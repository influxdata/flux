package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,string,dateTime:RFC3339,long
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,1,m1,f1,server01,2018-12-19T22:13:40Z,0
#datatype,string,long,string,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,2,m1,f2,server01,2018-12-19T22:13:50Z,1
#datatype,string,long,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,4,m1,f4,server01,2018-12-19T22:14:10Z,2018-12-19T22:14:10Z
#datatype,string,long,string,string,string,dateTime:RFC3339,dateTime:RFC3339
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,5,m1,f5,server01,2018-12-19T22:14:10Z,2018-12-19T22:14:10Z
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,dateTime:RFC3339
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:40Z,1970-01-01T00:00:00.000000000Z
,,2,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f2,server01,2018-12-19T22:13:50Z,1970-01-01T00:00:00.000000001Z
,,4,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f4,server01,2018-12-19T22:14:10Z,2018-12-19T22:14:10Z
,,5,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f5,server01,2018-12-19T22:14:10Z,2018-12-19T22:14:10Z
"

t_to_time = (table=<-) =>
	(table
		|> range(start: 2018-12-15T00:00:00Z)
		|> toTime())

test _to = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_to_time})

