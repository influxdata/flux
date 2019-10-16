package promql_test
import "experimental"
import "internal/promql"
import "testing"

option now = () =>
	(2030-01-01T00:00:00Z)

inData = "#datatype,string,long,dateTime:RFC3339,string,double,string,string
#group,false,false,false,true,false,true,true
#default,inData,,,,,,
,result,table,_time,_field,_value,le,_measurement
,,0,2018-05-22T19:53:00Z,x_duration_seconds,1,0.1,prometheus
,,1,2018-05-22T19:53:00Z,x_duration_seconds,2,0.2,prometheus
,,2,2018-05-22T19:53:00Z,x_duration_seconds,2,0.3,prometheus
,,3,2018-05-22T19:53:00Z,x_duration_seconds,2,0.4,prometheus
,,4,2018-05-22T19:53:00Z,x_duration_seconds,2,0.5,prometheus
,,5,2018-05-22T19:53:00Z,x_duration_seconds,2,0.6,prometheus
,,6,2018-05-22T19:53:00Z,x_duration_seconds,2,0.7,prometheus
,,7,2018-05-22T19:53:00Z,x_duration_seconds,8,0.8,prometheus
,,8,2018-05-22T19:53:00Z,x_duration_seconds,10,0.9,prometheus
,,9,2018-05-22T19:53:00Z,x_duration_seconds,10,+Inf,prometheus
,,10,2018-05-22T19:53:00Z,y_duration_seconds,10,0.2,prometheus
,,11,2018-05-22T19:53:00Z,y_duration_seconds,15,0.4,prometheus
,,12,2018-05-22T19:53:00Z,y_duration_seconds,25,0.6,prometheus
,,13,2018-05-22T19:53:00Z,y_duration_seconds,35,0.8,prometheus
,,14,2018-05-22T19:53:00Z,y_duration_seconds,45,1,prometheus
,,15,2018-05-22T19:53:00Z,y_duration_seconds,45,+Inf,prometheus
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,double
#group,false,false,true,true,true,true,false
#default,_result,,,,,,
,result,table,_start,_stop,_field,_measurement,_value
,,0,2018-05-22T19:53:00Z,2030-01-01T00:00:00Z,x_duration_seconds,prometheus,0.8500000000000001
,,1,2018-05-22T19:53:00Z,2030-01-01T00:00:00Z,y_duration_seconds,prometheus,0.91
"
t_histogram_quantile = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:00Z)
		|> group(mode: "except", columns: ["le", "_time", "_value"])
		|> promql.promHistogramQuantile(quantile: 0.9))

test _histogram_quantile = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_histogram_quantile})
