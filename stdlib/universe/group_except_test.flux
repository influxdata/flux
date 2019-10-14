package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,name
,,0,2018-05-22T19:53:26Z,15204688,io_time,diskio1,host.local,disk0
,,0,2018-05-22T19:53:36Z,15204894,io_time,diskio1,host.local,disk0
,,0,2018-05-22T19:53:46Z,15205102,io_time,diskio1,host.local,disk0
,,0,2018-05-22T19:53:56Z,15205226,io_time,diskio1,host.local,disk0
,,0,2018-05-22T19:54:06Z,15205499,io_time,diskio2,host.local,disk0
,,0,2018-05-22T19:54:16Z,15205755,io_time,diskio2,host.local,disk0
,,1,2018-05-22T19:53:26Z,648,io_time,diskio2,host.local,disk2
,,1,2018-05-22T19:53:36Z,648,io_time,diskio2,host.local,disk2
,,1,2018-05-22T19:53:46Z,648,io_time,diskio3,host.local,disk2
,,1,2018-05-22T19:53:56Z,648,io_time,diskio3,host.local,disk2
,,1,2018-05-22T19:54:06Z,648,io_time,diskio3,host.local,disk2
,,1,2018-05-22T19:54:16Z,648,io_time,diskio3,host.local,disk2
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#group,false,false,true,true,false,false,true,false,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,15205755,io_time,diskio2,host.local,disk0
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,648,io_time,diskio2,host.local,disk2
"

t_group_except = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> group(columns: ["_measurement", "_time", "_value"], mode: "except")
		|> max())

test _group_except = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_group_except})

