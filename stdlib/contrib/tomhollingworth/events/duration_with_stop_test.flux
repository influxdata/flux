package events_test

import "testing"
import "contrib/tomhollingworth/events"

option now = () => (2018-05-22T19:54:16Z)

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
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string,string,long
#group,false,false,true,true,false,false,true,true,true,true,true,true,false
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path,duration
,,0,2018-05-22T19:53:26Z,2018-05-22T19:54:36Z,2018-05-22T19:53:26Z,34.98234271799806,used_percent,disk,disk1s1,apfs,host.local,/,10
,,0,2018-05-22T19:53:26Z,2018-05-22T19:54:36Z,2018-05-22T19:53:36Z,34.98234941084654,used_percent,disk,disk1s1,apfs,host.local,/,10
,,0,2018-05-22T19:53:26Z,2018-05-22T19:54:36Z,2018-05-22T19:53:46Z,34.982447293755506,used_percent,disk,disk1s1,apfs,host.local,/,10
,,0,2018-05-22T19:53:26Z,2018-05-22T19:54:36Z,2018-05-22T19:53:56Z,34.982447293755506,used_percent,disk,disk1s1,apfs,host.local,/,10
,,0,2018-05-22T19:53:26Z,2018-05-22T19:54:36Z,2018-05-22T19:54:06Z,34.98204153981662,used_percent,disk,disk1s1,apfs,host.local,/,10
,,0,2018-05-22T19:53:26Z,2018-05-22T19:54:36Z,2018-05-22T19:54:16Z,34.982252364543626,used_percent,disk,disk1s1,apfs,host.local,/,30
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:36Z,2018-05-22T19:53:26Z,34.98234271799806,used_percent,disk,disk1s2,apfs,host.local,/,10
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:36Z,2018-05-22T19:53:36Z,34.98234941084654,used_percent,disk,disk1s2,apfs,host.local,/,10
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:36Z,2018-05-22T19:53:46Z,34.982447293755506,used_percent,disk,disk1s2,apfs,host.local,/,10
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:36Z,2018-05-22T19:53:56Z,34.982447293755506,used_percent,disk,disk1s2,apfs,host.local,/,10
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:36Z,2018-05-22T19:54:06Z,34.98204153981662,used_percent,disk,disk1s2,apfs,host.local,/,10
,,1,2018-05-22T19:53:26Z,2018-05-22T19:54:36Z,2018-05-22T19:54:16Z,34.982252364543626,used_percent,disk,disk1s2,apfs,host.local,/,30
"

t_duration = (table=<-) =>
	(table
        |> range(start:2018-05-22T19:53:26Z, stop:2018-05-22T19:54:36Z)
		|> events.duration(stop: 2018-05-22T19:54:46Z))

test _duration = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_duration})
