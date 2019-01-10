import "testing"

option now = () => 2030-01-01T00:00:00Z

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,false,false,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,name
,,0,2018-05-22T19:53:26Z,10,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:36Z,11,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:46Z,12,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:56Z,13,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:54:06Z,14,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:54:16Z,15,io_time,diskio,host.local,disk0
,,1,2018-05-22T19:53:26Z,16,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:36Z,17,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:46Z,18,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:56Z,19,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:06Z,20,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:16Z,21,io_time,diskio,host.local,disk2
"
outData = "
#datatype,string,long,string,string,string,long
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,host,name,_key,_value
,,0,host.local,disk0,_value,10
,,0,host.local,disk0,_value,11
,,0,host.local,disk0,_value,12
,,0,host.local,disk0,_value,13
,,0,host.local,disk0,_value,14
,,0,host.local,disk0,_value,15
,,1,host.local,disk2,_value,16
,,1,host.local,disk2,_value,17
,,1,host.local,disk2,_value,18
,,1,host.local,disk2,_value,19
,,1,host.local,disk2,_value,20
,,1,host.local,disk2,_value,21
"

t_key_values = (table=<-) =>
  table
  |> keyValues(keyColumns: ["_value"])

testing.test(name: "key_values",
            input: testing.loadStorage(csv: inData),
            want: testing.loadMem(csv: outData),
            testFn: t_key_values,
)
