package main
// 
import "testing"

option now = () =>
	(2030-01-01T00:00:00Z)

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
,,2,2018-05-22T19:53:26Z,3929,io_time,diskio,host.local,disk3
,,2,2018-05-22T19:53:36Z,3929,io_time,diskio,host.local,disk3
,,2,2018-05-22T19:53:46Z,3929,io_time,diskio,host.local,disk3
,,2,2018-05-22T19:53:56Z,3929,io_time,diskio,host.local,disk3
,,2,2018-05-22T19:54:06Z,3929,io_time,diskio,host.local,disk3
,,2,2018-05-22T19:54:16Z,3929,io_time,diskio,host.local,disk3
,,3,2018-05-22T19:53:26Z,0,iops_in_progress,diskio,host.local,disk0
,,3,2018-05-22T19:53:36Z,0,iops_in_progress,diskio,host.local,disk0
,,3,2018-05-22T19:53:46Z,0,iops_in_progress,diskio,host.local,disk0
,,3,2018-05-22T19:53:56Z,0,iops_in_progress,diskio,host.local,disk0
,,3,2018-05-22T19:54:06Z,0,iops_in_progress,diskio,host.local,disk0
,,3,2018-05-22T19:54:16Z,0,iops_in_progress,diskio,host.local,disk0
,,4,2018-05-22T19:53:26Z,0,iops_in_progress,diskio,host.local,disk2
,,4,2018-05-22T19:53:36Z,0,iops_in_progress,diskio,host.local,disk2
,,4,2018-05-22T19:53:46Z,0,iops_in_progress,diskio,host.local,disk2
,,4,2018-05-22T19:53:56Z,0,iops_in_progress,diskio,host.local,disk2
,,4,2018-05-22T19:54:06Z,0,iops_in_progress,diskio,host.local,disk2
,,4,2018-05-22T19:54:16Z,0,iops_in_progress,diskio,host.local,disk2
,,5,2018-05-22T19:53:26Z,0,iops_in_progress,diskio,host.local,disk3
,,5,2018-05-22T19:53:36Z,0,iops_in_progress,diskio,host.local,disk3
,,5,2018-05-22T19:53:46Z,0,iops_in_progress,diskio,host.local,disk3
,,5,2018-05-22T19:53:56Z,0,iops_in_progress,diskio,host.local,disk3
,,5,2018-05-22T19:54:06Z,0,iops_in_progress,diskio,host.local,disk3
,,5,2018-05-22T19:54:16Z,0,iops_in_progress,diskio,host.local,disk3
,,6,2018-05-22T19:53:26Z,228569833472,read_bytes,diskio,host.local,disk0
,,6,2018-05-22T19:53:36Z,228577058816,read_bytes,diskio,host.local,disk0
,,6,2018-05-22T19:53:46Z,228583690240,read_bytes,diskio,host.local,disk0
,,6,2018-05-22T19:53:56Z,228585836544,read_bytes,diskio,host.local,disk0
,,6,2018-05-22T19:54:06Z,228594925568,read_bytes,diskio,host.local,disk0
,,6,2018-05-22T19:54:16Z,228613324800,read_bytes,diskio,host.local,disk0
,,7,2018-05-22T19:53:26Z,202997248,read_bytes,diskio,host.local,disk2
,,7,2018-05-22T19:53:36Z,202997248,read_bytes,diskio,host.local,disk2
,,7,2018-05-22T19:53:46Z,202997248,read_bytes,diskio,host.local,disk2
,,7,2018-05-22T19:53:56Z,202997248,read_bytes,diskio,host.local,disk2
,,7,2018-05-22T19:54:06Z,202997248,read_bytes,diskio,host.local,disk2
,,7,2018-05-22T19:54:16Z,202997248,read_bytes,diskio,host.local,disk2
,,8,2018-05-22T19:53:26Z,24209920,read_bytes,diskio,host.local,disk3
,,8,2018-05-22T19:53:36Z,24209920,read_bytes,diskio,host.local,disk3
,,8,2018-05-22T19:53:46Z,24209920,read_bytes,diskio,host.local,disk3
,,8,2018-05-22T19:53:56Z,24209920,read_bytes,diskio,host.local,disk3
,,8,2018-05-22T19:54:06Z,24209920,read_bytes,diskio,host.local,disk3
,,8,2018-05-22T19:54:16Z,24209920,read_bytes,diskio,host.local,disk3
,,9,2018-05-22T19:53:26Z,3455708,read_time,diskio,host.local,disk0
,,9,2018-05-22T19:53:36Z,3455784,read_time,diskio,host.local,disk0
,,9,2018-05-22T19:53:46Z,3455954,read_time,diskio,host.local,disk0
,,9,2018-05-22T19:53:56Z,3456031,read_time,diskio,host.local,disk0
,,9,2018-05-22T19:54:06Z,3456184,read_time,diskio,host.local,disk0
,,9,2018-05-22T19:54:16Z,3456390,read_time,diskio,host.local,disk0
,,10,2018-05-22T19:53:26Z,648,read_time,diskio,host.local,disk2
,,10,2018-05-22T19:53:36Z,648,read_time,diskio,host.local,disk2
,,10,2018-05-22T19:53:46Z,648,read_time,diskio,host.local,disk2
,,10,2018-05-22T19:53:56Z,648,read_time,diskio,host.local,disk2
,,10,2018-05-22T19:54:06Z,648,read_time,diskio,host.local,disk2
,,10,2018-05-22T19:54:16Z,648,read_time,diskio,host.local,disk2
,,11,2018-05-22T19:53:26Z,3929,read_time,diskio,host.local,disk3
,,11,2018-05-22T19:53:36Z,3929,read_time,diskio,host.local,disk3
,,11,2018-05-22T19:53:46Z,3929,read_time,diskio,host.local,disk3
,,11,2018-05-22T19:53:56Z,3929,read_time,diskio,host.local,disk3
,,11,2018-05-22T19:54:06Z,3929,read_time,diskio,host.local,disk3
,,11,2018-05-22T19:54:16Z,3929,read_time,diskio,host.local,disk3
,,12,2018-05-22T19:53:26Z,6129420,reads,diskio,host.local,disk0
,,12,2018-05-22T19:53:36Z,6129483,reads,diskio,host.local,disk0
,,12,2018-05-22T19:53:46Z,6129801,reads,diskio,host.local,disk0
,,12,2018-05-22T19:53:56Z,6129864,reads,diskio,host.local,disk0
,,12,2018-05-22T19:54:06Z,6130176,reads,diskio,host.local,disk0
,,12,2018-05-22T19:54:16Z,6130461,reads,diskio,host.local,disk0
,,13,2018-05-22T19:53:26Z,729,reads,diskio,host.local,disk2
,,13,2018-05-22T19:53:36Z,729,reads,diskio,host.local,disk2
,,13,2018-05-22T19:53:46Z,729,reads,diskio,host.local,disk2
,,13,2018-05-22T19:53:56Z,729,reads,diskio,host.local,disk2
,,13,2018-05-22T19:54:06Z,729,reads,diskio,host.local,disk2
,,13,2018-05-22T19:54:16Z,729,reads,diskio,host.local,disk2
,,14,2018-05-22T19:53:26Z,100,reads,diskio,host.local,disk3
,,14,2018-05-22T19:53:36Z,100,reads,diskio,host.local,disk3
,,14,2018-05-22T19:53:46Z,100,reads,diskio,host.local,disk3
,,14,2018-05-22T19:53:56Z,100,reads,diskio,host.local,disk3
,,14,2018-05-22T19:54:06Z,100,reads,diskio,host.local,disk3
,,14,2018-05-22T19:54:16Z,100,reads,diskio,host.local,disk3
,,15,2018-05-22T19:53:26Z,0,weighted_io_time,diskio,host.local,disk0
,,15,2018-05-22T19:53:36Z,0,weighted_io_time,diskio,host.local,disk0
,,15,2018-05-22T19:53:46Z,0,weighted_io_time,diskio,host.local,disk0
,,15,2018-05-22T19:53:56Z,0,weighted_io_time,diskio,host.local,disk0
,,15,2018-05-22T19:54:06Z,0,weighted_io_time,diskio,host.local,disk0
,,15,2018-05-22T19:54:16Z,0,weighted_io_time,diskio,host.local,disk0
,,16,2018-05-22T19:53:26Z,0,weighted_io_time,diskio,host.local,disk2
,,16,2018-05-22T19:53:36Z,0,weighted_io_time,diskio,host.local,disk2
,,16,2018-05-22T19:53:46Z,0,weighted_io_time,diskio,host.local,disk2
,,16,2018-05-22T19:53:56Z,0,weighted_io_time,diskio,host.local,disk2
,,16,2018-05-22T19:54:06Z,0,weighted_io_time,diskio,host.local,disk2
,,16,2018-05-22T19:54:16Z,0,weighted_io_time,diskio,host.local,disk2
,,17,2018-05-22T19:53:26Z,0,weighted_io_time,diskio,host.local,disk3
,,17,2018-05-22T19:53:36Z,0,weighted_io_time,diskio,host.local,disk3
,,17,2018-05-22T19:53:46Z,0,weighted_io_time,diskio,host.local,disk3
,,17,2018-05-22T19:53:56Z,0,weighted_io_time,diskio,host.local,disk3
,,17,2018-05-22T19:54:06Z,0,weighted_io_time,diskio,host.local,disk3
,,17,2018-05-22T19:54:16Z,0,weighted_io_time,diskio,host.local,disk3
,,18,2018-05-22T19:53:26Z,373674287104,write_bytes,diskio,host.local,disk0
,,18,2018-05-22T19:53:36Z,373675814912,write_bytes,diskio,host.local,disk0
,,18,2018-05-22T19:53:46Z,373676670976,write_bytes,diskio,host.local,disk0
,,18,2018-05-22T19:53:56Z,373676830720,write_bytes,diskio,host.local,disk0
,,18,2018-05-22T19:54:06Z,373677928448,write_bytes,diskio,host.local,disk0
,,18,2018-05-22T19:54:16Z,373684617216,write_bytes,diskio,host.local,disk0
,,19,2018-05-22T19:53:26Z,0,write_bytes,diskio,host.local,disk2
,,19,2018-05-22T19:53:36Z,0,write_bytes,diskio,host.local,disk2
,,19,2018-05-22T19:53:46Z,0,write_bytes,diskio,host.local,disk2
,,19,2018-05-22T19:53:56Z,0,write_bytes,diskio,host.local,disk2
,,19,2018-05-22T19:54:06Z,0,write_bytes,diskio,host.local,disk2
,,19,2018-05-22T19:54:16Z,0,write_bytes,diskio,host.local,disk2
,,20,2018-05-22T19:53:26Z,0,write_bytes,diskio,host.local,disk3
,,20,2018-05-22T19:53:36Z,0,write_bytes,diskio,host.local,disk3
,,20,2018-05-22T19:53:46Z,0,write_bytes,diskio,host.local,disk3
,,20,2018-05-22T19:53:56Z,0,write_bytes,diskio,host.local,disk3
,,20,2018-05-22T19:54:06Z,0,write_bytes,diskio,host.local,disk3
,,20,2018-05-22T19:54:16Z,0,write_bytes,diskio,host.local,disk3
,,21,2018-05-22T19:53:26Z,11748980,write_time,diskio,host.local,disk0
,,21,2018-05-22T19:53:36Z,11749110,write_time,diskio,host.local,disk0
,,21,2018-05-22T19:53:46Z,11749148,write_time,diskio,host.local,disk0
,,21,2018-05-22T19:53:56Z,11749195,write_time,diskio,host.local,disk0
,,21,2018-05-22T19:54:06Z,11749315,write_time,diskio,host.local,disk0
,,21,2018-05-22T19:54:16Z,11749365,write_time,diskio,host.local,disk0
,,22,2018-05-22T19:53:26Z,0,write_time,diskio,host.local,disk2
,,22,2018-05-22T19:53:36Z,0,write_time,diskio,host.local,disk2
,,22,2018-05-22T19:53:46Z,0,write_time,diskio,host.local,disk2
,,22,2018-05-22T19:53:56Z,0,write_time,diskio,host.local,disk2
,,22,2018-05-22T19:54:06Z,0,write_time,diskio,host.local,disk2
,,22,2018-05-22T19:54:16Z,0,write_time,diskio,host.local,disk2
,,23,2018-05-22T19:53:26Z,0,write_time,diskio,host.local,disk3
,,23,2018-05-22T19:53:36Z,0,write_time,diskio,host.local,disk3
,,23,2018-05-22T19:53:46Z,0,write_time,diskio,host.local,disk3
,,23,2018-05-22T19:53:56Z,0,write_time,diskio,host.local,disk3
,,23,2018-05-22T19:54:06Z,0,write_time,diskio,host.local,disk3
,,23,2018-05-22T19:54:16Z,0,write_time,diskio,host.local,disk3
,,24,2018-05-22T19:53:26Z,12385052,writes,diskio,host.local,disk0
,,24,2018-05-22T19:53:36Z,12385245,writes,diskio,host.local,disk0
,,24,2018-05-22T19:53:46Z,12385303,writes,diskio,host.local,disk0
,,24,2018-05-22T19:53:56Z,12385330,writes,diskio,host.local,disk0
,,24,2018-05-22T19:54:06Z,12385531,writes,diskio,host.local,disk0
,,24,2018-05-22T19:54:16Z,12385646,writes,diskio,host.local,disk0
,,25,2018-05-22T19:53:26Z,0,writes,diskio,host.local,disk2
,,25,2018-05-22T19:53:36Z,0,writes,diskio,host.local,disk2
,,25,2018-05-22T19:53:46Z,0,writes,diskio,host.local,disk2
,,25,2018-05-22T19:53:56Z,0,writes,diskio,host.local,disk2
,,25,2018-05-22T19:54:06Z,0,writes,diskio,host.local,disk2
,,25,2018-05-22T19:54:16Z,0,writes,diskio,host.local,disk2
,,26,2018-05-22T19:53:26Z,0,writes,diskio,host.local,disk3
,,26,2018-05-22T19:53:36Z,0,writes,diskio,host.local,disk3
,,26,2018-05-22T19:53:46Z,0,writes,diskio,host.local,disk3
,,26,2018-05-22T19:53:56Z,0,writes,diskio,host.local,disk3
,,26,2018-05-22T19:54:06Z,0,writes,diskio,host.local,disk3
,,26,2018-05-22T19:54:16Z,0,writes,diskio,host.local,disk3
"
outData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,false,false,true,false
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,name
,,0,2018-05-22T19:53:26Z,648,io_time,diskio,host.local,
,,0,2018-05-22T19:53:26Z,3929,io_time,diskio,host.local,
,,0,2018-05-22T19:53:26Z,15204688,io_time,diskio,host.local,
,,0,2018-05-22T19:53:36Z,648,io_time,diskio,host.local,
,,0,2018-05-22T19:53:36Z,3929,io_time,diskio,host.local,
,,0,2018-05-22T19:53:36Z,15204894,io_time,diskio,host.local,
,,0,2018-05-22T19:53:46Z,648,io_time,diskio,host.local,
,,0,2018-05-22T19:53:46Z,3929,io_time,diskio,host.local,
,,0,2018-05-22T19:53:46Z,15205102,io_time,diskio,host.local,
,,0,2018-05-22T19:53:56Z,24209920,read_bytes,diskio,host.local,disk3
,,0,2018-05-22T19:53:56Z,202997248,read_bytes,diskio,host.local,disk2
,,0,2018-05-22T19:53:56Z,228585836544,read_bytes,diskio,host.local,disk0
,,0,2018-05-22T19:54:06Z,24209920,read_bytes,diskio,host.local,disk3
,,0,2018-05-22T19:54:06Z,202997248,read_bytes,diskio,host.local,disk2
,,0,2018-05-22T19:54:06Z,228594925568,read_bytes,diskio,host.local,disk0
,,0,2018-05-22T19:54:16Z,24209920,read_bytes,diskio,host.local,disk3
,,0,2018-05-22T19:54:16Z,202997248,read_bytes,diskio,host.local,disk2
,,0,2018-05-22T19:54:16Z,228613324800,read_bytes,diskio,host.local,disk0
"
left = testing.loadStorage(csv: inData)
	|> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:53:50Z)
	|> filter(fn: (r) =>
		(r._measurement == "diskio" and r._field == "io_time"))
	|> group(columns: ["host"])
	|> drop(columns: ["_start", "_stop", "name"])
right = testing.loadStorage(csv: inData)
	|> range(start: 2018-05-22T19:53:50Z, stop: 2018-05-22T19:54:20Z)
	|> filter(fn: (r) =>
		(r._measurement == "diskio" and r._field == "read_bytes"))
	|> group(columns: ["host"])
	|> drop(columns: ["_start", "_stop"])
got = union(tables: [left, right])
	|> sort(columns: ["_time", "_field", "_value"])
want = testing.loadStorage(csv: outData)

testing.assertEquals(name: "union_heterogeneous", want: want, got: got)