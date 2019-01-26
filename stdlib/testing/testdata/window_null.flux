import "testing"

inData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,true,false,false
#default,,,,,,,,
,result,table,_start,_stop,_measurement,_field,_time,_value
,,0,2019-01-15T21:39:30Z,2019-01-15T21:40:20Z,dlC,lDQVwm,2019-01-15T21:39:30Z,-61.68790887989735
,,0,2019-01-15T21:39:30Z,2019-01-15T21:40:20Z,dlC,lDQVwm,2019-01-15T21:39:40Z,-6.3173755351186465
,,0,2019-01-15T21:39:30Z,2019-01-15T21:40:20Z,dlC,lDQVwm,2019-01-15T21:39:50Z,-26.049728557657513
,,0,2019-01-15T21:39:30Z,2019-01-15T21:40:20Z,dlC,lDQVwm,2019-01-15T21:40:00Z,
,,0,2019-01-15T21:39:30Z,2019-01-15T21:40:20Z,dlC,lDQVwm,2019-01-15T21:40:10Z,114.285955884979
,,0,2019-01-15T21:39:30Z,2019-01-15T21:40:20Z,dlC,lDQVwm,2019-01-15T21:40:20Z,16.140262630578995

"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,true,false,false
#default,,,,,,,,
,result,table,_start,_stop,_measurement,_field,_time,_value
,,0,2019-01-15T21:39:30Z,2019-01-15T21:40:00Z,dlC,lDQVwm,2019-01-15T21:39:30Z,-61.68790887989735
,,0,2019-01-15T21:39:30Z,2019-01-15T21:40:00Z,dlC,lDQVwm,2019-01-15T21:39:40Z,-6.3173755351186465
,,0,2019-01-15T21:39:30Z,2019-01-15T21:40:00Z,dlC,lDQVwm,2019-01-15T21:39:50Z,-26.049728557657513
,,1,2019-01-15T21:40:00Z,2019-01-15T21:40:30Z,dlC,lDQVwm,2019-01-15T21:40:00Z,
,,1,2019-01-15T21:40:00Z,2019-01-15T21:40:30Z,dlC,lDQVwm,2019-01-15T21:40:10Z,114.285955884979
,,1,2019-01-15T21:40:00Z,2019-01-15T21:40:30Z,dlC,lDQVwm,2019-01-15T21:40:20Z,16.140262630578995

"

option now = () => 2019-01-15T21:40:32Z

t_window_null = (table=<-) => table
		|> range(start: -5m)
    |> window(every: 30s)

testing.run(
	name: "window_null",
	input: testing.loadStorage(csv: inData),
	want: testing.loadMem(csv: outData),
	testFn: t_window_null,
)
