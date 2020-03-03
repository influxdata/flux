package influxql_test

import "testing"
import "internal/influxql"

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,0,,,,,
,result,table,_time,_measurement,_field,_value
,,0,1970-01-01T00:00:00Z,ctr,n,0
,,0,1970-01-01T00:00:00.000000001Z,ctr,n,1
,,0,1970-01-01T00:00:00.000000002Z,ctr,n,2
,,0,1970-01-01T00:00:00.000000003Z,ctr,n,3
,,0,1970-01-01T00:00:00.000000004Z,ctr,n,4
,,0,1970-01-01T00:00:00.000000005Z,ctr,n,5
,,0,1970-01-01T00:00:00.000000006Z,ctr,n,6
,,0,1970-01-01T00:00:00.000000007Z,ctr,n,7
,,0,1970-01-01T00:00:00.000000008Z,ctr,n,8
,,0,1970-01-01T00:00:00.000000009Z,ctr,n,9
,,0,1970-01-01T00:00:00.00000001Z,ctr,n,10
,,0,1970-01-01T00:00:00.000000011Z,ctr,n,11
,,0,1970-01-01T00:00:00.000000012Z,ctr,n,12
,,0,1970-01-01T00:00:00.000000013Z,ctr,n,13
,,0,1970-01-01T00:00:00.000000014Z,ctr,n,14
,,0,1970-01-01T00:00:00.000000015Z,ctr,n,15
,,0,1970-01-01T00:00:00.000000016Z,ctr,n,16
,,0,1970-01-01T00:00:00.000000017Z,ctr,n,17
,,0,1970-01-01T00:00:00.000000018Z,ctr,n,18
,,0,1970-01-01T00:00:00.000000019Z,ctr,n,19
"

outData = "
#datatype,string,long,dateTime:RFC3339,string,double
#group,false,false,false,true,false
#default,0,,,,
,result,table,time,_measurement,n
,,0,1970-01-01T00:00:00.000000008Z,ctr,8
,,0,1970-01-01T00:00:00.000000009Z,ctr,9
,,0,1970-01-01T00:00:00.00000001Z,ctr,10
,,0,1970-01-01T00:00:00.000000011Z,ctr,11
,,0,1970-01-01T00:00:00.000000012Z,ctr,12
,,0,1970-01-01T00:00:00.000000013Z,ctr,13
,,0,1970-01-01T00:00:00.000000014Z,ctr,14
"

// SELECT n FROM ctr WHERE n >= 8 AND n <= 14
t_filter_by_values_with_and = (tables=<-) => tables
	|> range(start: influxql.minTime, stop: influxql.maxTime)
	|> filter(fn: (r) => r._measurement == "ctr")
	|> filter(fn: (r) => r._field == "n")
	|> filter(fn: (r) => r._value >= 8 and r._value <= 14)
	|> group(columns: ["_measurement", "_field"])
	|> sort(columns: ["_time"])
	|> keep(columns: ["_time", "_value", "_measurement"])
	|> rename(columns: {_time: "time", _value: "n"})

test _filter_by_values_with_and = () => ({
	input: testing.loadStorage(csv: inData),
	want: testing.loadMem(csv: outData),
	fn: t_filter_by_values_with_and,
})
