package universe_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

testcase histogram_quantile {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,string,double,double,string
#group,false,false,true,true,false,false,true
#default,_result,,,,,,
,result,table,_time,_field,count,upperBound,_measurement
,,0,2018-05-22T19:53:00Z,x_duration_seconds,1,0.1,l
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,0.2,l
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,0.3,l
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,0.4,l
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,0.5,l
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,0.6,l
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,0.7,l
,,0,2018-05-22T19:53:00Z,x_duration_seconds,8,0.8,l
,,0,2018-05-22T19:53:00Z,x_duration_seconds,10,0.9,l
,,0,2018-05-22T19:53:00Z,x_duration_seconds,10,+Inf,l
,,1,2018-05-22T19:53:00Z,y_duration_seconds,0,-Inf,l
,,1,2018-05-22T19:53:00Z,y_duration_seconds,10,0.2,l
,,1,2018-05-22T19:53:00Z,y_duration_seconds,15,0.4,l
,,1,2018-05-22T19:53:00Z,y_duration_seconds,25,0.6,l
,,1,2018-05-22T19:53:00Z,y_duration_seconds,35,0.8,l
,,1,2018-05-22T19:53:00Z,y_duration_seconds,45,1,l
,,1,2018-05-22T19:53:00Z,y_duration_seconds,45,+Inf,l
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,double,string
#group,false,false,true,true,true,true,false,true
#default,_result,,,,,,,
,result,table,_start,_stop,_time,_field,quant,_measurement
,,0,2018-05-22T19:53:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,x_duration_seconds,0.8500000000000001,l
,,1,2018-05-22T19:53:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,y_duration_seconds,0.91,l
"
    got =
        csv.from(csv: inData)
            |> range(start: 2018-05-22T19:53:00Z)
            |> histogramQuantile(
                quantile: 0.9,
                upperBoundColumn: "upperBound",
                countColumn: "count",
                valueColumn: "quant",
            )
    want = csv.from(csv: outData)

    testing.diff(got, want)
}

testcase histogram_quantile_minvalue {
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
    got =
        csv.from(csv: inData)
            |> range(start: 2018-05-22T19:53:00Z)
            |> histogramQuantile(quantile: 0.25, minValue: -100.0)
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
