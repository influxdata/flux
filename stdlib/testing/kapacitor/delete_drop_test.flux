package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,string,dateTime:RFC3339,boolean
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m1,f1,server01,2018-12-19T22:13:30Z,false
,,0,m1,f1,server01,2018-12-19T22:13:40Z,true
,,0,m1,f1,server01,2018-12-19T22:13:50Z,false
,,0,m1,f1,server01,2018-12-19T22:14:00Z,false
,,0,m1,f1,server01,2018-12-19T22:14:10Z,true
,,0,m1,f1,server01,2018-12-19T22:14:20Z,true
,,1,m1,f1,server02,2018-12-19T22:13:30Z,false
,,1,m1,f1,server02,2018-12-19T22:13:40Z,true
,,1,m1,f1,server02,2018-12-19T22:13:50Z,false
,,1,m1,f1,server02,2018-12-19T22:14:00Z,true
,,1,m1,f1,server02,2018-12-19T22:14:10Z,true
,,1,m1,f1,server02,2018-12-19T22:14:20Z,false
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339
#group,false,false,true,true,true,true,true,false
#default,_result,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:30Z
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:40Z
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:50Z
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:00Z
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:10Z
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:20Z
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:30Z
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:40Z
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:50Z
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:00Z
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:10Z
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:20Z
"

t_drop = (table=<-) =>
	(table
		|> range(start: 2018-12-15T00:00:00Z)
		|> drop(columns: ["_value"]))

test _drop = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_drop})

// Equivalent TICKscript query:
// stream
//   |delete()
//     .field('_value')