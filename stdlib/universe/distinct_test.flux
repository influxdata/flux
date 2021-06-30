package universe_test


import "csv"
import "testing"

t_distinct = (table=<-) => table
    |> range(start: 2018-05-20T19:53:26Z, stop: 2030-01-01T00:00:00Z)
    |> distinct(column: "_value")
    |> drop(columns: ["_start", "_stop"])

testcase normal {
    got = csv.from(
        csv: "
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
",
    )
        |> t_distinct()

    want = csv.from(
        csv: "
#datatype,string,long,string,string,string,string,long
#group,false,false,true,true,true,true,false
#default,0,,,,,,
,result,table,_field,_measurement,host,name,_value
,,0,io_time,diskio,host.local,disk0,15204688
,,0,io_time,diskio,host.local,disk0,15204894
,,0,io_time,diskio,host.local,disk0,15205102
,,0,io_time,diskio,host.local,disk0,15205226
,,0,io_time,diskio,host.local,disk0,15205499
,,0,io_time,diskio,host.local,disk0,15205755
,,1,io_time,diskio,host.local,disk2,648
",
    )

    testing.diff(got, want) |> yield()
}

testcase nulls {
    got = csv.from(
        csv: "
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,name
,,0,2018-05-22T19:53:26Z,15204688,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:36Z,15204894,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:46Z,,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:56Z,15205226,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:54:06Z,15205499,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:54:16Z,15205755,io_time,diskio,host.local,disk0
,,1,2018-05-22T19:53:26Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:36Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:46Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:56Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:06Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:16Z,,io_time,diskio,host.local,disk2
",
    )
        |> t_distinct()

    want = csv.from(
        csv: "
#datatype,string,long,string,string,string,string,long
#group,false,false,true,true,true,true,false
#default,0,,,,,,
,result,table,_field,_measurement,host,name,_value
,,0,io_time,diskio,host.local,disk0,15204688
,,0,io_time,diskio,host.local,disk0,15204894
,,0,io_time,diskio,host.local,disk0,
,,0,io_time,diskio,host.local,disk0,15205226
,,0,io_time,diskio,host.local,disk0,15205499
,,0,io_time,diskio,host.local,disk0,15205755
,,1,io_time,diskio,host.local,disk2,648
,,1,io_time,diskio,host.local,disk2,
",
    )

    testing.diff(got, want) |> yield()
}
