package testdata_test

import "testing"
import "date"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:01:00Z,_m,FF,1
,,0,2018-05-22T19:02:00Z,_m,FF,1
,,0,2018-05-22T19:03:00Z,_m,FF,1
,,0,2018-05-22T19:04:00Z,_m,FF,1
,,0,2018-05-22T19:05:00Z,_m,FF,1
,,0,2018-05-22T19:06:00Z,_m,FF,1
,,1,2018-05-22T19:07:00Z,_m,QQ,1
,,1,2018-05-22T19:08:00Z,_m,QQ,1
,,1,2018-05-22T19:09:00Z,_m,QQ,1
,,1,2018-05-22T19:10:00Z,_m,QQ,1
,,1,2018-05-22T19:13:00Z,_m,QQ,1
,,1,2018-05-22T19:15:00Z,_m,QQ,1
,,1,2018-05-22T19:20:00Z,_m,QQ,1
,,1,2018-05-22T19:23:00Z,_m,QQ,1
,,1,2018-05-22T19:25:00Z,_m,QQ,1
,,2,2018-05-22T19:28:00Z,_m,RR,1
,,2,2018-05-22T19:36:00Z,_m,RR,1
,,2,2018-05-22T19:38:00Z,_m,RR,1
,,2,2018-05-22T19:47:00Z,_m,RR,1
,,3,2018-05-22T19:48:00Z,_m,SR,1
,,3,2018-05-22T19:59:00Z,_m,SR,1
,,3,2018-05-22T20:00:00Z,_m,SR,1
"

outData = "
#group,false,false,true,true,false,false
#datatype,string,long,string,string,dateTime:RFC3339,long
#default,_result,,,,,
,result,table,_field,_measurement,_time,_value
,,0,FF,_m,2018-05-22T19:01:00Z,1
,,0,FF,_m,2018-05-22T19:02:00Z,2
,,0,FF,_m,2018-05-22T19:03:00Z,3
,,0,FF,_m,2018-05-22T19:04:00Z,4
,,0,FF,_m,2018-05-22T19:05:00Z,5
,,0,FF,_m,2018-05-22T19:06:00Z,6
,,1,QQ,_m,2018-05-22T19:07:00Z,7
,,1,QQ,_m,2018-05-22T19:08:00Z,8
,,1,QQ,_m,2018-05-22T19:09:00Z,9
,,1,QQ,_m,2018-05-22T19:10:00Z,10
,,1,QQ,_m,2018-05-22T19:13:00Z,13
,,1,QQ,_m,2018-05-22T19:15:00Z,15
,,1,QQ,_m,2018-05-22T19:20:00Z,20
,,1,QQ,_m,2018-05-22T19:23:00Z,23
,,1,QQ,_m,2018-05-22T19:25:00Z,25
,,2,RR,_m,2018-05-22T19:28:00Z,28
,,2,RR,_m,2018-05-22T19:36:00Z,36
,,2,RR,_m,2018-05-22T19:38:00Z,38
,,2,RR,_m,2018-05-22T19:47:00Z,47
,,3,SR,_m,2018-05-22T19:48:00Z,48
,,3,SR,_m,2018-05-22T19:59:00Z,59
,,3,SR,_m,2018-05-22T20:00:00Z,00
"

t_time_minute = (table=<-) =>
	(table
		|> map(fn: (r) => ({r with _value: date.minute(t: r._time)})))

test _time_minute = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_time_minute})
