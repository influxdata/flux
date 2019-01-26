import "testing"

option now = () => 2030-01-01T00:00:00Z

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:53:00Z,_m,FF,1
,,0,2018-05-22T19:53:10Z,_m,FF,1
,,0,2018-05-22T19:53:20Z,_m,FF,1
,,0,2018-05-22T19:53:30Z,_m,FF,1
,,0,2018-05-22T19:53:40Z,_m,FF,1
,,0,2018-05-22T19:53:50Z,_m,FF,1
,,1,2018-05-22T19:53:00Z,_m,QQ,1
,,1,2018-05-22T19:53:10Z,_m,QQ,1
,,1,2018-05-22T19:53:20Z,_m,QQ,1
,,1,2018-05-22T19:53:30Z,_m,QQ,1
,,1,2018-05-22T19:53:40Z,_m,QQ,1
,,1,2018-05-22T19:53:50Z,_m,QQ,1
,,1,2018-05-22T19:54:00Z,_m,QQ,1
,,1,2018-05-22T19:54:10Z,_m,QQ,1
,,1,2018-05-22T19:54:20Z,_m,QQ,1
,,2,2018-05-22T19:53:00Z,_m,RR,1
,,2,2018-05-22T19:53:10Z,_m,RR,1
,,2,2018-05-22T19:53:20Z,_m,RR,1
,,2,2018-05-22T19:53:30Z,_m,RR,1
,,3,2018-05-22T19:53:40Z,_m,SR,1
,,3,2018-05-22T19:53:50Z,_m,SR,1
,,3,2018-05-22T19:54:00Z,_m,SR,1
"
outData = "
#datatype,string,long,string,string,double
#group,false,false,true,true,false
#default,_result,,,,
,result,table,_measurement,_field,_value
,,0,_m,FF,5
,,1,_m,QQ,8
,,2,_m,RR,3
,,3,_m,SR,2
"

t_integral = (table=<-) =>
  table
  |> integral(unit: 10s)

testing.run(
    name: "integral",
     input: testing.loadStorage(csv: inData),
     want: testing.loadMem(csv: outData),
     testFn:  t_integral,
)
