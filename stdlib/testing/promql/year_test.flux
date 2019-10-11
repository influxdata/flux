package promql_test

import "testing"
import "internal/promql"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,double
#group,false,true,false,true,false
#default,_result,,,,
,result,table,_time,_field,_value
,,0,2018-01-03T20:00:00Z,metric_name,0
,,0,2019-02-04T20:00:00Z,metric_name,0
,,0,2020-03-05T20:00:00Z,metric_name,0
,,0,2021-10-06T20:00:00Z,metric_name,0
,,0,2022-11-07T20:00:00Z,metric_name,0
,,0,2023-12-08T20:00:00Z,metric_name,0
"

outData = "
#datatype,string,long,dateTime:RFC3339,string,double
#group,false,true,false,true,false
#default,_result,,,,
,result,table,_time,_field,_value
,,0,2018-01-03T20:00:00Z,metric_name,2018
,,0,2019-02-04T20:00:00Z,metric_name,2019
,,0,2020-03-05T20:00:00Z,metric_name,2020
,,0,2021-10-06T20:00:00Z,metric_name,2021
,,0,2022-11-07T20:00:00Z,metric_name,2022
,,0,2023-12-08T20:00:00Z,metric_name,2023
"

t_promqlYear = (table=<-) =>
	(table
	  |> promql.timestamp()
		|> map(fn: (r) => ({r with _value: promql.promqlYear(timestamp: r._value)})))

test _promqlYear = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_promqlYear})
