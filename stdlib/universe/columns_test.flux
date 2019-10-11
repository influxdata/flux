package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,false,false,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,name
,,0,2018-05-22T19:53:26Z,15204688,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:36Z,15204894,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:46Z,15205102,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:56Z,15205226,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:54:06Z,15205499,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:54:16Z,15205755,io_time,diskio,host.local,disk0
,,1,2018-05-22T19:53:26Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:36Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:46Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:56Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:06Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:16Z,648,io_time,diskio,host.local,disk2
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string
#group,false,false,true,true,true,true,false
#default,_result,,,,,,
,result,table,_start,_stop,host,name,_value
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk0,_start
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk0,_stop
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk0,_time
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk0,_value
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk0,_field
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk0,_measurement
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk0,host
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk0,name
,,1,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk2,_start
,,1,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk2,_stop
,,1,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk2,_time
,,1,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk2,_value
,,1,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk2,_field
,,1,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk2,_measurement
,,1,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk2,host
,,1,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,host.local,disk2,name
"

t_columns = (table=<-) =>
	(table
		|> range(start: 2018-05-20T19:53:26Z)
		|> columns())

test _columns = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_columns})

