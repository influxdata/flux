package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,true,true,true,true
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
#datatype,string,long,dateTime:RFC3339,long
#group,false,false,false,false
#default,0,,,
,result,table,_time,io_time
,,0,2018-05-22T19:53:26Z,15204688
,,0,2018-05-22T19:53:36Z,15204894
,,0,2018-05-22T19:53:46Z,15205102
,,0,2018-05-22T19:53:56Z,15205226
,,0,2018-05-22T19:54:06Z,15205499
,,0,2018-05-22T19:54:16Z,15205755
,,0,2018-05-22T19:53:26Z,648
,,0,2018-05-22T19:53:36Z,648
,,0,2018-05-22T19:53:46Z,648
,,0,2018-05-22T19:53:56Z,648
,,0,2018-05-22T19:54:06Z,648
,,0,2018-05-22T19:54:16Z,648
"

t_group_ungroup = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> group(columns: ["name"])
		|> group()
		|> map(fn: (r) =>
			({_time: r._time, io_time: r._value}))
		|> yield(name: "0"))

test _group_ungroup = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_group_ungroup})

