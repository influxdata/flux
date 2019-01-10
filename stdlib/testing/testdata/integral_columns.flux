import "testing"

option now = () => 2030-01-01T00:00:00Z

inData = "
#datatype,string,long,dateTime:RFC3339,string,double,double
#group,false,false,false,true,false,false
#default,_result,,,,,
,result,table,_time,_measurement,v1,v2
,,0,2018-05-22T19:53:00Z,_m0,0,1
,,0,2018-05-22T19:53:10Z,_m0,1,1
,,0,2018-05-22T19:53:20Z,_m0,2,1
,,0,2018-05-22T19:53:30Z,_m0,3,1
,,0,2018-05-22T19:53:40Z,_m0,4,1
,,0,2018-05-22T19:53:50Z,_m0,5,1
,,1,2018-05-22T19:53:00Z,_m1,0,1
,,1,2018-05-22T19:53:10Z,_m1,2,1
,,1,2018-05-22T19:53:20Z,_m1,4,1
,,1,2018-05-22T19:53:30Z,_m1,6,1
,,1,2018-05-22T19:53:40Z,_m1,8,1
,,1,2018-05-22T19:53:50Z,_m1,6,1
,,1,2018-05-22T19:54:00Z,_m1,4,1
,,1,2018-05-22T19:54:10Z,_m1,2,1
,,1,2018-05-22T19:54:20Z,_m1,0,1
,,2,2018-05-22T19:53:00Z,_m2,0,1
,,2,2018-05-22T19:53:10Z,_m2,8,1
,,2,2018-05-22T19:53:20Z,_m2,2,1
,,2,2018-05-22T19:53:30Z,_m2,6,1
,,3,2018-05-22T19:53:40Z,_m3,1,1
,,3,2018-05-22T19:53:50Z,_m3,1,1
,,3,2018-05-22T19:54:00Z,_m3,1,1
"
outData = "
#datatype,string,long,string,double,double
#group,false,false,true,false,false
#default,_result,,,,
,result,table,_measurement,v1,v2
,,0,_m0,12.5,5
,,1,_m1,32,8
,,2,_m2,13,3
,,3,_m3,2,2
"

t_integral_columns = (table=<-) =>
  table
  |> integral(columns: ["v1", "v2"], unit: 10s)

testing.test(
    name: "integral_columns",
     input: testing.loadStorage(csv: inData),
     want: testing.loadMem(csv: outData),
     testFn:  t_integral_columns,
)
