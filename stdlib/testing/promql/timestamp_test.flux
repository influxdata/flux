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
,,0,2018-12-01T20:00:10.123Z,metric_name,0,prometheus
,,0,2018-12-02T20:00:20.123Z,metric_name,0,prometheus
,,0,2018-12-03T20:00:30.123Z,metric_name,0,prometheus
,,0,2018-12-29T20:00:40.123Z,metric_name,0,prometheus
,,0,2018-12-30T20:00:50.123Z,metric_name,0,prometheus
,,0,2018-12-31T20:01:00.123Z,metric_name,0,prometheus
"
outData = "
#datatype,string,long,dateTime:RFC3339,string,double,string
#group,false,false,false,true,false,true
#default,outData,,,,,
,result,table,_time,_field,_value,_measurement
,,0,2018-12-01T20:00:10.123Z,metric_name,1543694410.123,prometheus
,,0,2018-12-02T20:00:20.123Z,metric_name,1543780820.123,prometheus
,,0,2018-12-03T20:00:30.123Z,metric_name,1543867230.123,prometheus
,,0,2018-12-29T20:00:40.123Z,metric_name,1546113640.123,prometheus
,,0,2018-12-30T20:00:50.123Z,metric_name,1546200050.123,prometheus
,,0,2018-12-31T20:01:00.123Z,metric_name,1546286460.123,prometheus
"
t_timestamp = (table=<-) =>
	(table
		|> range(start: 1980-01-01T00:00:00Z)
		|> drop(columns: ["_start", "_stop"])
		|> promql.timestamp())

test _timestamp = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_timestamp})
