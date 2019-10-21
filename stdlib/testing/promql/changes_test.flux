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
,,0,2018-12-18T20:52:33Z,metric_name1,1,prometheus
,,0,2018-12-18T20:52:43Z,metric_name1,1,prometheus
,,0,2018-12-18T20:52:53Z,metric_name1,1,prometheus
,,0,2018-12-18T20:53:03Z,metric_name1,1,prometheus
,,0,2018-12-18T20:53:13Z,metric_name1,1,prometheus
,,0,2018-12-18T20:53:23Z,metric_name1,1,prometheus
,,1,2018-12-18T20:52:33Z,metric_name2,1,prometheus
,,1,2018-12-18T20:52:43Z,metric_name2,1,prometheus
,,1,2018-12-18T20:52:53Z,metric_name2,1,prometheus
,,1,2018-12-18T20:53:03Z,metric_name2,2,prometheus
,,1,2018-12-18T20:53:13Z,metric_name2,2,prometheus
,,1,2018-12-18T20:53:23Z,metric_name2,2,prometheus
,,2,2018-12-18T20:52:33Z,metric_name3,1,prometheus
,,2,2018-12-18T20:52:43Z,metric_name3,2,prometheus
,,2,2018-12-18T20:52:53Z,metric_name3,3,prometheus
,,2,2018-12-18T20:53:03Z,metric_name3,1,prometheus
,,2,2018-12-18T20:53:13Z,metric_name3,2,prometheus
,,2,2018-12-18T20:53:23Z,metric_name3,3,prometheus
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,double,string
#group,false,false,true,true,true,false,true
#default,outData,,,,,,
,result,table,_start,_stop,_field,_value,_measurement
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,metric_name1,0,prometheus
,,1,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,metric_name2,1,prometheus
,,2,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,metric_name3,5,prometheus
"
t_changes = (table=<-) =>
	(table
		|> range(start: 2018-12-01T00:00:00Z)
		|> promql.changes())

test _changes = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_changes})
