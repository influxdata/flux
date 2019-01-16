import "testing"

inData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,false,false,false,false,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,name
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:26Z,15204688,io_time,diskio,host.local,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:36Z,15204894,io_time,diskio,host.local,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:46Z,15205102,io_time,diskio,host.local,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:56Z,15205226,io_time,diskio,host.local,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:06Z,15205499,io_time,diskio,host.local,disk0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:16Z,15205755,io_time,diskio,host.local,disk0
,,10,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:26Z,648,io_time,diskio,host.local,disk2
,,10,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:36Z,648,io_time,diskio,host.local,disk2
,,10,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:46Z,648,io_time,diskio,host.local,disk2
,,10,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:53:56Z,648,io_time,diskio,host.local,disk2
,,10,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:06Z,648,io_time,diskio,host.local,disk2
,,10,2018-05-22T19:53:26Z,2018-05-22T19:54:16Z,2018-05-22T19:54:16Z,648,io_time,diskio,host.local,disk2
"
outData = "
#datatype,string,long,string,string,string
#group,false,false,true,true,false
#default,,,,,
,result,table,host,name,_value
,,0,host.local,disk0,_start
,,0,host.local,disk0,_stop
,,0,host.local,disk0,_time
,,0,host.local,disk0,_value
,,0,host.local,disk0,_field
,,0,host.local,disk0,_measurement
,,0,host.local,disk0,host
,,0,host.local,disk0,name
,,1,host.local,disk2,_start
,,1,host.local,disk2,_stop
,,1,host.local,disk2,_time
,,1,host.local,disk2,_value
,,1,host.local,disk2,_field
,,1,host.local,disk2,_measurement
,,1,host.local,disk2,host
,,1,host.local,disk2,name
"

t_columns = (table=<-) =>
  table
  |> range(start: 2018-05-20T19:53:26Z)
  |> columns()

testing.test(
    name: "columns",
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    testFn: t_columns)