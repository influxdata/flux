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
,,0,2018-05-22T00:01:40Z,100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:50Z,100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:00Z,100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:10Z,100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:20Z,100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:30Z,80,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:40Z,60,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:50Z,40,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:00Z,20,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:10Z,0,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:20Z,-20,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:30Z,-40,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:40Z,-60,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:50Z,-80,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:00Z,-100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:10Z,-100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:20Z,-100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:30Z,-100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:40Z,-100,used_percent,disk,disk1s1,apfs,host.local,/
"

cmo = (table=<-) =>
    (table
        |> range(start:2018-05-22T00:00:00Z)
        |> drop(columns: ["_start", "_stop"])
        |> chandeMomentumOscillator(n:10))

test _cmo = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: cmo})
