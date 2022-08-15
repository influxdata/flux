package universe_test


import "testing"
import "csv"

inData =
    "
#datatype,string,long,dateTime:RFC3339,double,string
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,y,_measurement
,,0,2018-05-22T19:53:26Z,0,cpu
,,0,2018-05-22T19:53:36Z,0,cpu
,,0,2018-05-22T19:53:46Z,2,cpu
,,0,2018-05-22T19:53:56Z,7,cpu
"

testcase covariance_missing_column_2 {
    testing.shouldError(
        fn: () =>
            csv.from(csv: inData)
                |> covariance(columns: ["x", "r"])
                |> tableFind(fn: (key) => true),
        want: /covariance: specified column does not exist in table: x$/,
    )
}
