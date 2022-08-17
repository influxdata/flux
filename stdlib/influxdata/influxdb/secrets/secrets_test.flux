package secrets_test


import "testing"
import "influxdata/influxdb/secrets"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T19:53:26Z,1.83,load1,system
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,false
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,token
,,0,2018-05-22T19:53:26Z,1.83,load1,system,mysecrettoken
"
token = secrets.get(key: "token")

// Passes in flux, fails in C2 and OSS
testcase secrets {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> set(key: "token", value: token)
            |> drop(columns: ["_start", "_stop"])
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
