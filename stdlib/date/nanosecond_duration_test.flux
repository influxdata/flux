package date_test

import "testing"
import "date"

option now = () => (2030-01-01T00:00:00.000000100Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,long
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:01:00.254819212Z,_m,FF,-1
,,0,2018-05-22T19:02:00.748691723Z,_m,FF,-2
,,0,2018-05-22T19:03:00.947182316Z,_m,FF,-3
,,0,2018-05-22T19:04:00.538816341Z,_m,FF,0
,,0,2018-05-22T19:05:00.676423456Z,_m,FF,1
,,0,2018-05-22T19:06:00.982342357Z,_m,FF,2
"

outData = "
#group,false,false,true,true,false,false
#datatype,string,long,string,string,dateTime:RFC3339,long
#default,_result,,,,,
,result,table,_field,_measurement,_time,_value
,,0,FF,_m,2018-05-22T19:01:00.254819212Z,99
,,0,FF,_m,2018-05-22T19:02:00.748691723Z,98
,,0,FF,_m,2018-05-22T19:03:00.947182316Z,97
,,0,FF,_m,2018-05-22T19:04:00.538816341Z,100
,,0,FF,_m,2018-05-22T19:05:00.676423456Z,101
,,0,FF,_m,2018-05-22T19:06:00.982342357Z,102
"

t_duration_nanosecond = (table=<-) =>
	(table
		|> range(start: 2018-01-01T00:00:00Z)
	    |> drop(columns: ["_start", "_stop"])
		|> map(fn: (r) => ({r with _value: int(v: date.nanosecond(t: duration(v: r._value)))})))

test _time_nanosecond = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_duration_nanosecond})
