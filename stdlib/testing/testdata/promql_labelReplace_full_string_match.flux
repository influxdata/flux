package testdata_test

import "testing"
import "promql"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,true,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_field,src,dst,_value
,,1,2018-12-18T20:52:33Z,metric_name,source-value-10,original-destination-value,1
,,1,2018-12-18T20:52:43Z,metric_name,source-value-10,original-destination-value,2

#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,true,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_field,src,dst,_value
,,1,2018-12-18T20:52:33Z,metric_name,source-value-20,original-destination-value,3
,,1,2018-12-18T20:52:43Z,metric_name,source-value-20,original-destination-value,4
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double
#group,false,false,true,true,true,true,true,false
#default,got,,,,,,,
,result,table,_start,_stop,_field,src,dst,_value
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,metric_name,source-value-10,destination-value-10,1
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,metric_name,source-value-10,destination-value-10,2
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,metric_name,source-value-20,destination-value-20,3
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,metric_name,source-value-20,destination-value-20,4
"

t_labelReplace = (table=<-) =>
	(table
		|> range(start: 2018-01-01T00:00:00Z)
		|> promql.labelReplace(source: "src", destination: "dst", regex: "source-value-(.*)", replacement: "destination-value-$1"))

test _labelReplace = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_labelReplace})
