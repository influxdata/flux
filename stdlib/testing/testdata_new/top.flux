package main
import "testing"

option now = () =>
	(2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,_field
,,0,2018-05-22T19:54:16Z,20000,used
,,0,2018-05-22T19:53:56Z,55000,used
,,0,2018-05-22T19:54:06Z,20000,used
,,0,2018-05-22T19:53:26Z,35000,used
,,0,2018-05-22T19:53:46Z,70000,used
,,0,2018-05-22T19:53:36Z,15000,used
,,1,2018-05-22T19:54:16Z,20,used_percent
,,1,2018-05-22T19:53:56Z,55,used_percent
,,1,2018-05-22T19:54:06Z,20,used_percent
,,1,2018-05-22T19:53:26Z,35,used_percent
,,1,2018-05-22T19:53:46Z,70,used_percent
,,1,2018-05-22T19:53:36Z,15,used_percent
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string
#group,false,false,true,true,false,false,true
#default,_result,,,,,,
,result,table,_start,_stop,_time,_value,_field
,,0,2018-05-22T19:53:24.421470485Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,70000,used
,,0,2018-05-22T19:53:24.421470485Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,55000,used
,,1,2018-05-22T19:53:24.421470485Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,70,used_percent
,,1,2018-05-22T19:53:24.421470485Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,55,used_percent
"
t_top = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:24.421470485Z)
		|> top(n: 2))

test top = {input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_top}