package testdata_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,double,dateTime:RFC3339
#default,humidity,,,
,result,table,_value,_time
,,0,36.59,2019-08-20T13:34:00Z
,,0,36.61666666666667,2019-08-20T13:34:30Z
,,0,36.57,2019-08-20T13:35:00Z
,,0,36.684000000000005,2019-08-20T13:35:30Z
,,0,36.66166666666666,2019-08-20T13:36:00Z
,,0,36.652,2019-08-20T13:36:30Z
,,0,36.747499999999995,2019-08-20T13:37:00Z
,,0,36.69,2019-08-20T13:37:30Z
,,0,36.664,2019-08-20T13:38:00Z
,,0,36.644999999999996,2019-08-20T13:38:30Z

#datatype,string,long,double,dateTime:RFC3339
#default,Temperature,,,
,result,table,_value,_time
,,0,92.71,2019-08-20T13:34:00Z
,,0,92.7,2019-08-20T13:34:30Z
,,0,92.7,2019-08-20T13:35:00Z
,,0,92.7,2019-08-20T13:35:30Z
,,0,92.70166666666667,2019-08-20T13:36:00Z
,,0,92.708,2019-08-20T13:36:30Z
,,0,92.72500000000001,2019-08-20T13:37:00Z
,,0,92.76,2019-08-20T13:37:30Z
,,0,92.80799999999999,2019-08-20T13:38:00Z
,,0,92.87,2019-08-20T13:38:30Z
"

outData = "
#datatype,string,long,dateTime:RFC3339,double
#default,Heat-Index,,,
,result,table,_time,_value
,,0,2019-08-20T13:34:00Z,93.40072999999998
,,0,2019-08-20T13:34:30Z,93.39098333333332
,,0,2019-08-20T13:35:00Z,93.38878999999999
,,0,2019-08-20T13:35:30Z,93.39414799999999
,,0,2019-08-20T13:36:00Z,93.39493166666666
,,0,2019-08-20T13:36:30Z,93.401444
,,0,2019-08-20T13:37:00Z,93.42463250000002
,,0,2019-08-20T13:37:30Z,93.46042999999999
,,0,2019-08-20T13:38:00Z,93.512008
,,0,2019-08-20T13:38:30Z,93.579315
"


shi = (table=<-) => {
    t1 = table
		|> range(start: 2019-08-20T13:34:00Z, stop: 2019-08-20T13:38:30Z)
		|> drop(columns: ["_start", "_stop"])
        |> map(fn: (r) => ({humidity: r._value, _time: r._time}))
		
    t2 = table
        |> range(start: 2019-08-20T13:34:00Z, stop: 2019-08-20T13:38:30Z)
        |> drop(columns: ["_start", "_stop"])
        |> map(fn: (r) => ({temperature: r._value, _time: r._time}))
    return simpleHeatIndex()

test _simpleHeatIndex = () =>
	({input: testing.loadStorage(csv: [humidityData, temperatureData]), want: testing.loadMem(csv: outData), fn: t_cov})
