package universe_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,double,string,string,double,double
#group,false,false,false,false,true,true,false,false
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,x,y
,,0,2018-05-22T19:53:26Z,34.98234271799806,f,m1,1,2
,,0,2018-05-22T19:53:36Z,34.98234941084654,f,m1,2,4
,,0,2018-05-22T19:53:46Z,34.982447293755506,f,m1,5,4
,,0,2018-05-22T19:53:56Z,34.982447293755506,f,m1,12,80
,,0,2018-05-22T19:54:06Z,34.98204153981662,f,m1,11,0
,,0,2018-05-22T19:54:16Z,34.982252364543626,f,m1,33,42
,,1,2018-05-22T19:53:26Z,34.98234271799806,f,m2,1,2
,,1,2018-05-22T19:53:36Z,34.98234941084654,f,m2,2,4
,,1,2018-05-22T19:53:46Z,34.982447293755506,f,m2,5,4
,,1,2018-05-22T19:53:56Z,34.982447293755506,f,m2,12,80
,,1,2018-05-22T19:54:06Z,34.98204153981662,f,m2,11,0
,,1,2018-05-22T19:54:16Z,34.982252364543626,f,m2,33,42
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,double,double
#group,false,false,true,true,false,false,true,true,false,false
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,x,y
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,34.98234941084654,f,m1,1,2
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,34.982447293755506,f,m1,3,0
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,34.982447293755506,f,m1,7,76
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,34.98204153981662,f,m1,-1,-80
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,34.982252364543626,f,m1,22,42
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,34.98234941084654,f,m2,1,2
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,34.982447293755506,f,m2,3,0
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,34.982447293755506,f,m2,7,76
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,34.98204153981662,f,m2,-1,-80
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,34.982252364543626,f,m2,22,42
"

// Passes in flux, fails in C2 and OSS
testcase difference_columns {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> difference(columns: ["x", "y"])
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
