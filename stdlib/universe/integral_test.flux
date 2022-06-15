package universe_test


import "testing"
import "csv"

testcase integral {
    inData =
        "
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
    outData =
        "
#datatype,string,long,string,string,double
#group,false,false,true,true,false
#default,_result,,,,
,result,table,_measurement,_field,_value
,,0,_m,FF,5
,,1,_m,QQ,8
,,2,_m,RR,3
,,3,_m,SR,2
"
    got =
        csv.from(csv: inData)
            |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:55:00Z)
            |> integral(unit: 10s)
            |> drop(columns: ["_start", "_stop"])

    want = csv.from(csv: outData)

    testing.diff(want: want, got: got)
}

testcase integral_interpolate {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:53:00Z,_m,FF,1
,,0,2018-05-22T19:53:10Z,_m,FF,2
,,0,2018-05-22T19:53:20Z,_m,FF,4
,,0,2018-05-22T19:53:30Z,_m,FF,6
,,0,2018-05-22T19:53:40Z,_m,FF,8
,,0,2018-05-22T19:53:50Z,_m,FF,10
,,1,2018-05-22T19:53:10Z,_m,QQ,1
,,1,2018-05-22T19:53:20Z,_m,QQ,1
,,1,2018-05-22T19:53:30Z,_m,QQ,1
,,1,2018-05-22T19:53:40Z,_m,QQ,1
,,2,2018-05-22T19:53:00Z,_m,RR,1
,,2,2018-05-22T19:53:10Z,_m,RR,1
,,2,2018-05-22T19:53:20Z,_m,RR,1
,,2,2018-05-22T19:53:30Z,_m,RR,0
"
    outData =
        "
#datatype,string,long,string,string,double
#group,false,false,true,true,false
#default,_result,,,,
,result,table,_measurement,_field,_value
,,0,_m,FF,36.5
,,1,_m,QQ,6
,,2,_m,RR,-2
"
    got =
        csv.from(csv: inData)
            |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:54:00Z)
            |> integral(unit: 10s, interpolate: "linear")
            |> drop(columns: ["_start", "_stop"])

    want = csv.from(csv: outData)

    testing.diff(want: want, got: got)
}

// Test integral interpolation when there is only a single value in each group.
testcase integral_interpolate_single {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:53:00Z,_m,FF,1
,,1,2018-05-22T19:53:10Z,_m,QQ,-2
,,2,2018-05-22T19:53:00Z,_m,RR,3
"
    outData =
        "
#datatype,string,long,string,string,double
#group,false,false,true,true,false
#default,_result,,,,
,result,table,_measurement,_field,_value
,,0,_m,FF,6
,,1,_m,QQ,-12
,,2,_m,RR,18
"
    got =
        csv.from(csv: inData)
            |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:54:00Z)
            |> integral(unit: 10s, interpolate: "linear")
            |> drop(columns: ["_start", "_stop"])

    want = csv.from(csv: outData)

    testing.diff(want: want, got: got)
}
