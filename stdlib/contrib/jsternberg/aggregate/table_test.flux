package aggregate_test

import "testing"
import "contrib/jsternberg/aggregate"

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
,,1,2018-05-22T00:00:00Z,35,used_percent,disk,disk1s1,apfs,host.local,/tmp
,,1,2018-05-22T00:00:10Z,35,used_percent,disk,disk1s1,apfs,host.local,/tmp
,,1,2018-05-22T00:00:20Z,35,used_percent,disk,disk1s1,apfs,host.local,/tmp
,,1,2018-05-22T00:00:30Z,45,used_percent,disk,disk1s1,apfs,host.local,/tmp
,,1,2018-05-22T00:00:40Z,45,used_percent,disk,disk1s1,apfs,host.local,/tmp
,,1,2018-05-22T00:00:50Z,45,used_percent,disk,disk1s1,apfs,host.local,/tmp
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,double,double,double,double,long
#group,false,false,true,true,true,true,true,true,true,true,false,false,false,false,false
#default,_result,,,,,,,,,,,,,,
,result,table,_start,_stop,_field,_measurement,device,fstype,host,path,sum,mean,min,max,count
,,0,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,used_percent,disk,disk1s1,apfs,host.local,/,210,35,30,40,6
,,1,2018-05-22T00:00:00Z,2018-05-22T00:01:00Z,used_percent,disk,disk1s1,apfs,host.local,/tmp,240,40,35,45,6
"

aggregate_table = (table=<-) => table
	|> range(start: 2018-05-22T00:00:00Z, stop: 2018-05-22T00:01:00Z)
	|> aggregate.table(columns: {
		sum: aggregate.sum(),
		mean: aggregate.mean(),
		min: aggregate.min(),
		max: aggregate.max(),
		count: aggregate.count(),
	})

test _aggregate_table = () => ({
	input: testing.loadStorage(csv: inData),
	want: testing.loadMem(csv: outData),
	fn: aggregate_table,
})
