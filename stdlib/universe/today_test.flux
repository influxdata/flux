package universe_test


import "testing"

option now = () => 2030-01-01T12:00:00Z

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,0,usage_guest,cpu,cpu-total,host.local
"
outData = "
#datatype,string,long,string,string,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,dateTime:RFC3339
#group,false,false,true,true,false,true,true,false,true,true,false
#default,_result,,,,,,,,,,
,result,table,_field,_measurement,_time,_start,_stop,_value,cpu,host,today
,,0,usage_guest,cpu,2018-05-22T19:53:26Z,2018-05-22T19:53:26Z,2030-01-01T12:00:00Z,0,cpu-total,host.local,2030-01-01T00:00:00Z
"
t_today = (table=<-) => table
    |> range(start: 2018-05-22T19:53:26Z)
    |> map(fn: (r) => ({r with today: today()}))

test _today = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_today})
