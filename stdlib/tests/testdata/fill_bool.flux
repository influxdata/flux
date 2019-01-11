inData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,boolean
#group,false,false,false,false,true,true,true,false,false
#default,,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:30Z,false
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:40Z,true
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:50Z,false
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:00Z,false
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:10Z,
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:20Z,true
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:30Z,false
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:40Z,true
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:50Z,
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:00Z,true
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:10Z,true
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:20Z,
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,boolean
#group,false,false,false,false,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:30Z,false
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:40Z,true
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:13:50Z,false
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:00Z,false
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:10Z,false
,,0,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server01,2018-12-19T22:14:20Z,true
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:30Z,false
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:40Z,true
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:13:50Z,false
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:00Z,true
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:10Z,true
,,1,2018-12-19T22:13:30Z,2018-12-19T22:14:20Z,m1,f1,server02,2018-12-19T22:14:20Z,false
"

option now = () => 2018-12-19T22:15:00Z

t_fill_bool = (table=<-) => table
  |> range(start: -5m)
  |> fill(value: false)

testingTest(name: "fill",
            input: testLoadStorage(csv: inData),
            want: testLoadMem(csv: outData),
            test: t_fill_bool)