package influxql_test

import "testing"
import "internal/influxql"

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,0,,,,,
,result,table,_time,_measurement,_field,_value
,,0,1970-01-01T00:00:00Z,m_0,n,0
,,1,1970-01-01T00:00:00.000000001Z,m_1,n,1
,,2,1970-01-01T00:00:00.00000001Z,m_10,n,10
,,3,1970-01-01T00:00:00.000000011Z,m_11,n,11
,,4,1970-01-01T00:00:00.000000012Z,m_12,n,12
,,5,1970-01-01T00:00:00.000000013Z,m_13,n,13
,,6,1970-01-01T00:00:00.000000014Z,m_14,n,14
,,7,1970-01-01T00:00:00.000000015Z,m_15,n,15
,,8,1970-01-01T00:00:00.000000016Z,m_16,n,16
,,9,1970-01-01T00:00:00.000000017Z,m_17,n,17
,,10,1970-01-01T00:00:00.000000018Z,m_18,n,18
,,11,1970-01-01T00:00:00.000000019Z,m_19,n,19
,,12,1970-01-01T00:00:00.000000002Z,m_2,n,2
,,13,1970-01-01T00:00:00.000000003Z,m_3,n,3
,,14,1970-01-01T00:00:00.000000004Z,m_4,n,4
,,15,1970-01-01T00:00:00.000000005Z,m_5,n,5
,,16,1970-01-01T00:00:00.000000006Z,m_6,n,6
,,17,1970-01-01T00:00:00.000000007Z,m_7,n,7
,,18,1970-01-01T00:00:00.000000008Z,m_8,n,8
,,19,1970-01-01T00:00:00.000000009Z,m_9,n,9
"

outData = "
#datatype,string,long,dateTime:RFC3339,string,double
#group,false,false,false,true,false
#default,0,,,,
,result,table,time,_measurement,n
,,0,1970-01-01T00:00:00Z,m_0,0
,,1,1970-01-01T00:00:00.000000001Z,m_1,1
,,2,1970-01-01T00:00:00.00000001Z,m_10,10
,,3,1970-01-01T00:00:00.000000011Z,m_11,11
,,4,1970-01-01T00:00:00.000000012Z,m_12,12
,,5,1970-01-01T00:00:00.000000013Z,m_13,13
,,6,1970-01-01T00:00:00.000000014Z,m_14,14
,,7,1970-01-01T00:00:00.000000015Z,m_15,15
,,8,1970-01-01T00:00:00.000000016Z,m_16,16
,,9,1970-01-01T00:00:00.000000017Z,m_17,17
,,10,1970-01-01T00:00:00.000000018Z,m_18,18
,,11,1970-01-01T00:00:00.000000019Z,m_19,19
,,12,1970-01-01T00:00:00.000000002Z,m_2,2
,,13,1970-01-01T00:00:00.000000003Z,m_3,3
,,14,1970-01-01T00:00:00.000000004Z,m_4,4
,,15,1970-01-01T00:00:00.000000005Z,m_5,5
,,16,1970-01-01T00:00:00.000000006Z,m_6,6
,,17,1970-01-01T00:00:00.000000007Z,m_7,7
,,18,1970-01-01T00:00:00.000000008Z,m_8,8
,,19,1970-01-01T00:00:00.000000009Z,m_9,9
"

// SELECT n FROM /^m/
t_regex_measurement = (tables=<-) => tables
	|> range(start: influxql.minTime, stop: influxql.maxTime)
	|> filter(fn: (r) => r._measurement =~ /^m/)
	|> filter(fn: (r) => r._field == "n")
	|> group(columns: ["_measurement", "_field"])
	|> sort(columns: ["_time"])
	|> keep(columns: ["_time", "_value", "_measurement"])
	|> rename(columns: {_time: "time", _value: "n"})

test _regex_measurement = () => ({
	input: testing.loadStorage(csv: inData),
	want: testing.loadMem(csv: outData),
	fn: t_regex_measurement,
})
