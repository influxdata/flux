import "testing"

option now = () => 2030-01-01T00:00:00Z

inData = "
#datatype,string,long,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m1,f1,server01,2018-12-19T22:13:30Z,-61.68790887989735
,,0,m1,f1,server01,2018-12-19T22:13:40Z,-6.3173755351186465
,,0,m1,f1,server01,2018-12-19T22:13:50Z,
,,0,m1,f1,server01,2018-12-19T22:14:00Z,114.285955884979
,,0,m1,f1,server01,2018-12-19T22:14:10Z,16.140262630578995
,,0,m1,f1,server01,2018-12-19T22:14:20Z,29.50336437998469
,,1,m1,f1,server02,2018-12-19T22:13:30Z,
,,1,m1,f1,server02,2018-12-19T22:13:40Z,49.460104214779086
,,1,m1,f1,server02,2018-12-19T22:13:50Z,-36.564150808873954
,,1,m1,f1,server02,2018-12-19T22:14:00Z,34.319039251798635
,,1,m1,f1,server02,2018-12-19T22:14:10Z,
,,1,m1,f1,server02,2018-12-19T22:14:20Z,41.91029522104053
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:30Z,-61.68790887989735
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:40Z,-6.3173755351186465
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:13:50Z,0.01
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:00Z,114.285955884979
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:10Z,16.140262630578995
,,0,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server01,2018-12-19T22:14:20Z,29.50336437998469
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:30Z,0.01
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:40Z,49.460104214779086
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:13:50Z,-36.564150808873954
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:00Z,34.319039251798635
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:10Z,0.01
,,1,2018-12-15T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,server02,2018-12-19T22:14:20Z,41.91029522104053
"


t_fill_float = (table=<-) => table
  |> range(start: 2018-12-15T00:00:00Z)
  |> fill(value: 0.01)

testing.test(name: "fill",
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    testFn: t_fill_float)
