package testdata_test

import "testing"
import "date"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T00:53:00Z,_m,FF,1
,,0,2018-05-22T01:53:00Z,_m,FF,1
,,0,2018-05-22T02:53:10Z,_m,FF,1
,,0,2018-05-22T03:53:20Z,_m,FF,1
,,0,2018-05-22T04:53:30Z,_m,FF,1
,,0,2018-05-22T05:53:40Z,_m,FF,1
,,0,2018-05-22T06:53:50Z,_m,FF,1
,,1,2018-05-22T07:53:00Z,_m,QQ,1
,,1,2018-05-22T08:53:10Z,_m,QQ,1
,,1,2018-05-22T09:53:20Z,_m,QQ,1
,,1,2018-05-22T10:53:30Z,_m,QQ,1
,,1,2018-05-22T11:53:40Z,_m,QQ,1
,,1,2018-05-22T12:53:50Z,_m,QQ,1
,,1,2018-05-22T13:54:00Z,_m,QQ,1
,,1,2018-05-22T14:54:10Z,_m,QQ,1
,,1,2018-05-22T15:54:20Z,_m,QQ,1
,,2,2018-05-22T16:53:00Z,_m,RR,1
,,2,2018-05-22T17:53:10Z,_m,RR,1
,,2,2018-05-22T18:53:20Z,_m,RR,1
,,2,2018-05-22T19:53:30Z,_m,RR,1
,,3,2018-05-22T20:53:40Z,_m,SR,1
,,3,2018-05-22T21:53:50Z,_m,SR,1
,,3,2018-05-22T22:53:00Z,_m,SR,1
,,3,2018-05-22T23:53:50Z,_m,SR,1
"

outData = "
#group,false,false,true,true,false,false
#datatype,string,long,string,string,dateTime:RFC3339,long
#default,_result,,,,,
,result,table,_field,_measurement,_time,_value
,,0,FF,_m,2018-05-22T00:53:00Z,0
,,0,FF,_m,2018-05-22T01:53:00Z,1
,,0,FF,_m,2018-05-22T02:53:10Z,2
,,0,FF,_m,2018-05-22T03:53:20Z,3
,,0,FF,_m,2018-05-22T04:53:30Z,4
,,0,FF,_m,2018-05-22T05:53:40Z,5
,,0,FF,_m,2018-05-22T06:53:50Z,6
,,1,QQ,_m,2018-05-22T07:53:00Z,7
,,1,QQ,_m,2018-05-22T08:53:10Z,8
,,1,QQ,_m,2018-05-22T09:53:20Z,9
,,1,QQ,_m,2018-05-22T10:53:30Z,10
,,1,QQ,_m,2018-05-22T11:53:40Z,11
,,1,QQ,_m,2018-05-22T12:53:50Z,12
,,1,QQ,_m,2018-05-22T13:54:00Z,13
,,1,QQ,_m,2018-05-22T14:54:10Z,14
,,1,QQ,_m,2018-05-22T15:54:20Z,15
,,2,RR,_m,2018-05-22T16:53:00Z,16
,,2,RR,_m,2018-05-22T17:53:10Z,17
,,2,RR,_m,2018-05-22T18:53:20Z,18
,,2,RR,_m,2018-05-22T19:53:30Z,19
,,3,SR,_m,2018-05-22T20:53:40Z,20
,,3,SR,_m,2018-05-22T21:53:50Z,21
,,3,SR,_m,2018-05-22T22:53:00Z,22
,,3,SR,_m,2018-05-22T23:53:50Z,23
"

t_time_hour = (table=<-) =>
	(table
		|> map(fn: (r) => ({r with _value: date.hour(t: r._time)})))

test _time_hour = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_time_hour})
