package universe_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:00Z,1,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:10Z,2,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:20Z,3,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:30Z,4,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:40Z,5,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:50Z,6,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:00Z,7,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:10Z,8,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:20Z,9,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:30Z,10,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:40Z,11,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:50Z,12,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:00Z,13,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:10Z,14,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:20Z,15,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:30Z,14,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:40Z,13,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:50Z,12,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:00Z,11,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:10Z,10,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:20Z,9,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:30Z,8,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:40Z,7,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:50Z,6,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:00Z,5,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:10Z,4,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:20Z,3,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:30Z,2,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:40Z,1,used_percent,disk,disk1s1,apfs,host.local,/
"

outData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:01:40Z,10.444444444444445,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:50Z,11.135802469135802,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:00Z,11.964334705075446,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:10Z,12.869074836153025,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:20Z,13.81615268675168,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:30Z,13.871008014588556,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:40Z,13.71308456353558,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:50Z,13.553331356741122,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:00Z,13.46599437575161,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:10Z,13.4515677602438,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:20Z,13.29930139347417,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:30Z,12.805116570729282,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:40Z,11.752584300922965,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:50Z,10.036160535131101,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:00Z,7.797866963961722,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:10Z,6.109926091089845,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:20Z,4.727736717272135,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:30Z,3.515409287373408,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:40Z,2.3974496040963373,used_percent,disk,disk1s1,apfs,host.local,/
"

kama = (table=<-) =>
    (table
        |> range(start:2018-05-22T00:00:00Z)
        |> drop(columns: ["_start", "_stop"])
        |> kaufmansAMA(n: 10)
    )

test _kama = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: kama})
