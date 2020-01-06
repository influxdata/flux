package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,false
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,36.946678161621094,available_percent,mem,host1
,,0,2018-05-22T19:53:36Z,37.28463649749756,available_percent,mem,host2
,,0,2018-05-22T19:53:46Z,37.61239051818848,available_percent,mem,host3
,,0,2018-05-22T19:53:56Z,37.25404739379883,available_percent,mem,host4
,,0,2018-05-22T19:54:06Z,37.21816539764404,available_percent,mem,host5
,,0,2018-05-22T19:54:16Z,37.53254413604736,available_percent,mem,host5
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,false,false,false,false,true
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,37.28463649749756,available_percent,mem,host2
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,37.61239051818848,available_percent,mem,host3
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,37.53254413604736,available_percent,mem,host5
"

t_selector_preserve_time = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> top(n: 3)
		|> group(columns: ["host"]))

test _selector_preserve_time = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_selector_preserve_time})

