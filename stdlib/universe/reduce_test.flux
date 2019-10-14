package universe_test
import "testing"

option now = () =>
	(2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,3,2018-05-22T19:53:26Z,1,used_percent,swap,host.local
,,3,2018-05-22T19:53:36Z,1,used_percent,swap,host.local
,,3,2018-05-22T19:53:46Z,1,used_percent,swap,host.local
,,3,2018-05-22T19:53:56Z,1,used_percent,swap,host.local
,,3,2018-05-22T19:54:06Z,1,used_percent,swap,host.local
,,3,2018-05-22T19:54:16Z,1,used_percent,swap,host.local
,,4,2018-05-22T19:53:26Z,1,used_percent,swap,host.local2
,,4,2018-05-22T19:53:36Z,1,used_percent,swap,host.local2
,,4,2018-05-22T19:53:46Z,1,used_percent,swap,host.local2
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double
#group,false,false,true,true,true,true,true,false
#default,got,,,,,,,
,result,table,_start,_stop,_field,_measurement,host,sum
,,0,2018-05-21T13:09:22.885021542Z,2030-01-01T00:00:00Z,used_percent,swap,host.local,6
,,1,2018-05-21T13:09:22.885021542Z,2030-01-01T00:00:00Z,used_percent,swap,host.local2,3
"
t_reduce = (table=<-) =>
	(table
		|> range(start: 2018-05-21T13:09:22.885021542Z)
		|> reduce(fn: (r, accumulator) =>
			({sum: r._value + accumulator.sum}), identity: {sum: 0.0}))

test _reduce = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_reduce})
