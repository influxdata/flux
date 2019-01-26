import "testing"

option now = () => 2030-01-01T00:00:00Z

// The purpose of this test is to test the behavior of Group in case of nulls:
//  - tables have null values here and there (also in group key columns),
//  - the second group of tables misses column "name": its values should be filled with nulls.

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,name
,,0,2018-05-22T19:54:16Z,15205755,io_time,diskio,,disk1

#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,1,2018-05-22T19:53:46Z,15205102,io_time,diskio,
,,1,,15205226,io_time,diskio,
,,1,2018-05-22T19:54:06Z,15205499,io_time,diskio,
,,2,2018-05-22T19:53:26Z,15204688,io_time,diskio,host1
,,2,,15204894,io_time,diskio,host1
,,2,2018-05-22T19:53:46Z,15205102,io_time,diskio,host1
,,2,2018-05-22T19:53:56Z,15205226,io_time,diskio,host1
,,2,2018-05-22T19:54:06Z,15205499,io_time,diskio,host1
,,2,2018-05-22T19:54:16Z,,io_time,diskio,host1

#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,name
,,3,2018-05-22T19:53:26Z,,io_time,diskio,host1,disk0
,,3,2018-05-22T19:53:36Z,15204894,io_time,diskio,host1,disk0
,,3,2018-05-22T19:53:46Z,15205102,io_time,diskio,host1,disk0
,,3,2018-05-22T19:53:56Z,,io_time,diskio,host1,disk0
,,3,2018-05-22T19:54:06Z,15205499,io_time,diskio,host1,disk0
,,3,,15205755,io_time,diskio,host1,disk0
,,4,2018-05-22T19:53:26Z,15204688,io_time,diskio,host1,disk1
,,4,,,io_time,diskio,host1,disk1
,,4,2018-05-22T19:53:46Z,,io_time,diskio,host1,disk1
,,4,2018-05-22T19:53:56Z,15205226,io_time,diskio,host1,disk1
,,4,2018-05-22T19:54:06Z,15205499,io_time,diskio,host1,disk1
,,4,2018-05-22T19:54:16Z,15205755,io_time,diskio,host1,disk1

#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,5,,15204688,io_time,diskio,host2
,,5,2018-05-22T19:53:36Z,,io_time,diskio,host2
,,5,2018-05-22T19:53:46Z,15205102,io_time,diskio,host2
,,5,2018-05-22T19:53:56Z,15205226,io_time,diskio,host2
,,5,2018-05-22T19:54:06Z,15205499,io_time,diskio,host2
,,5,2018-05-22T19:54:16Z,,io_time,diskio,host2
"

outData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,false,false,true,false
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,name
,,0,2018-05-22T19:54:16Z,15205755,io_time,diskio,,disk1
,,0,2018-05-22T19:53:46Z,15205102,io_time,diskio,,
,,0,,15205226,io_time,diskio,,
,,0,2018-05-22T19:54:06Z,15205499,io_time,diskio,,
,,1,2018-05-22T19:53:26Z,15204688,io_time,diskio,host1,
,,1,,15204894,io_time,diskio,host1,
,,1,2018-05-22T19:53:46Z,15205102,io_time,diskio,host1,
,,1,2018-05-22T19:53:56Z,15205226,io_time,diskio,host1,
,,1,2018-05-22T19:54:06Z,15205499,io_time,diskio,host1,
,,1,2018-05-22T19:54:16Z,,io_time,diskio,host1,
,,1,2018-05-22T19:53:26Z,,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:36Z,15204894,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:46Z,15205102,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:56Z,,io_time,diskio,host1,disk0
,,1,2018-05-22T19:54:06Z,15205499,io_time,diskio,host1,disk0
,,1,,15205755,io_time,diskio,host1,disk0
,,1,2018-05-22T19:53:26Z,15204688,io_time,diskio,host1,disk1
,,1,,,io_time,diskio,host1,disk1
,,1,2018-05-22T19:53:46Z,,io_time,diskio,host1,disk1
,,1,2018-05-22T19:53:56Z,15205226,io_time,diskio,host1,disk1
,,1,2018-05-22T19:54:06Z,15205499,io_time,diskio,host1,disk1
,,1,2018-05-22T19:54:16Z,15205755,io_time,diskio,host1,disk1
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,false,false,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,2,,15204688,io_time,diskio,host2
,,2,2018-05-22T19:53:36Z,,io_time,diskio,host2
,,2,2018-05-22T19:53:46Z,15205102,io_time,diskio,host2
,,2,2018-05-22T19:53:56Z,15205226,io_time,diskio,host2
,,2,2018-05-22T19:54:06Z,15205499,io_time,diskio,host2
,,2,2018-05-22T19:54:16Z,,io_time,diskio,host2
"

t_group = (table=<-) =>
  table
  |> group(columns: ["host"])

testing.run(name: "group",
            input: testing.loadStorage(csv: inData),
            want: testing.loadMem(csv: outData),
            testFn: t_group)
