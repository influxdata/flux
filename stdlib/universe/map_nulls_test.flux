package universe_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,long,string
#group,false,false,false,true,false,true
#default,_result,,,,,
,result,table,_time,_field,_value,_measurement
,,0,2018-05-22T19:53:26Z,a,1,aa
,,0,2018-05-22T19:53:36Z,a,1,aa
,,0,2018-05-22T19:53:46Z,a,1,aa
,,1,2018-05-22T19:53:36Z,b,1,aa
,,1,2018-05-22T19:53:46Z,b,1,aa
"

outData = "
#datatype,string,long,dateTime:RFC3339,long
#group,false,false,false,false
#default,0,,,
,result,table,_time,_value
,,0,2018-05-22T19:53:26Z,
,,0,2018-05-22T19:53:36Z,1
,,0,2018-05-22T19:53:46Z,1
"

t_pivot = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> map(fn: (r) => ({_time: r._time, _value: r.a / r.b})))

test _pivot = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_pivot})

