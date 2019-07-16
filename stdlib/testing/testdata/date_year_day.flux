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
,,0,2018-05-23T19:53:10Z,_m,FF,1
,,0,2018-05-24T19:53:20Z,_m,FF,1
,,0,2018-05-25T19:53:30Z,_m,FF,1
,,0,2018-05-26T19:53:40Z,_m,FF,1
,,0,2018-05-27T19:53:50Z,_m,FF,1
,,1,2018-05-28T19:53:00Z,_m,QQ,1
,,1,2018-05-29T19:53:10Z,_m,QQ,1
,,1,2018-05-30T19:53:20Z,_m,QQ,1
,,1,2018-05-31T19:53:30Z,_m,QQ,1
,,1,2018-06-01T19:53:40Z,_m,QQ,1
,,1,2018-06-02T19:53:50Z,_m,QQ,1
,,1,2018-06-03T19:54:00Z,_m,QQ,1
,,1,2018-12-31T19:54:00Z,_m,QQ,1
,,1,2019-01-01T19:54:00Z,_m,QQ,1
"

outData = "
#group,false,false,true,true,false,false
#datatype,string,long,string,string,dateTime:RFC3339,long
#default,_result,,,,,
,result,table,_field,_measurement,_time,_value
,,0,FF,_m,2018-05-22T19:53:00Z,142
,,0,FF,_m,2018-05-23T19:53:10Z,143
,,0,FF,_m,2018-05-24T19:53:20Z,144
,,0,FF,_m,2018-05-25T19:53:30Z,145
,,0,FF,_m,2018-05-26T19:53:40Z,146
,,0,FF,_m,2018-05-27T19:53:50Z,147
,,1,QQ,_m,2018-05-28T19:53:00Z,148
,,1,QQ,_m,2018-05-29T19:53:10Z,149
,,1,QQ,_m,2018-05-30T19:53:20Z,150
,,1,QQ,_m,2018-05-31T19:53:30Z,151
,,1,QQ,_m,2018-06-01T19:53:40Z,152
,,1,QQ,_m,2018-06-02T19:53:50Z,153
,,1,QQ,_m,2018-06-03T19:54:00Z,154
,,1,QQ,_m,2018-12-31T19:54:00Z,365
,,1,QQ,_m,2019-01-01T19:54:00Z,1
"

t_time_year_day = (table=<-) =>
	(table
		|> map(fn: (r) => ({r with _value: date.yearDay(t: r._time)})))

test _time_year_day = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_time_year_day})
