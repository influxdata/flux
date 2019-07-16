package testdata_test

import "testing"
import "date"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-20T19:53:00Z,_m,FF,1
,,0,2018-05-21T19:53:10Z,_m,FF,1
,,0,2018-05-22T19:53:20Z,_m,FF,1
,,0,2018-05-23T19:53:30Z,_m,FF,1
,,0,2018-05-24T19:53:40Z,_m,FF,1
,,0,2018-05-25T19:53:50Z,_m,FF,1
,,1,2018-05-26T19:53:00Z,_m,QQ,1
,,1,2018-05-27T19:53:10Z,_m,QQ,1
"

outData = "
#group,false,false,true,true,false,false
#datatype,string,long,string,string,dateTime:RFC3339,long
#default,_result,,,,,
,result,table,_field,_measurement,_time,_value
,,0,FF,_m,2018-05-20T19:53:00Z,0
,,0,FF,_m,2018-05-21T19:53:10Z,1
,,0,FF,_m,2018-05-22T19:53:20Z,2
,,0,FF,_m,2018-05-23T19:53:30Z,3
,,0,FF,_m,2018-05-24T19:53:40Z,4
,,0,FF,_m,2018-05-25T19:53:50Z,5
,,1,QQ,_m,2018-05-26T19:53:00Z,6
,,1,QQ,_m,2018-05-27T19:53:10Z,0
"

t_time_week_day = (table=<-) =>
	(table
		|> map(fn: (r) => ({r with _value: date.weekDay(t: r._time)})))

test _time_week_day = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_time_week_day})
