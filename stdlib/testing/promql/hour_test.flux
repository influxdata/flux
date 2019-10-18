package promql_test
import "testing"
import "internal/promql"

option now = () =>
	(2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,double,string
#group,false,false,false,true,false,true
#default,inData,,,,,
,result,table,_time,_field,_value,_measurement
,,0,2018-01-03T00:00:00Z,metric_name,0,prometheus
,,0,2018-02-04T01:00:00Z,metric_name,0,prometheus
,,0,2018-03-05T02:00:00Z,metric_name,0,prometheus
,,0,2018-10-06T21:00:00Z,metric_name,0,prometheus
,,0,2018-11-07T22:00:00Z,metric_name,0,prometheus
,,0,2018-12-08T23:00:00Z,metric_name,0,prometheus
"
outData = "
#datatype,string,long,dateTime:RFC3339,string,double,string
#group,false,false,false,true,false,true
#default,outData,,,,,
,result,table,_time,_field,_value,_measurement
,,0,2018-01-03T00:00:00Z,metric_name,0,prometheus
,,0,2018-02-04T01:00:00Z,metric_name,1,prometheus
,,0,2018-03-05T02:00:00Z,metric_name,2,prometheus
,,0,2018-10-06T21:00:00Z,metric_name,21,prometheus
,,0,2018-11-07T22:00:00Z,metric_name,22,prometheus
,,0,2018-12-08T23:00:00Z,metric_name,23,prometheus
"
t_promqlHour = (table=<-) =>
	(table
		|> range(start: 1980-01-01T00:00:00Z)
		|> drop(columns: ["_start", "_stop"])
		|> promql.timestamp()
		|> map(fn: (r) =>
			({r with _value: promql.promqlHour(timestamp: r._value)})))

test _promqlHour = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_promqlHour})
