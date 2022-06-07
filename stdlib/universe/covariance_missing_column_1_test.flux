package universe_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,double,string
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,x,_measurement
,,0,2018-05-22T19:53:26Z,0,cpu
,,0,2018-05-22T19:53:36Z,0,cpu
,,0,2018-05-22T19:53:46Z,2,cpu
,,0,2018-05-22T19:53:56Z,7,cpu
"
outData =
    "
#datatype,string,string
#group,true,true
#default,,
,error,reference
,specified column does not exist in table: y,
"

testcase covariance_missing_column_1 {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> covariance(columns: ["x", "r"])
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
