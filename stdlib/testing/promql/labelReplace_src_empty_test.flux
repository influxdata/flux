package promql_test
import "testing"
import "internal/promql"

option now = () =>
	(2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,double,string
#group,false,false,false,true,true,true,false,true
#default,inData,,,,,,,
,result,table,_time,_field,src,dst,_value,_measurement
,,0,2018-12-18T20:52:33Z,metric_name,source-value-10,original-destination-value,1,prometheus
,,0,2018-12-18T20:52:43Z,metric_name,source-value-10,original-destination-value,2,prometheus
,,1,2018-12-18T20:52:33Z,metric_name,source-value-20,original-destination-value,3,prometheus
,,1,2018-12-18T20:52:43Z,metric_name,source-value-20,original-destination-value,4,prometheus
"
outData = "
#datatype,string,long,string,string,string,double,string
#group,false,false,true,true,true,false,true
#default,outData,,,,,,
,result,table,_field,src,dst,_value,_measurement
,,0,metric_name,source-value-10,value-,1,prometheus
,,0,metric_name,source-value-10,value-,2,prometheus
,,1,metric_name,source-value-20,value-,3,prometheus
,,1,metric_name,source-value-20,value-,4,prometheus
"
t_labelReplace = (table=<-) =>
	(table
		|> range(start: 1980-01-01T00:00:00Z)
		|> drop(columns: ["_start", "_stop"])
		|> promql.labelReplace(
			source: "nonexistent-src",
			destination: "dst",
			regex: "(.*)",
			replacement: "value-$1",
		))

test _labelReplace = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_labelReplace})
