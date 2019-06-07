package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m1,f1,server01,2018-12-19T22:13:30Z,
,,0,m1,f1,server01,2018-12-19T22:13:40Z,
,,0,m1,f1,server01,2018-12-19T22:13:50Z,
,,0,m1,f1,server01,2018-12-19T22:14:00Z,
,,0,m1,f1,server01,2018-12-19T22:14:10Z,
,,0,m1,f1,server01,2018-12-19T22:14:20Z,
,,1,m1,f1,server02,2018-12-19T22:13:30Z,
,,1,m1,f1,server02,2018-12-19T22:13:40Z,
,,1,m1,f1,server02,2018-12-19T22:13:50Z,
,,1,m1,f1,server02,2018-12-19T22:14:00Z,
,,1,m1,f1,server02,2018-12-19T22:14:10Z,
,,1,m1,f1,server02,2018-12-19T22:14:20Z,
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,true,true,false,false
#default,got,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
"

t_fill_float = (table=<-) =>
	(table
		|> range(start: 2018-12-15T00:00:00Z)
		|> filter(fn: (r) => exists(r:r, key: "_value")))

test _fill = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_fill_float})

