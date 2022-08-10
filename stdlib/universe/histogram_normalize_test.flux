package universe_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,string,double,string
#group,false,false,true,true,false,true
#default,_result,,,,,
,result,table,_time,_field,theValue,_measurement
,,0,2018-05-22T19:53:00Z,x_duration_seconds,0,m0
,,0,2018-05-22T19:53:00Z,x_duration_seconds,1,m0
,,1,2018-05-22T19:53:00Z,y_duration_seconds,0,m0
,,1,2018-05-22T19:53:00Z,y_duration_seconds,0,m0
,,1,2018-05-22T19:53:00Z,y_duration_seconds,1.5,m0
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,string,double,double,string
#group,false,false,true,true,false,false,true
#default,_result,,,,,,
,result,table,_time,_field,ub,theCount,_measurement
,,0,2018-05-22T19:53:00Z,x_duration_seconds,-1,0,m0
,,0,2018-05-22T19:53:00Z,x_duration_seconds,0,0.5,m0
,,0,2018-05-22T19:53:00Z,x_duration_seconds,1,1,m0
,,0,2018-05-22T19:53:00Z,x_duration_seconds,2,1,m0
,,1,2018-05-22T19:53:00Z,y_duration_seconds,-1,0,m0
,,1,2018-05-22T19:53:00Z,y_duration_seconds,0,0.6666666666666666,m0
,,1,2018-05-22T19:53:00Z,y_duration_seconds,1,0.6666666666666666,m0
,,1,2018-05-22T19:53:00Z,y_duration_seconds,2,1,m0
"

testcase histogram {
    option testing.tags = ["skip"]

    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:53:01Z)
            |> histogram(
                bins: [-1.0, 0.0, 1.0, 2.0],
                normalize: true,
                column: "theValue",
                countColumn: "theCount",
                upperBoundColumn: "ub",
            )
            |> drop(columns: ["_start", "_stop"])
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
