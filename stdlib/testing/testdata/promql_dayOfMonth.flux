package testdata_test

import "testing"
import "promql"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,double
#group,false,true,false,true,false
#default,_result,,,,
,result,table,_time,_field,_value
,,0,2018-12-01T20:00:00Z,metric_name,0
,,0,2018-12-02T20:00:00Z,metric_name,0
,,0,2018-12-03T20:00:00Z,metric_name,0
,,0,2018-12-29T20:00:00Z,metric_name,0
,,0,2018-12-30T20:00:00Z,metric_name,0
,,0,2018-12-31T20:00:00Z,metric_name,0
"

outData = "
#datatype,string,long,string,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double
#group,false,false,true,true,true,false,false
#default,got,,,,,,
,result,table,_field,_start,_stop,_time,_value
,,0,metric_name,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,2018-12-01T20:00:00Z,1
,,0,metric_name,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,2018-12-02T20:00:00Z,2
,,0,metric_name,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,2018-12-03T20:00:00Z,3
,,0,metric_name,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,2018-12-29T20:00:00Z,29
,,0,metric_name,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,2018-12-30T20:00:00Z,30
,,0,metric_name,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,2018-12-31T20:00:00Z,31
"

t_promqlDayOfMonth = (table=<-) =>
	(table
	  |> range(start: 2018-01-01T00:00:00Z)
	  |> promql.timestamp()
		|> map(fn: (r) => ({r with _value: promql.promqlDayOfMonth(timestamp: r._value)})))

test _promqlDayOfMonth = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_promqlDayOfMonth})
