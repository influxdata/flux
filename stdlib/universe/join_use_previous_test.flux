package universe_test


import "testing"
import "csv"

option now = () => 2020-05-15T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,true,true,true,true,true
#default,,,,,,
,result,table,_time,value,_field,_measurement
,_result,0,2020-05-14T17:23:00Z,1,a,a
,_result,1,2020-05-14T17:40:00Z,2,a,a
,_result,2,2020-05-14T17:41:00Z,3,a,a
,_result,3,2020-05-14T17:42:00Z,4,a,a
,_result,4,2020-05-14T17:43:00Z,5,a,a
,_result,5,2020-05-14T17:44:00Z,6,a,a
,_result,6,2020-05-14T17:45:00Z,,a,a
,_result,7,2020-05-14T17:46:00Z,7,a,a
,_result,8,2020-05-14T17:47:00Z,,a,a
,_result,9,2020-05-14T17:48:00Z,8,a,a
,_result,10,2020-05-14T17:49:00Z,9,a,a
,_result,11,2020-05-14T17:50:00Z,10,a,a

#datatype,string,long,dateTime:RFC3339,boolean,string,string
#group,false,true,true,true,true,true
#default,,,,,,
,result,table,_time,flag,_field,_measurement
,_result,0,2020-05-14T17:23:00Z,false,a,a
,_result,1,2020-05-14T17:40:00Z,,a,a
,_result,2,2020-05-14T17:41:00Z,,a,a
,_result,3,2020-05-14T17:42:00Z,,a,a
,_result,4,2020-05-14T17:43:00Z,,a,a
,_result,5,2020-05-14T17:44:00Z,,a,a
,_result,6,2020-05-14T17:45:00Z,true,a,a
,_result,7,2020-05-14T17:46:00Z,false,a,a
,_result,8,2020-05-14T17:47:00Z,true,a,a
,_result,9,2020-05-14T17:48:00Z,false,a,a
,_result,10,2020-05-14T17:49:00Z,,a,a
,_result,11,2020-05-14T17:50:00Z,,a,a
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,boolean,double,string,string
#group,false,false,false,false,false,false,false
#default,_result,,,,,,
,result,table,_time,flag,value,_field,_measurement
,_result,0,2020-05-14T17:23:00Z,false,1,a,a
,_result,0,2020-05-14T17:46:00Z,false,7,a,a
,_result,0,2020-05-14T17:48:00Z,false,8,a,a
"

testcase join_use_previous_test {
    tables =
        csv.from(csv: inData)

    lhs =
        tables
            |> range(start: 2020-05-01T00:00:00Z)
            |> filter(fn: (r) => exists r.value)
            |> yield(name: "L")
            |> drop(columns: ["_start", "_stop"])
    rhs =
        tables
            |> range(start: 2020-05-01T00:00:00Z)
            |> filter(fn: (r) => exists r.flag)
            |> yield(name: "R")
            |> drop(columns: ["_start", "_stop"])
    got =
        join(
            tables: {left: lhs, right: rhs},
            on: ["_time", "_field", "_measurement"],
            method: "inner",
        )
            |> group(columns: [])
            |> fill(column: "flag", usePrevious: true)
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
