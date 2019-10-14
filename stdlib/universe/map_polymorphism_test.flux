package universe_test
 
import "testing"

option now = () => 2030-01-01T00:00:00Z

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T19:53:26Z,49,load1,system
,,0,2018-05-22T19:53:36Z,50,load1,system
,,0,2018-05-22T19:53:46Z,51,load1,system
"

outData = "
#datatype,string,long,string
#group,false,false,false
#default,_result,,
,result,table,out
,,0,Y
,,0,N
,,0,N
"

f = (r) => r._value < 50

// test that map can call polymorphic functions
t_map = (table=<-) => table
    |> range(start: 2018-05-22T19:53:16Z)
    |> map(fn: (r) => ({out: if f(r: r) then "Y" else "N"}))

test _map = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map})
