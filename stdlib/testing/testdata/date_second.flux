package testdata_test

import "testing"
import "date"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:53:00Z,_m,FF,1
,,0,2018-05-22T19:53:10Z,_m,FF,1
,,0,2018-05-22T19:53:20Z,_m,FF,1
,,0,2018-05-22T19:53:30Z,_m,FF,1
,,0,2018-05-22T19:53:40Z,_m,FF,1
,,0,2018-05-22T19:53:50Z,_m,FF,1
,,1,2018-05-22T19:53:00Z,_m,QQ,1
,,1,2018-05-22T19:53:10Z,_m,QQ,1
,,1,2018-05-22T19:53:20Z,_m,QQ,1
,,1,2018-05-22T19:53:30Z,_m,QQ,1
,,1,2018-05-22T19:53:40Z,_m,QQ,1
,,1,2018-05-22T19:53:50Z,_m,QQ,1
,,1,2018-05-22T19:54:00Z,_m,QQ,1
,,1,2018-05-22T19:54:10Z,_m,QQ,1
,,1,2018-05-22T19:54:20Z,_m,QQ,1
,,2,2018-05-22T19:53:00Z,_m,RR,1
,,2,2018-05-22T19:53:10Z,_m,RR,1
,,2,2018-05-22T19:53:20Z,_m,RR,1
,,2,2018-05-22T19:53:30Z,_m,RR,1
,,3,2018-05-22T19:53:40Z,_m,SR,1
,,3,2018-05-22T19:53:50Z,_m,SR,1
,,3,2018-05-22T19:54:00Z,_m,SR,1
"

outData = "
#group,false,false,true,true,false,false
#datatype,string,long,string,string,dateTime:RFC3339,long
#default,_result,,,,,
,result,table,_field,_measurement,_time,_value
,,0,FF,_m,2018-05-22T19:53:00Z,0
,,0,FF,_m,2018-05-22T19:53:10Z,10
,,0,FF,_m,2018-05-22T19:53:20Z,20
,,0,FF,_m,2018-05-22T19:53:30Z,30
,,0,FF,_m,2018-05-22T19:53:40Z,40
,,0,FF,_m,2018-05-22T19:53:50Z,50
,,1,QQ,_m,2018-05-22T19:53:00Z,0
,,1,QQ,_m,2018-05-22T19:53:10Z,10
,,1,QQ,_m,2018-05-22T19:53:20Z,20
,,1,QQ,_m,2018-05-22T19:53:30Z,30
,,1,QQ,_m,2018-05-22T19:53:40Z,40
,,1,QQ,_m,2018-05-22T19:53:50Z,50
,,1,QQ,_m,2018-05-22T19:54:00Z,0
,,1,QQ,_m,2018-05-22T19:54:10Z,10
,,1,QQ,_m,2018-05-22T19:54:20Z,20
,,2,RR,_m,2018-05-22T19:53:00Z,0
,,2,RR,_m,2018-05-22T19:53:10Z,10
,,2,RR,_m,2018-05-22T19:53:20Z,20
,,2,RR,_m,2018-05-22T19:53:30Z,30
,,3,SR,_m,2018-05-22T19:53:40Z,40
,,3,SR,_m,2018-05-22T19:53:50Z,50
,,3,SR,_m,2018-05-22T19:54:00Z,0
"

t_time_second = (table=<-) =>
	(table
		|> map(fn: (r) => ({r with _value: date.second(t: r._time)})))

test _time_second = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_time_second})
