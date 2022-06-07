package universe_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,string,double,double,string
#group,false,false,true,true,false,false,true
#default,_result,,,,,,
,result,table,_time,_field,_value,le,_measurement
,,0,2018-05-22T19:53:00Z,x_duration_seconds,10,-80,mm
,,0,2018-05-22T19:53:00Z,x_duration_seconds,11,-60,mm
,,0,2018-05-22T19:53:00Z,x_duration_seconds,12,-40,mm
,,0,2018-05-22T19:53:00Z,x_duration_seconds,13,-20,mm
,,0,2018-05-22T19:53:00Z,x_duration_seconds,14,-0,mm
,,0,2018-05-22T19:53:00Z,x_duration_seconds,15,20,mm
,,0,2018-05-22T19:53:00Z,x_duration_seconds,16,40,mm
,,0,2018-05-22T19:53:00Z,x_duration_seconds,17,60,mm
,,0,2018-05-22T19:53:00Z,x_duration_seconds,18,80,mm
,,0,2018-05-22T19:53:00Z,x_duration_seconds,19,+Inf,mm
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,double,string
#group,false,false,true,true,true,true,false,true
#default,_result,,,,,,,
,result,table,_start,_stop,_time,_field,_value,_measurement
,,0,2018-05-22T19:53:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,x_duration_seconds,-90.5,mm
"

testcase histogram_quantile_minvalue {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:00Z)
            |> histogramQuantile(quantile: 0.25, minValue: -100.0)
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
