import "testing"

inData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,long
#group,false,false,false,false,true,true,true,false,false
#default,,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:30Z,
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:40Z,-25
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:50Z,46
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:00Z,-2
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:10Z,
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:20Z,-53
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:30Z,17
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:40Z,-44
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:50Z,-99
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:00Z,-85
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:10Z,
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:20Z,99
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,long
#group,false,false,false,false,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:30Z,-1
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:40Z,-25
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:50Z,46
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:00Z,-2
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:10Z,-1
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:20Z,-53
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:30Z,17
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:40Z,-44
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:50Z,-99
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:00Z,-85
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:10Z,-1
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:20Z,99
"

option now = () => 2018-12-19T22:15:00Z

t_fill_int = (table=<-) => table
  |> range(start: -5m)
  |> fill(column: "_value", value: -1)

testFn = testing.test

testFn(name: "fill",
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    testFn: t_fill_int)