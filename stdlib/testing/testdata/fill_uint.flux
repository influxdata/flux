import "testing"

inData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,false,false,true,true,true,false,false
#default,,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:30Z,84
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:40Z,52
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:50Z,
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:00Z,62
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:10Z,22
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:20Z,78
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:30Z,
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:40Z,33
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:50Z,97
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:00Z,90
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:10Z,96
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:20Z,
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,false,false,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:30Z,84
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:40Z,52
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:50Z,0
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:00Z,62
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:10Z,22
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:20Z,78
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:30Z,0
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:40Z,33
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:50Z,97
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:00Z,90
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:10Z,96
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:20Z,0
"

option now = () => 2018-12-19T22:15:00Z

t_fill_uint = (table=<-) => table
  |> range(start: -5m)
  |> fill(column: "_value", value: uint(v:0))

testFn = testing.test

testFn(name: "fill",
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    testFn: t_fill_uint)