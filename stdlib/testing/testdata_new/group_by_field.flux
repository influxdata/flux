package main
import "testing"

option now = () =>
	(2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,91.7364670583823,usage_idle,cpu,cpu-total,host1
,,0,2018-05-22T19:53:36Z,89.51118889861233,usage_idle,cpu,cpu-total,host2
,,0,2018-05-22T19:53:46Z,91.0977744436109,usage_idle,cpu,cpu-total,host1
,,0,2018-05-22T19:53:56Z,91.02836436336374,usage_idle,cpu,cpu-total,host2
,,0,2018-05-22T19:54:06Z,68.304576144036,usage_idle,cpu,cpu-total,host1
,,0,2018-05-22T19:54:16Z,87.88598574821853,usage_idle,cpu,cpu-total,host2
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,false,true,false,false,false,false
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,68.304576144036,usage_idle,cpu,cpu-total,host1
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,87.88598574821853,usage_idle,cpu,cpu-total,host2
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,89.51118889861233,usage_idle,cpu,cpu-total,host2
,,3,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,91.02836436336374,usage_idle,cpu,cpu-total,host2
,,4,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,91.0977744436109,usage_idle,cpu,cpu-total,host1
,,5,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,91.7364670583823,usage_idle,cpu,cpu-total,host1
"
t_group_by_field = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> group(columns: ["_value"]))

test group_by_field = {input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_group_by_field}