package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,double
#group,false,false,true,true,false
#default,_result,,,,
,result,table,_time,_field,theValue
,,0,2018-05-22T19:53:00Z,x_duration_seconds,0
,,0,2018-05-22T19:53:00Z,x_duration_seconds,1
,,1,2018-05-22T19:53:00Z,y_duration_seconds,0
,,1,2018-05-22T19:53:00Z,y_duration_seconds,0
,,1,2018-05-22T19:53:00Z,y_duration_seconds,1.5
"

outData = "
#datatype,string,long,dateTime:RFC3339,string,double,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_time,_field,ub,theCount
,,0,2018-05-22T19:53:00Z,x_duration_seconds,-1,0
,,0,2018-05-22T19:53:00Z,x_duration_seconds,0,0.5
,,0,2018-05-22T19:53:00Z,x_duration_seconds,1,1
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,1
,,1,2018-05-22T19:53:00Z,y_duration_seconds,-1,0
,,1,2018-05-22T19:53:00Z,y_duration_seconds,0,0.6666666666666666
,,1,2018-05-22T19:53:00Z,y_duration_seconds,1,0.6666666666666666
,,1,2018-05-22T19:53:00Z,y_duration_seconds,2,1
"

t_histogram = (table=<-) =>
	(table
		|> histogram(
			bins: [-1.0, 0.0, 1.0, 2.0],
			normalize: true,
			column: "theValue",
			countColumn: "theCount",
			upperBoundColumn: "ub",
		))

test _histogram = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_histogram})

