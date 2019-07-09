package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,string,dateTime:RFC3339,boolean
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m1,f0,server01,2018-12-19T22:13:30Z,false
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
#datatype,string,long,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,3,m1,f3,server01,2018-12-19T22:14:00Z,1.0
#datatype,string,long,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,4,m1,f4,server01,2018-12-19T22:14:10Z,false
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f0,server01,2018-12-19T22:13:30Z,false
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:40Z,0
,,2,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f2,server01,2018-12-19T22:13:50Z,1
,,3,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f3,server01,2018-12-19T22:14:00Z,1
,,4,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f4,server01,2018-12-19T22:14:10Z,false
"

t_to_string = (table=<-) =>
	(table
		|> range(start: 2018-12-15T00:00:00Z)
		|> toString())

test _to = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_to_string})

