package universe_test
 
import c "csv"
import "testing"

option now = () => (2018-12-19T22:15:00Z)
option testing.loadStorage = (csv) => c.from(csv: csv)

inData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,true,true,false,false
#default,,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:30Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:40Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:50Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:10Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:20Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:40Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:50Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:00Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:10Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:20Z,A
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,string
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:30Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:40Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:50Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2077-12-19T22:14:00Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:10Z,A
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:20Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2077-12-19T22:14:00Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:40Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:50Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:00Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:10Z,A
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:20Z,A
"

t_fill_float = (table=<-) =>
	(table
		|> fill(column: "_time", value: 2077-12-19T22:14:00Z))

test _fill = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_fill_float})

