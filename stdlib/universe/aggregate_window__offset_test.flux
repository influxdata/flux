package universe_test


import "csv"
import "testing"

option now = () => 2030-01-01T00:00:00Z

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:00Z,20,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:10Z,30,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:20Z,30,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:30Z,30,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:40Z,40,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:50Z,40,used_percent,disk,disk1s1,apfs,host.local,/
,,1,2018-05-22T00:00:00Z,25,used_percent,disk,disk1s1,apfs,host.local,/tmp
,,1,2018-05-22T00:00:10Z,35,used_percent,disk,disk1s1,apfs,host.local,/tmp
,,1,2018-05-22T00:00:20Z,35,used_percent,disk,disk1s1,apfs,host.local,/tmp
,,1,2018-05-22T00:00:30Z,35,used_percent,disk,disk1s1,apfs,host.local,/tmp
,,1,2018-05-22T00:00:40Z,45,used_percent,disk,disk1s1,apfs,host.local,/tmp
,,1,2018-05-22T00:00:50Z,45,used_percent,disk,disk1s1,apfs,host.local,/tmp
"

testcase aggregate_window_offset {
    outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,double
#group,false,false,true,true,false,true,true,true,true,true,true,false
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_field,_measurement,device,fstype,host,path,_value
,,0,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,2018-05-22T00:00:05Z,used_percent,disk,disk1s1,apfs,host.local,/,20
,,0,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,2018-05-22T00:00:35Z,used_percent,disk,disk1s1,apfs,host.local,/,30
,,0,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,2018-05-22T00:01:00Z,used_percent,disk,disk1s1,apfs,host.local,/,40
,,1,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,2018-05-22T00:00:05Z,used_percent,disk,disk1s1,apfs,host.local,/tmp,25
,,1,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,2018-05-22T00:00:35Z,used_percent,disk,disk1s1,apfs,host.local,/tmp,35
,,1,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,2018-05-22T00:01:00Z,used_percent,disk,disk1s1,apfs,host.local,/tmp,45
"
    got = testing.loadStorage(csv: inData)
        |> range(start: 2018-05-22T00:00:00Z, stop: 2018-05-22T00:01:00Z)
        |> aggregateWindow(every: 30s, fn: mean, offset: 5s)
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got)
        |> yield(name: "diff")
}
