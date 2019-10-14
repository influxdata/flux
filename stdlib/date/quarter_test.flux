package date_test

import "testing"
import "date"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-01-22T19:53:00Z,_m,FF,1
,,0,2018-02-22T19:53:10Z,_m,FF,1
,,0,2018-03-22T19:53:20Z,_m,FF,1
,,0,2018-04-22T19:53:30Z,_m,FF,1
,,0,2018-05-22T19:53:40Z,_m,FF,1
,,0,2018-06-22T19:53:50Z,_m,FF,1
,,1,2018-07-22T19:53:00Z,_m,QQ,1
,,1,2018-08-22T19:53:10Z,_m,QQ,1
,,1,2018-09-22T19:53:20Z,_m,QQ,1
,,1,2018-10-22T19:53:30Z,_m,QQ,1
,,1,2018-11-22T19:53:40Z,_m,QQ,1
,,1,2018-12-22T19:53:50Z,_m,QQ,1
"

outData = "
#group,false,false,true,true,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,long
#default,_result,,,,,,,
,result,table,_start,_stop,_field,_measurement,_time,_value
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-01-22T19:53:00Z,1
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-02-22T19:53:10Z,1
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-03-22T19:53:20Z,1
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-04-22T19:53:30Z,2
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:53:40Z,2
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-06-22T19:53:50Z,2
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-07-22T19:53:00Z,3
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-08-22T19:53:10Z,3
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-09-22T19:53:20Z,3
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-10-22T19:53:30Z,4
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-11-22T19:53:40Z,4
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2018-12-22T19:53:50Z,4
"

t_time_quarter = (table=<-) =>
	(table
	    |> range(start: 2018-01-01T00:00:00Z)
		|> map(fn: (r) => ({r with _value: date.quarter(t: r._time)})))

test _time_quarter = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_time_quarter})
