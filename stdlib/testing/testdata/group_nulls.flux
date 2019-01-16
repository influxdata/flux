import "testing"

// The purpose of this test is to test the behavior of Group in case of nulls:
//  - tables have null values here and there (also in group key columns),
//  - the second group of tables misses column "name": its values should be filled with nulls.

inData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:26Z,,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:36Z,15204894,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:46Z,15205102,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:56Z,,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:06Z,15205499,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,,15205755,io_time,diskio,host1,disk0
,,2,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:26Z,15204688,io_time,diskio,host1,disk1
,,2,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,,,io_time,diskio,host1,disk1
,,2,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:46Z,,io_time,diskio,host1,disk1
,,2,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:56Z,15205226,io_time,diskio,host1,disk1
,,2,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:06Z,15205499,io_time,diskio,host1,disk1
,,2,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:16Z,15205755,io_time,diskio,host1,disk1
,,3,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:16Z,15205755,io_time,diskio,,disk1

#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,false,false,true,true,true
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,4,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:26Z,15204688,io_time,diskio,host1
,,4,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,,15204894,io_time,diskio,host1
,,4,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:46Z,15205102,io_time,diskio,host1
,,4,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:56Z,15205226,io_time,diskio,host1
,,4,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:06Z,15205499,io_time,diskio,host1
,,4,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:16Z,,io_time,diskio,host1
,,5,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,,15204688,io_time,diskio,host2
,,5,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:36Z,,io_time,diskio,host2
,,5,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:46Z,15205102,io_time,diskio,host2
,,5,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:56Z,15205226,io_time,diskio,host2
,,5,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:06Z,15205499,io_time,diskio,host2
,,5,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:16Z,,io_time,diskio,host2
,,6,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:46Z,15205102,io_time,diskio,
,,6,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,,15205226,io_time,diskio,
,,6,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:06Z,15205499,io_time,diskio,
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,false,false,false,false,true,false
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:26Z,15204688,io_time,diskio,host1,
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,,15204894,io_time,diskio,host1,
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:46Z,15205102,io_time,diskio,host1,
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:56Z,15205226,io_time,diskio,host1,
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:06Z,15205499,io_time,diskio,host1,
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:16Z,,io_time,diskio,host1,
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:26Z,,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:36Z,15204894,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:46Z,15205102,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:56Z,,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:06Z,15205499,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,,15205755,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:26Z,15204688,io_time,diskio,host1,disk1
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,,,io_time,diskio,host1,disk1
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:46Z,,io_time,diskio,host1,disk1
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:56Z,15205226,io_time,diskio,host1,disk1
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:06Z,15205499,io_time,diskio,host1,disk1
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:16Z,15205755,io_time,diskio,host1,disk1
,,3,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:16Z,15205755,io_time,diskio,,disk1
,,3,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:46Z,15205102,io_time,diskio,,
,,3,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,,15205226,io_time,diskio,,
,,3,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:06Z,15205499,io_time,diskio,,

#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,false,false,false,false,true
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,2,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,,15204688,io_time,diskio,host2
,,2,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:36Z,,io_time,diskio,host2
,,2,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:46Z,15205102,io_time,diskio,host2
,,2,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:56Z,15205226,io_time,diskio,host2
,,2,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:06Z,15205499,io_time,diskio,host2
,,2,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:16Z,,io_time,diskio,host2
"

t_group = (table=<-) =>
  table
  |> range(start: 1902-05-22T19:53:26Z)
  |> group(columns: ["host"])

testing.test(
    name: "group",
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    testFn: t_group)
