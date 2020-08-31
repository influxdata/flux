package universe_test

import "testing"

option now = () => (2020-05-15T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,true,true,true,true,true
#default,,,,,,
,result,table,_time,value,_field,_measurement
,_result,0,2020-05-14T17:23:00Z,1,a,a
,_result,0,2020-05-14T17:40:00Z,2,a,a
,_result,0,2020-05-14T17:41:00Z,3,a,a
,_result,0,2020-05-14T17:42:00Z,4,a,a
,_result,0,2020-05-14T17:43:00Z,5,a,a
,_result,0,2020-05-14T17:44:00Z,6,a,a
,_result,0,2020-05-14T17:45:00Z,,a,a
,_result,0,2020-05-14T17:46:00Z,7,a,a
,_result,0,2020-05-14T17:47:00Z,,a,a
,_result,0,2020-05-14T17:48:00Z,8,a,a
,_result,0,2020-05-14T17:49:00Z,9,a,a
,_result,0,2020-05-14T17:50:00Z,10,a,a

#datatype,string,long,dateTime:RFC3339,boolean,string,string
#group,false,true,true,true,true,true
#default,,,,,,
,result,table,_time,flag,_field,_measurement
,_result,0,2020-05-14T17:23:00Z,false,a,a
,_result,0,2020-05-14T17:40:00Z,,a,a
,_result,0,2020-05-14T17:41:00Z,,a,a
,_result,0,2020-05-14T17:42:00Z,,a,a
,_result,0,2020-05-14T17:43:00Z,,a,a
,_result,0,2020-05-14T17:44:00Z,,a,a
,_result,0,2020-05-14T17:45:00Z,true,a,a
,_result,0,2020-05-14T17:46:00Z,false,a,a
,_result,0,2020-05-14T17:47:00Z,true,a,a
,_result,0,2020-05-14T17:48:00Z,false,a,a
,_result,0,2020-05-14T17:49:00Z,,a,a
,_result,0,2020-05-14T17:50:00Z,,a,a
"

outData = "
#datatype,string,long,dateTime:RFC3339,boolean,double,string,string
#group,false,false,false,false,false,false,false
#default,_result,,,,,,
,result,table,_time,flag,value,_field,_measurement
,_result,0,2020-05-14T17:23:00Z,false,1,a,a
,_result,0,2020-05-14T17:46:00Z,false,7,a,a
,_result,0,2020-05-14T17:48:00Z,false,8,a,a
"

t_join_use_previous_test = (tables=<-) => {
    lhs = tables
        |> range(start: 2020-05-01T00:00:00Z)
        |> filter(fn: (r) => exists r.value)
        |> drop(columns: ["_start", "_stop"])
    rhs = tables
        |> range(start: 2020-05-01T00:00:00Z)
        |> filter(fn: (r) => exists r.flag)
        |> drop(columns: ["_start", "_stop"])
    return join(tables: {left: lhs, right: rhs}, on: ["_time", "_field", "_measurement"], method: "inner") |> group(columns: []) |> fill(column: "flag", usePrevious: true)
}

test _join_use_previous_test = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_join_use_previous_test})