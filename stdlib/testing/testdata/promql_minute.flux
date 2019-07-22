package testdata_test

import "testing"
import "promql"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,double
#group,false,true,false,true,false
#default,_result,,,,
,result,table,_time,_field,_value
,,0,2018-01-03T00:00:00Z,metric_name,0
,,0,2018-02-04T00:01:00Z,metric_name,0
,,0,2018-03-05T00:02:00Z,metric_name,0
,,0,2018-10-06T00:57:00Z,metric_name,0
,,0,2018-11-07T00:58:00Z,metric_name,0
,,0,2018-12-08T00:59:00Z,metric_name,0
"

outData = "
#datatype,string,long,string,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double
#group,false,false,true,true,true,false,false
#default,got,,,,,,
,result,table,_field,_start,_stop,_time,_value
,,0,metric_name,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,2018-01-03T00:00:00Z,0
,,0,metric_name,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,2018-02-04T00:01:00Z,1
,,0,metric_name,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,2018-03-05T00:02:00Z,2
,,0,metric_name,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,2018-10-06T00:57:00Z,57
,,0,metric_name,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,2018-11-07T00:58:00Z,58
,,0,metric_name,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,2018-12-08T00:59:00Z,59
"

t_promqlMinute = (table=<-) =>
	(table
	  |> range(start: 2018-01-01T00:00:00Z)
	  |> promql.timestamp()
	  |> map(fn: (r) => ({r with _value: promql.promqlMinute(timestamp: r._value)})))

test _promqlMinute = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_promqlMinute})
