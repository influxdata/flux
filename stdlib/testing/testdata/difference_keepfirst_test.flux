package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,34.98234271799806,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T19:53:36Z,34.98234941084654,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T19:53:46Z,34.982447293755506,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T19:53:56Z,34.982447293755506,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T19:54:06Z,34.98204153981662,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T19:54:16Z,34.982252364543626,used_percent,disk,disk1s1,apfs,host.local,/
,,1,2018-05-22T19:53:26Z,34.98234271799806,used_percent,disk,disk1s2,apfs,host.local,/
,,1,2018-05-22T19:53:36Z,34.98234941084654,used_percent,disk,disk1s2,apfs,host.local,/
,,1,2018-05-22T19:53:46Z,34.982447293755506,used_percent,disk,disk1s2,apfs,host.local,/
,,1,2018-05-22T19:53:56Z,34.982447293755506,used_percent,disk,disk1s2,apfs,host.local,/
,,1,2018-05-22T19:54:06Z,34.98204153981662,used_percent,disk,disk1s2,apfs,host.local,/
,,1,2018-05-22T19:54:16Z,34.982252364543626,used_percent,disk,disk1s2,apfs,host.local,/
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,0.000006692848479872282,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,0.00009788290896750595,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,0,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,-0.0004057539388853115,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,0.00021082472700584276,used_percent,disk,disk1s1,apfs,host.local,/
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,,used_percent,disk,disk1s2,apfs,host.local,/
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,0.000006692848479872282,used_percent,disk,disk1s2,apfs,host.local,/
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,0.00009788290896750595,used_percent,disk,disk1s2,apfs,host.local,/
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,0,used_percent,disk,disk1s2,apfs,host.local,/
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,-0.0004057539388853115,used_percent,disk,disk1s2,apfs,host.local,/
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,0.00021082472700584276,used_percent,disk,disk1s2,apfs,host.local,/
"

t_difference = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> difference(keepFirst: true))

test _difference = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_difference})

