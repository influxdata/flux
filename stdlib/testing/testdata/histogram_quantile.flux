import "testing"

option now = () => 2030-01-01T00:00:00Z

inData = "
#datatype,string,long,dateTime:RFC3339,string,double,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_time,_field,count,upperBound
,,0,2018-05-22T19:53:00Z,x_duration_seconds,1,0.1
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,0.2
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,0.3
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,0.4
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,0.5
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,0.6
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,0.7
,,0,2018-05-22T19:53:00Z,x_duration_seconds,8,0.8
,,0,2018-05-22T19:53:00Z,x_duration_seconds,10,0.9
,,0,2018-05-22T19:53:00Z,x_duration_seconds,10,+Inf
,,1,2018-05-22T19:53:00Z,y_duration_seconds,0,-Inf
,,1,2018-05-22T19:53:00Z,y_duration_seconds,10,0.2
,,1,2018-05-22T19:53:00Z,y_duration_seconds,15,0.4
,,1,2018-05-22T19:53:00Z,y_duration_seconds,25,0.6
,,1,2018-05-22T19:53:00Z,y_duration_seconds,35,0.8
,,1,2018-05-22T19:53:00Z,y_duration_seconds,45,1
,,1,2018-05-22T19:53:00Z,y_duration_seconds,45,+Inf
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,double
#group,false,false,true,true,true,true,false
#default,_result,,,,,,
,result,table,_start,_stop,_time,_field,quant
,,0,2018-05-22T19:53:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,x_duration_seconds,0.8500000000000001
,,1,2018-05-22T19:53:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,y_duration_seconds,0.91
"

t_histogram_quantile = (table=<-) =>
  table
    |> range(start: 2018-05-22T19:53:00Z)
    |> histogramQuantile(quantile:0.90,
           upperBoundColumn:"upperBound",
           countColumn:"count",
           valueColumn:"quant")

testing.test(name: "histogram_quantile",
            input: testing.loadStorage(csv: inData),
            want: testing.loadMem(csv: outData),
            testFn: t_histogram_quantile)
