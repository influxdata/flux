package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,long
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:53:30Z,_m,FF,10
,,0,2018-05-22T19:53:40Z,_m,FF,16
,,0,2018-05-22T19:53:50Z,_m,FF,93
,,0,2018-05-22T19:53:00Z,_m,FF,56
,,0,2018-05-22T19:53:10Z,_m,FF,11
,,0,2018-05-22T19:53:20Z,_m,FF,29
,,1,2018-05-22T19:53:30Z,_m,QQ,26
,,1,2018-05-22T19:53:40Z,_m,QQ,88
,,1,2018-05-22T19:53:50Z,_m,QQ,47
,,1,2018-05-22T19:54:00Z,_m,QQ,78
,,1,2018-05-22T19:54:10Z,_m,QQ,51
,,1,2018-05-22T19:54:30Z,_m,QQ,22
,,1,2018-05-22T19:54:40Z,_m,QQ,19
,,1,2018-05-22T19:53:50Z,_m,QQ,69
,,1,2018-05-22T19:54:00Z,_m,QQ,63
,,2,2018-05-22T19:53:10Z,_m,RR,62
,,2,2018-05-22T19:53:30Z,_m,RR,18
,,2,2018-05-22T19:54:40Z,_m,RR,19
,,2,2018-05-22T19:53:50Z,_m,RR,90
,,3,2018-05-22T19:54:00Z,_m,SR,36
,,3,2018-05-22T19:54:10Z,_m,SR,72
,,3,2018-05-22T19:54:20Z,_m,SR,88
"

outData = "
#datatype,string,long,dateTime:RFC3339,string,string,long
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:48:30Z,_m,FF,10
,,0,2018-05-22T19:48:40Z,_m,FF,16
,,0,2018-05-22T19:48:50Z,_m,FF,93
,,0,2018-05-22T19:48:00Z,_m,FF,56
,,0,2018-05-22T19:48:10Z,_m,FF,11
,,0,2018-05-22T19:48:20Z,_m,FF,29
,,1,2018-05-22T19:48:30Z,_m,QQ,26
,,1,2018-05-22T19:48:40Z,_m,QQ,88
,,1,2018-05-22T19:48:50Z,_m,QQ,47
,,1,2018-05-22T19:49:00Z,_m,QQ,78
,,1,2018-05-22T19:49:10Z,_m,QQ,51
,,1,2018-05-22T19:49:30Z,_m,QQ,22
,,1,2018-05-22T19:49:40Z,_m,QQ,19
,,1,2018-05-22T19:48:50Z,_m,QQ,69
,,1,2018-05-22T19:49:00Z,_m,QQ,63
,,2,2018-05-22T19:48:10Z,_m,RR,62
,,2,2018-05-22T19:48:30Z,_m,RR,18
,,2,2018-05-22T19:49:40Z,_m,RR,19
,,2,2018-05-22T19:48:50Z,_m,RR,90
,,3,2018-05-22T19:49:00Z,_m,SR,36
,,3,2018-05-22T19:49:10Z,_m,SR,72
,,3,2018-05-22T19:49:20Z,_m,SR,88
"

t_shift_negative_duration = (table=<-) =>
	(table
		|> timeShift(duration: -5m))
		|> drop(columns: ["_start", "_stop"])

test _shift_negative_duration = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_shift_negative_duration})

