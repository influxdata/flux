package main
import "testing"

option now = () =>
	(2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,long,long,long
#group,false,false,false,true,false,false,false
#default,_result,,,,,,
,result,table,_time,_measurement,v0,v1,v2
,,0,2018-05-22T19:53:30Z,_m0,1,10,100
,,0,2018-05-22T19:53:40Z,_m0,1,10,100
,,0,2018-05-22T19:53:50Z,_m0,1,10,100
,,0,2018-05-22T19:53:00Z,_m0,1,10,100
,,0,2018-05-22T19:53:10Z,_m0,1,10,100
,,0,2018-05-22T19:53:20Z,_m0,1,10,100
,,1,2018-05-22T19:53:30Z,_m1,1,10,100
,,1,2018-05-22T19:53:40Z,_m1,1,10,100
,,1,2018-05-22T19:53:50Z,_m1,1,10,100
,,1,2018-05-22T19:54:00Z,_m1,1,10,100
,,1,2018-05-22T19:54:10Z,_m1,1,10,100
,,1,2018-05-22T19:54:30Z,_m1,1,10,100
,,1,2018-05-22T19:54:40Z,_m1,1,10,100
,,1,2018-05-22T19:53:50Z,_m1,1,10,100
,,1,2018-05-22T19:54:00Z,_m1,1,10,100
,,2,2018-05-22T19:53:10Z,_m2,1,10,100
,,2,2018-05-22T19:53:30Z,_m2,1,10,100
,,2,2018-05-22T19:54:40Z,_m2,1,10,100
,,2,2018-05-22T19:53:50Z,_m2,1,10,100
,,3,2018-05-22T19:54:00Z,_m3,1,10,100
,,3,2018-05-22T19:54:10Z,_m3,1,10,100
,,3,2018-05-22T19:54:20Z,_m3,1,10,100
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,long,long,long
#group,false,false,true,true,false,true,false,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_measurement,v0,v1,v2
,,0,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:30Z,_m0,1,10,100
,,0,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:40Z,_m0,2,20,100
,,0,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:50Z,_m0,3,30,100
,,0,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,_m0,4,40,100
,,0,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:10Z,_m0,5,50,100
,,0,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:20Z,_m0,6,60,100
,,1,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:30Z,_m1,1,10,100
,,1,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:40Z,_m1,2,20,100
,,1,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:50Z,_m1,3,30,100
,,1,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:54:00Z,_m1,4,40,100
,,1,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:54:10Z,_m1,5,50,100
,,1,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:54:30Z,_m1,6,60,100
,,1,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:54:40Z,_m1,7,70,100
,,1,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:50Z,_m1,8,80,100
,,1,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:54:00Z,_m1,9,90,100
,,2,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:10Z,_m2,1,10,100
,,2,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:30Z,_m2,2,20,100
,,2,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:54:40Z,_m2,3,30,100
,,2,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:50Z,_m2,4,40,100
,,3,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:54:00Z,_m3,1,10,100
,,3,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:54:10Z,_m3,2,20,100
,,3,2018-05-20T00:00:00Z,2030-01-01T00:00:00Z,2018-05-22T19:54:20Z,_m3,3,30,100
"
t_cumulative_sum = (table=<-) =>
	(table
		|> range(start: 2018-05-20T00:00:00Z)
		|> cumulativeSum(columns: ["v0", "v1"]))

test cumulative_sum = {input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_cumulative_sum}