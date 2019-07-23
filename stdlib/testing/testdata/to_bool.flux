package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,boolean,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T19:53:00Z,true,k0,m
,,0,2018-05-22T19:53:01Z,false,k0,m
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,1,2018-05-22T19:53:10Z,1e0,k1,m
,,1,2018-05-22T19:53:11Z,1.00,k1,m
,,1,2018-05-22T19:53:12Z,1,k1,m
,,1,2018-05-22T19:53:13Z,1.0,k1,m
,,1,2018-05-22T19:53:14Z,0.0,k1,m
,,1,2018-05-22T19:53:15Z,0.00,k1,m
,,1,2018-05-22T19:53:16Z,0,k1,m
#datatype,string,long,dateTime:RFC3339,long,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,2,2018-05-22T19:53:20Z,1,k2,m
,,2,2018-05-22T19:53:21Z,0,k2,m
#datatype,string,long,dateTime:RFC3339,string,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,3,2018-05-22T19:53:30Z,true,k3,m
,,3,2018-05-22T19:53:31Z,false,k3,m
#datatype,string,long,dateTime:RFC3339,unsignedLong,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,4,2018-05-22T19:53:40Z,1,k4,m
,,4,2018-05-22T19:53:41Z,0,k4,m
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,boolean
#group,false,false,true,true,false,true,true,false
#default,want,,,,,,,
,result,table,_start,_stop,_time,_field,_measurement,_value
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,k0,m,true
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:01Z,k0,m,false
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:10Z,k1,m,true
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:11Z,k1,m,true
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:12Z,k1,m,true
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:13Z,k1,m,true
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:14Z,k1,m,false
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:15Z,k1,m,false
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:16Z,k1,m,false
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:20Z,k2,m,true
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:21Z,k2,m,false
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:30Z,k3,m,true
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:31Z,k3,m,false
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:40Z,k4,m,true
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:41Z,k4,m,false
"

t_to_bool = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:52:00Z)
		|> toBool())

test _to = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_to_bool})


