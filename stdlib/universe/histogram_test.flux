package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,double,string
#group,false,false,true,true,false,true
#default,_result,,,,,
,result,table,_time,_field,_value,_measurement
,,0,2018-05-22T19:53:00Z,x_duration_seconds,0,foo
,,0,2018-05-22T19:53:00Z,x_duration_seconds,1,foo
,,1,2018-05-22T19:53:00Z,y_duration_seconds,0,foo
,,1,2018-05-22T19:53:00Z,y_duration_seconds,0,foo
,,1,2018-05-22T19:53:00Z,y_duration_seconds,1.5,foo
"

outData = "
#datatype,string,long,dateTime:RFC3339,string,double,double,string
#group,false,false,true,true,false,false,true
#default,_result,,,,,,
,result,table,_time,_field,le,_value,_measurement
,,0,2018-05-22T19:53:00Z,x_duration_seconds,-1,0,foo
,,0,2018-05-22T19:53:00Z,x_duration_seconds,0,1,foo
,,0,2018-05-22T19:53:00Z,x_duration_seconds,1,2,foo
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,2,foo
,,1,2018-05-22T19:53:00Z,y_duration_seconds,-1,0,foo
,,1,2018-05-22T19:53:00Z,y_duration_seconds,0,2,foo
,,1,2018-05-22T19:53:00Z,y_duration_seconds,1,2,foo
,,1,2018-05-22T19:53:00Z,y_duration_seconds,2,3,foo
"

t_histogram = (table=<-) =>
	(table
		|> histogram(bins: [-1.0, 0.0, 1.0, 2.0]))
		|> drop(columns: ["_start", "_stop"])

test _histogram = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_histogram})

