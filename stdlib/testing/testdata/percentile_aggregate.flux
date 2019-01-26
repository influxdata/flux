import "testing"

option now = () => 2030-01-01T00:00:00Z

inData = "
#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,SOYcRk,NC7N,2018-12-18T21:12:45Z,-61.68790887989735
,,0,SOYcRk,NC7N,2018-12-18T21:12:55Z,-6.3173755351186465
,,0,SOYcRk,NC7N,2018-12-18T21:13:05Z,-26.049728557657513
,,0,SOYcRk,NC7N,2018-12-18T21:13:15Z,114.285955884979
,,0,SOYcRk,NC7N,2018-12-18T21:13:25Z,16.140262630578995
,,0,SOYcRk,NC7N,2018-12-18T21:13:35Z,29.50336437998469
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,double
#group,false,false,true,true,true,true,false
#default,_result,,,,,,
,result,table,_start,_stop,_measurement,_field,_value
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SOYcRk,NC7N,26.162588942633263
"


t_percentile = (table=<-) => table
    |> range(start: 2018-01-01T00:00:00Z)
    |> percentile(percentile: 0.75, method: "exact_mean")

testing.run(
    name: "percentile_aggregate",
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    testFn: t_percentile)
