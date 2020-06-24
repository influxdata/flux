package date_test

import "testing"
import "date"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,long
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:01:00.254819212Z,_m,FF,-3
,,0,2018-05-22T19:02:00.748691723Z,_m,FF,-2
,,0,2018-05-22T19:03:00.947182316Z,_m,FF,-1
,,0,2018-05-22T19:04:00.538816341Z,_m,FF,0
,,0,2018-05-22T19:05:00.676423456Z,_m,FF,1
,,0,2018-05-22T19:06:00.982342357Z,_m,FF,2
"

outData = "
#group,false,false,true,true,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,long
#default,_result,,,,,,,
,result,table,_start,_stop,_field,_measurement,_time,_value
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:01:00.254819212Z,999999
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:02:00.748691723Z,999999
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:03:00.947182316Z,999999
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:04:00.538816341Z,0
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:05:00.676423456Z,0
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-05-22T19:06:00.982342357Z,0
"

t_duration_microsecond = (table=<-) =>
	(table
		|> range(start: 2018-01-01T00:00:00Z)
		|> map(fn: (r) => ({r with _value: date.microsecond(t: duration(v : r._value))})))

test _duration_microsecond = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_duration_microsecond})
