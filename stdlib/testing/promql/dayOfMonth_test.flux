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
,,0,2018-12-01T20:00:00Z,metric_name,0,prometheus
,,0,2018-12-02T20:00:00Z,metric_name,0,prometheus
,,0,2018-12-03T20:00:00Z,metric_name,0,prometheus
,,0,2018-12-29T20:00:00Z,metric_name,0,prometheus
,,0,2018-12-30T20:00:00Z,metric_name,0,prometheus
,,0,2018-12-31T20:00:00Z,metric_name,0,prometheus
"
outData = "
#datatype,string,long,dateTime:RFC3339,string,double,string
#group,false,false,false,true,false,true
#default,outData,,,,,
,result,table,_time,_field,_value,_measurement
,,0,2018-12-01T20:00:00Z,metric_name,1,prometheus
,,0,2018-12-02T20:00:00Z,metric_name,2,prometheus
,,0,2018-12-03T20:00:00Z,metric_name,3,prometheus
,,0,2018-12-29T20:00:00Z,metric_name,29,prometheus
,,0,2018-12-30T20:00:00Z,metric_name,30,prometheus
,,0,2018-12-31T20:00:00Z,metric_name,31,prometheus
"
t_promqlDayOfMonth = (table=<-) =>
	(table
		|> range(start: 1980-01-01T00:00:00Z)
		|> drop(columns: ["_start", "_stop"])
		|> promql.timestamp()
		|> map(fn: (r) =>
			({r with _value: promql.promqlDayOfMonth(timestamp: r._value)})))

test _promqlDayOfMonth = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_promqlDayOfMonth})
