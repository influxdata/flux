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
#group,false,false,true,true,false,false,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string,string
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:01:30Z,10,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:01:40Z,11,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:01:50Z,12,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:02:00Z,13,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:02:10Z,14,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:02:20Z,15,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:02:30Z,14.431999999999995,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:02:40Z,13.345599999999994,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:02:50Z,12.155520000000006,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:00Z,11,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:10Z,9.906687999999997,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:20Z,8.865630719999995,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:30Z,7.8589122560000035,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:40Z,6.871005491200005,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:50Z,5.891160883200005,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:00Z,4.912928706560004,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:10Z,3.9329551040511994,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:20Z,2.9498469349785585,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:30Z,1.96332557120307,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:40Z,0.9736696408637426,used_percent,disk,disk1s1,apfs,host.local,/
"

triple_exponential_moving_average = (table=<-) =>
    (table
        |> range(start:2018-05-22T00:00:00Z)
        |> tripleEMA(n:4))

test _triple_exponential_moving_average = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: triple_exponential_moving_average})
