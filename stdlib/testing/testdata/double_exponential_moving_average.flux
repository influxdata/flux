package testdata_test

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
#group,false,false,true,true,false,false,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string,string
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:00Z,13.568840926166239,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:10Z,12.70174811931398,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:20Z,11.701405062848782,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:30Z,10.611872766773772,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:40Z,9.465595022565747,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:50Z,8.286166283961508,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:00Z,7.0904770859219255,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:10Z,5.890371851336026,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:20Z,4.6939254760732005,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:30Z,3.5064225149113675,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:40Z,2.3311049123183603,used_percent,disk,disk1s1,apfs,host.local,/
"

double_exponential_moving_average = (table=<-) =>
    (table
        |> range(start:2018-05-22T00:00:00Z)
        |> doubleEMA(n:10))

test _double_exponential_moving_average = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: double_exponential_moving_average})