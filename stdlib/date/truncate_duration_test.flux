package date_test

import "testing"
import "date"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,long
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:53:00.000000000Z,_m,FF,-3600000000000
,,0,2018-05-22T19:53:10.000000000Z,_m,FF,-7200000000000
,,0,2018-05-22T19:53:20.000000000Z,_m,FF,-10800000000000
,,0,2018-05-22T19:53:30.000000000Z,_m,FF,0
,,0,2018-05-22T19:53:40.000000000Z,_m,FF,3600000000000
,,0,2018-05-22T19:53:50.000000000Z,_m,FF,7200000000000
,,0,2018-05-22T19:54:00.000000000Z,_m,FF,10800000000000
"

outData = "
#datatype,string,long,string,string,dateTime:RFC3339,long
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_field,_measurement,_time,_value
,,0,FF,_m,2018-05-22T19:53:00.000000000Z,1893452400000000000
,,0,FF,_m,2018-05-22T19:53:10.000000000Z,1893448800000000000
,,0,FF,_m,2018-05-22T19:53:20.000000000Z,1893445200000000000
,,0,FF,_m,2018-05-22T19:53:30.000000000Z,1893456000000000000
,,0,FF,_m,2018-05-22T19:53:40.000000000Z,1893459600000000000
,,0,FF,_m,2018-05-22T19:53:50.000000000Z,1893463200000000000
,,0,FF,_m,2018-05-22T19:54:00.000000000Z,1893466800000000000
"

t_duration_truncate = (table=<-) =>
	(table
	    |> range(start: 2018-05-22T19:53:00.000000000Z)
	    |> drop(columns: ["_start", "_stop"])
		|> map(fn: (r) => ({r with _value: int(v: date.truncate(t: duration(v: r._value), unit: 1s))})))

test _duration_truncate = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_duration_truncate})
