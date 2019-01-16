import "testing"

inData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,long
#group,false,false,true,true,false,true,false
#default,,,,,,,
,result,table,_start,_stop,_time,_measurement,values
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:30Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:40Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:50Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:00Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:10Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:20Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:30Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:40Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:50Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:00Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:10Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:30Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:40Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:50Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:00Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:10Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:30Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:40Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:50Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:00Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:10Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:20Z,_m,1
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,long
#group,false,false,true,true,false,true,false
#default,_result,,,,,,
,result,table,_start,_stop,_time,_measurement,values
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:30Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:40Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:50Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:00Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:10Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:20Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:30Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:40Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:50Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:00Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:10Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:30Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:40Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:50Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:00Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:10Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:30Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:40Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:53:50Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:00Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:10Z,_m,1
,,0,2018-05-22T19:53:00Z,2018-05-22T19:54:50Z,2018-05-22T19:54:20Z,_m,1
"

t_cumulative_sum_noop = (table=<-) => table
  |> cumulativeSum()

testing.test(
    name: "cumulative_sum_noop",
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    testFn: t_cumulative_sum_noop)