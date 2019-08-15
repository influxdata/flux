package testdata_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m1,f1,server01,2018-12-19T22:13:30Z,84
,,0,m1,f1,server01,2018-12-19T22:13:40Z,52
,,0,m1,f1,server01,2018-12-19T22:13:50Z,
,,0,m1,f1,server01,2018-12-19T22:14:00Z,62
,,0,m1,f1,server01,2018-12-19T22:14:10Z,22
,,0,m1,f1,server01,2018-12-19T22:14:20Z,78
,,1,m1,f1,server02,2018-12-19T22:13:30Z,
,,1,m1,f1,server02,2018-12-19T22:13:40Z,33
,,1,m1,f1,server02,2018-12-19T22:13:50Z,97
,,1,m1,f1,server02,2018-12-19T22:14:00Z,90
,,1,m1,f1,server02,2018-12-19T22:14:10Z,96
,,1,m1,f1,server02,2018-12-19T22:14:20Z,
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:30Z,84
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:40Z,52
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:50Z,0
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:00Z,62
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:10Z,22
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:20Z,78
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:30Z,0
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:40Z,33
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:50Z,97
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:00Z,90
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:10Z,96
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:20Z,0
"

t_fill_uint = (table=<-) =>
	(table
		|> range(start: 2018-12-15T00:00:00Z)
		|> fill(column: "_value", value: uint(v: 0)))

test _fill = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_fill_uint})

