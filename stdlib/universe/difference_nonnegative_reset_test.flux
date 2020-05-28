package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,interface
,,0,2018-05-22T19:53:26Z,100,bytes_out,net,eth0
,,0,2018-05-22T19:53:36Z,120,bytes_out,net,eth0
,,0,2018-05-22T19:53:46Z,10,bytes_out,net,eth0
,,0,2018-05-22T19:53:56Z,15,bytes_out,net,eth0
,,0,2018-05-22T19:54:06Z,20,bytes_out,net,eth0
,,0,2018-05-22T19:54:16Z,40,bytes_out,net,eth0
,,1,2018-05-22T19:53:26Z,1000,bytes_out,net,eth1
,,1,2018-05-22T19:53:36Z,1005,bytes_out,net,eth1
,,1,2018-05-22T19:53:46Z,0,bytes_out,net,eth1
,,1,2018-05-22T19:53:56Z,1,bytes_out,net,eth1
,,1,2018-05-22T19:54:06Z,2,bytes_out,net,eth1
,,1,2018-05-22T19:54:16Z,10,bytes_out,net,eth1
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#group,false,false,true,true,false,false,true,true,true
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,interface
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,20,bytes_out,net,eth0
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,10,bytes_out,net,eth0
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,5,bytes_out,net,eth0
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,5,bytes_out,net,eth0
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,20,bytes_out,net,eth0
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,5,bytes_out,net,eth1
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,0,bytes_out,net,eth1
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,1,bytes_out,net,eth1
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,1,bytes_out,net,eth1
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,8,bytes_out,net,eth1
"

t_difference = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> difference(nonNegative: true))

test _difference_nonnegative = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_difference})

