package testdata_test

import "testing"
import "promql"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,double
#group,false,true,false,true,false
#default,_result,,,,
,result,table,_time,_field,_value
,,0,2018-12-03T20:00:00Z,metric_name,0
,,0,2018-12-04T20:00:00Z,metric_name,0
,,0,2018-12-05T20:00:00Z,metric_name,0
,,0,2018-12-06T20:00:00Z,metric_name,0
,,0,2018-12-07T20:00:00Z,metric_name,0
,,0,2018-12-08T20:00:00Z,metric_name,0
,,0,2018-12-09T20:00:00Z,metric_name,0
"

outData = "
#datatype,string,long,dateTime:RFC3339,string,double
#group,false,true,false,true,false
#default,_result,,,,
,result,table,_time,_field,_value
,,0,2018-12-03T20:00:00Z,metric_name,1
,,0,2018-12-04T20:00:00Z,metric_name,2
,,0,2018-12-05T20:00:00Z,metric_name,3
,,0,2018-12-06T20:00:00Z,metric_name,4
,,0,2018-12-07T20:00:00Z,metric_name,5
,,0,2018-12-08T20:00:00Z,metric_name,6
,,0,2018-12-09T20:00:00Z,metric_name,0
"

t_promqlDayOfWeek = (table=<-) =>
	(table
	  |> promql.timestamp()
		|> map(fn: (r) => ({r with _value: promql.promqlDayOfWeek(timestamp: r._value)})))

test _promqlDayOfWeek = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_promqlDayOfWeek})
