package universe_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
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
outData =
    "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,dateTime:RFC3339,long
#group,false,false,true,false,true,false,false,true,false,false
#default,_result,,,,,,,,,
,result,table,_start,_stop,_measurement,_field,host,name,_time,_value
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,diskio,io_time,host.local,disk0,2018-05-22T19:54:16Z,15205755
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,diskio,io_time,host.local,disk2,2018-05-22T19:53:26Z,648
"

// Passes in flux, fails in C2 and OSS
testcase group {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> filter(fn: (r) => r._measurement == "diskio" and r._field == "io_time")
            |> group(columns: ["_measurement", "_start", "name"])
            |> max()
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
