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
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T00:01:40Z,100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T00:01:50Z,100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T00:02:00Z,100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T00:02:10Z,100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T00:02:20Z,100,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T00:02:30Z,90,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T00:02:40Z,81,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T00:02:50Z,72.9,used_percent,disk,disk1s1,apfs,host.local,/
"
relative_strength_index = (table=<-) =>
	(table
        |> range(start:2018-05-22T00:00:00Z)
		|> relativeStrengthIndex(n:10))

test _relative_strength_index = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: relative_strength_index})

