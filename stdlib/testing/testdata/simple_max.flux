import "testing"

option now = () => 2030-01-01T00:00:00Z

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-04-17T00:00:00Z,m1,f1,42
,,0,2018-04-17T00:00:01Z,m1,f1,43
"
outData = "
#datatype,string,long,dateTime:RFC3339,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_start,_measurement,_time,max
,,0,2018-04-17T00:00:00Z,m1,2018-04-17T00:00:01Z,43
"

simple_max = (table=<-) =>
  table
  |> range(start:2018-04-17T00:00:00Z)
  |> group(columns: ["_measurement", "_start"])
  |> max(column: "_value")
  |> map(fn: (r) => ({_time: r._time,max:r._value}))

testing.run(
    name: "simple_max",
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    testFn: simple_max)
