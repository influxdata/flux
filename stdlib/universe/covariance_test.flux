package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,double,string
#group,false,false,false,false,false,true
#default,_result,,,,,
,result,table,_time,x,y,_measurement
,,0,2018-05-22T19:53:26Z,1,4,cpu
,,0,2018-05-22T19:53:36Z,2,3,cpu
,,0,2018-05-22T19:53:46Z,3,2,cpu
,,0,2018-05-22T19:53:56Z,4,1,cpu
,,1,2018-05-22T19:53:26Z,10,40,mem
,,1,2018-05-22T19:53:36Z,20,30,mem
,,1,2018-05-22T19:53:46Z,30,20,mem
,,1,2018-05-22T19:53:56Z,40,10,mem
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,double
#group,false,false,true,true,true,false
#default,_result,,,,,
,result,table,_start,_stop,_measurement,_value
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,cpu,-1.6666666666666667
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,mem,-166.66666666666666
"

t_covariance = (tables=<-) =>
	(tables
		|> range(start: 2018-05-22T19:53:26Z)
		|> covariance(columns: ["x", "y"]))

test _t_covariance = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_covariance})

