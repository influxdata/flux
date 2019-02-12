package testdata_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,1.83,load1,system,host.local
,,0,2018-05-22T19:53:36Z,1.7,load1,system,host.local
,,0,2018-05-22T19:53:46Z,1.74,load1,system,host.local
,,0,2018-05-22T19:53:56Z,1.63,load1,system,host.local
,,0,2018-05-22T19:54:06Z,1.91,load1,system,host.local
,,0,2018-05-22T19:54:16Z,1.84,load1,system,host.local
,,0,2018-05-22T19:54:18Z,1.85,load1,system,host.local
,,0,2018-05-22T19:54:21Z,1.86,load1,system,host.local
,,0,2018-05-22T19:54:22Z,1.87,load1,system,host.local
,,1,2018-05-22T19:53:26Z,1.98,load15,system,host.local
,,1,2018-05-22T19:53:36Z,1.97,load15,system,host.local
,,1,2018-05-22T19:53:46Z,1.97,load15,system,host.local
,,1,2018-05-22T19:53:56Z,1.96,load15,system,host.local
,,1,2018-05-22T19:54:06Z,1.98,load15,system,host.local
,,1,2018-05-22T19:54:16Z,1.97,load15,system,host.local
,,2,2018-05-22T19:53:26Z,1.95,load5,system,host.local
,,2,2018-05-22T19:53:36Z,1.92,load5,system,host.local
,,2,2018-05-22T19:53:46Z,1.92,load5,system,host.local
,,2,2018-05-22T19:53:56Z,1.89,load5,system,host.local
,,2,2018-05-22T19:54:06Z,1.94,load5,system,host.local
,,2,2018-05-22T19:54:16Z,1.93,load5,system,host.local
,,3,2018-05-22T19:53:26Z,82.9833984375,used_percent,swap,host.local
,,3,2018-05-22T19:53:36Z,82.598876953125,used_percent,swap,host.local
,,3,2018-05-22T19:53:46Z,82.598876953125,used_percent,swap,host.local
,,3,2018-05-22T19:53:56Z,82.598876953125,used_percent,swap,host.local
,,3,2018-05-22T19:54:06Z,82.598876953125,used_percent,swap,host.local
,,3,2018-05-22T19:54:16Z,82.6416015625,used_percent,swap,host.local
"

outData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:36Z,1.7,load1,system,host.local
,,0,2018-05-22T19:54:06Z,1.91,load1,system,host.local
,,0,2018-05-22T19:54:21Z,1.86,load1,system,host.local
,,1,2018-05-22T19:53:36Z,1.97,load15,system,host.local
,,1,2018-05-22T19:54:06Z,1.98,load15,system,host.local
,,2,2018-05-22T19:53:36Z,1.92,load5,system,host.local
,,2,2018-05-22T19:54:06Z,1.94,load5,system,host.local
,,3,2018-05-22T19:53:36Z,82.598876953125,used_percent,swap,host.local
,,3,2018-05-22T19:54:06Z,82.598876953125,used_percent,swap,host.local
"

t_sample = (table=<-) =>
	(table
		|> sample(n: 3, pos: 1))

test _sample = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_sample})

