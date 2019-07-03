package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:00Z,30,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:10Z,30,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:20Z,30,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:30Z,40,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:40Z,40,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:50Z,40,used_percent,disk,disk1s1,apfs,host.local,/
"

outData = "
#group,false,false,true,true,true,true,true,true,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,double,dateTime:RFC3339
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_field,_measurement,device,fstype,host,path,_value,_time
,,0,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,used_percent,disk,disk1s1,apfs,host.local,/,30,2018-05-22T00:00:10Z
,,0,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,used_percent,disk,disk1s1,apfs,host.local,/,30,2018-05-22T00:00:20Z
,,0,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,used_percent,disk,disk1s1,apfs,host.local,/,30,2018-05-22T00:00:30Z
,,0,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,used_percent,disk,disk1s1,apfs,host.local,/,33.333333333333336,2018-05-22T00:00:40Z
,,0,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,used_percent,disk,disk1s1,apfs,host.local,/,36.666666666666664,2018-05-22T00:00:50Z
,,0,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,used_percent,disk,disk1s1,apfs,host.local,/,40,2018-05-22T00:01:00Z
,,0,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,used_percent,disk,disk1s1,apfs,host.local,/,40,2018-05-22T00:01:00Z
,,0,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,used_percent,disk,disk1s1,apfs,host.local,/,40,2018-05-22T00:01:00Z
"
moving_average = (table=<-) =>
	(table
		|> range(start: 2018-05-22T00:00:00Z, stop: 2018-05-22T00:01:00Z)
		|> movingAverage(every: 10s, period: 30s))

test _moving_average = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: moving_average})

