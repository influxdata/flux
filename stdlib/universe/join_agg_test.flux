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
,,0,2019-05-10T20:50:00Z,11930171,reads,diskio,ip-192-168-1-16.ec2.internal,disk0
,,0,2019-05-10T20:50:10Z,11930171,reads,diskio,ip-192-168-1-16.ec2.internal,disk0
,,1,2019-05-10T20:50:00Z,391,reads,diskio,ip-192-168-1-16.ec2.internal,disk2
,,1,2019-05-10T20:50:10Z,391,reads,diskio,ip-192-168-1-16.ec2.internal,disk2
,,2,2019-05-10T20:50:00Z,34399675,writes,diskio,ip-192-168-1-16.ec2.internal,disk0
,,2,2019-05-10T20:50:10Z,34399831,writes,diskio,ip-192-168-1-16.ec2.internal,disk0
,,3,2019-05-10T20:50:00Z,0,writes,diskio,ip-192-168-1-16.ec2.internal,disk2
,,3,2019-05-10T20:50:10Z,0,writes,diskio,ip-192-168-1-16.ec2.internal,disk2
"
outData =
    "
#datatype,string,long,string,long,long
#group,false,false,true,false,false
#default,_result,,,,
,result,table,name,total_reads,total_writes
,,0,disk0,23860342,68799506
,,1,disk2,782,0
"

testcase join_agg {
    table =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2019-05-10T20:50:00Z, stop: 2019-05-10T20:51:00Z)

    left =
        table
            |> range(start: 2018-05-22T19:53:00Z)
            |> filter(fn: (r) => r._field == "reads")
            |> group(columns: ["name"])
            |> keep(columns: ["name", "_value"])
            |> sum()
            |> rename(columns: {_value: "total_reads"})
    right =
        table
            |> range(start: 2018-05-22T19:53:00Z)
            |> filter(fn: (r) => r._field == "writes")
            |> group(columns: ["name"])
            |> keep(columns: ["name", "_value"])
            |> sum()
            |> rename(columns: {_value: "total_writes"})
    got = join(tables: {left, right}, on: ["name"])
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
