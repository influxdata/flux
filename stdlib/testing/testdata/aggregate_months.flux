package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:00Z,30,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:10Z,32,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-06-22T00:00:20Z,37,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-06-22T00:00:30Z,49,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-06-22T00:00:40Z,43,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-07-22T00:00:50Z,61,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-08-22T00:01:00Z,63,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-09-22T00:01:10Z,53,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-09-22T00:01:20Z,50,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-10-22T00:01:30Z,49,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-11-22T00:01:40Z,69,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-12-22T00:01:50Z,75,used_percent,disk,disk1s1,apfs,host.local,/
"

outData = "
#group,false,false,true,true,true,true,true,true,true,true,true,false
#datatype,string,long,string,string,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,double
#default,_result,,,,,,,,,,,
,result,table,_field,_measurement,_start,_stop,device,fstype,host,monthYear,path,_value
,,0,used_percent,disk,2018-05-22T00:00:00Z,2019-01-01T00:01:00Z,disk1s1,apfs,host.local,2018-05,/,31
,,1,used_percent,disk,2018-05-22T00:00:00Z,2019-01-01T00:01:00Z,disk1s1,apfs,host.local,2018-06,/,43
,,2,used_percent,disk,2018-05-22T00:00:00Z,2019-01-01T00:01:00Z,disk1s1,apfs,host.local,2018-07,/,61
,,3,used_percent,disk,2018-05-22T00:00:00Z,2019-01-01T00:01:00Z,disk1s1,apfs,host.local,2018-08,/,63
,,4,used_percent,disk,2018-05-22T00:00:00Z,2019-01-01T00:01:00Z,disk1s1,apfs,host.local,2018-09,/,51.5
,,5,used_percent,disk,2018-05-22T00:00:00Z,2019-01-01T00:01:00Z,disk1s1,apfs,host.local,2018-10,/,49
,,6,used_percent,disk,2018-05-22T00:00:00Z,2019-01-01T00:01:00Z,disk1s1,apfs,host.local,2018-11,/,69
,,7,used_percent,disk,2018-05-22T00:00:00Z,2019-01-01T00:01:00Z,disk1s1,apfs,host.local,2018-12,/,75
"

aggregate_months = (table=<-) =>
	(table
    	|> range(start: 2018-05-22T00:00:00Z, stop: 2019-01-01T00:01:00Z)
		|> aggregateMonths(fn: mean, column: "_value"))

test _aggregate_months = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: aggregate_months})
