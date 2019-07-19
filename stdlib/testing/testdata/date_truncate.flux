package testdata_test

import "testing"
import "date"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,long
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:53:00.000000000Z,_m,FF,1
,,0,2018-05-22T19:53:10.000000000Z,_m,FF,1
,,0,2018-05-22T19:53:20.000000000Z,_m,FF,1
,,0,2018-05-22T19:53:30.000000000Z,_m,FF,1
,,0,2018-05-22T19:53:40.000000000Z,_m,FF,1
,,0,2018-05-22T19:53:50.000000000Z,_m,FF,1
"

outData = "
#datatype,string,long,string,string,dateTime:RFC3339,long
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_field,_measurement,_time,_value
,,0,FF,_m,2018-05-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-05-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-05-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-05-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-05-22T19:00:00.000000000Z,1
,,0,FF,_m,2018-05-22T19:00:00.000000000Z,1
"

t_time_truncate = (table=<-) =>
	(table
		|> map(fn: (r) => ({r with _time: date.truncate(t: r._time, unit: 1h)})))

test _time_truncate = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_time_truncate})
