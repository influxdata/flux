package promql_test

import "testing"
import "internal/promql"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,double
#group,false,true,false,true,false
#default,_result,,,,
,result,table,_time,_field,_value
,,0,2018-12-18T20:52:33Z,metric_name1,1
,,0,2018-12-18T20:52:43Z,metric_name1,1
,,0,2018-12-18T20:52:53Z,metric_name1,1
,,0,2018-12-18T20:53:03Z,metric_name1,1
,,0,2018-12-18T20:53:13Z,metric_name1,1
,,0,2018-12-18T20:53:23Z,metric_name1,1

#datatype,string,long,dateTime:RFC3339,string,double
#group,false,true,false,true,false
#default,_result,,,,
,result,table,_time,_field,_value
,,1,2018-12-18T20:52:33Z,metric_name2,1
,,1,2018-12-18T20:52:43Z,metric_name2,1
,,1,2018-12-18T20:52:53Z,metric_name2,1
,,1,2018-12-18T20:53:03Z,metric_name2,100
,,1,2018-12-18T20:53:13Z,metric_name2,100
,,1,2018-12-18T20:53:23Z,metric_name2,100

#datatype,string,long,dateTime:RFC3339,string,double
#group,false,true,false,true,false
#default,_result,,,,
,result,table,_time,_field,_value
,,2,2018-12-18T20:52:33Z,metric_name3,100
,,2,2018-12-18T20:52:43Z,metric_name3,200
,,2,2018-12-18T20:52:53Z,metric_name3,300
,,2,2018-12-18T20:53:03Z,metric_name3,200
,,2,2018-12-18T20:53:13Z,metric_name3,300
,,2,2018-12-18T20:53:23Z,metric_name3,400

#datatype,string,long,dateTime:RFC3339,string,double
#group,false,true,false,true,false
#default,_result,,,,
,result,table,_time,_field,_value
,,2,2018-12-18T20:52:33Z,metric_name4,100
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,double
#group,false,true,true,true,true,false
#default,_result,,,,,
,result,table,_start,_stop,_field,_value
,,0,2018-12-18T20:50:00Z,2018-12-18T20:55:00Z,metric_name1,1
,,1,2018-12-18T20:50:00Z,2018-12-18T20:55:00Z,metric_name2,32.78296
,,2,2018-12-18T20:50:00Z,2018-12-18T20:55:00Z,metric_name3,535.7920000000001
"

t_holtWinters = (table=<-) =>
	(table
		|> range(start: 2018-12-18T20:50:00Z, stop: 2018-12-18T20:55:00Z)
		|> promql.holtWinters(smoothingFactor: 0.1, trendFactor: 0.2))

test _holtWinters = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_holtWinters})
