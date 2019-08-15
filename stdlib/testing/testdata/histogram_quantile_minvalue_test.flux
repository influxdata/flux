package testdata_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,double,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_time,_field,_value,le
,,0,2018-05-22T19:53:00Z,x_duration_seconds,10,-80
,,0,2018-05-22T19:53:00Z,x_duration_seconds,11,-60
,,0,2018-05-22T19:53:00Z,x_duration_seconds,12,-40
,,0,2018-05-22T19:53:00Z,x_duration_seconds,13,-20
,,0,2018-05-22T19:53:00Z,x_duration_seconds,14,-0
,,0,2018-05-22T19:53:00Z,x_duration_seconds,15,20
,,0,2018-05-22T19:53:00Z,x_duration_seconds,16,40
,,0,2018-05-22T19:53:00Z,x_duration_seconds,17,60
,,0,2018-05-22T19:53:00Z,x_duration_seconds,18,80
,,0,2018-05-22T19:53:00Z,x_duration_seconds,19,+Inf
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,double
#group,false,false,true,true,true,true,false
#default,_result,,,,,,
,result,table,_start,_stop,_time,_field,_value
,,0,2018-05-22T19:53:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,x_duration_seconds,-90.5
"

t_histogram_quantile = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:00Z)
		|> histogramQuantile(quantile: 0.25, minValue: -100.0))

test _histogram_quantile_minvalue = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_histogram_quantile})

