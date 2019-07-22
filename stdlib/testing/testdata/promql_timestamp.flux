package testdata_test

import "testing"
import "promql"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,double
#group,false,true,false,true,false
#default,_result,,,,
,result,table,_time,_field,_value
,,0,2018-12-01T20:00:10.123Z,metric_name,0
,,0,2018-12-02T20:00:20.123Z,metric_name,0
,,0,2018-12-03T20:00:30.123Z,metric_name,0
,,0,2018-12-29T20:00:40.123Z,metric_name,0
,,0,2018-12-30T20:00:50.123Z,metric_name,0
,,0,2018-12-31T20:01:00.123Z,metric_name,0
"

outData = "
#datatype,string,long,dateTime:RFC3339,string,double
#group,false,true,false,true,false
#default,_result,,,,
,result,table,_time,_field,_value
,,0,2018-12-01T20:00:10.123Z,metric_name,1543694410.123
,,0,2018-12-02T20:00:20.123Z,metric_name,1543780820.123
,,0,2018-12-03T20:00:30.123Z,metric_name,1543867230.123
,,0,2018-12-29T20:00:40.123Z,metric_name,1546113640.123
,,0,2018-12-30T20:00:50.123Z,metric_name,1546200050.123
,,0,2018-12-31T20:01:00.123Z,metric_name,1546286460.123
"

t_timestamp = (table=<-) =>
	(table
	  |> promql.timestamp())

test _timestamp = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_timestamp})
