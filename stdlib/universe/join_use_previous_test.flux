package universe_test

import "csv"
import "testing"

option now = () => (2020-05-15T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double
#group,false,true,true,true
#default,,,,
,result,table,_time,value
,_result,0,2020-05-14T17:23:00Z,1
,_result,0,2020-05-14T17:40:00Z,2
,_result,0,2020-05-14T17:41:00Z,3
,_result,0,2020-05-14T17:42:00Z,4
,_result,0,2020-05-14T17:43:00Z,5
,_result,0,2020-05-14T17:44:00Z,6
,_result,0,2020-05-14T17:45:00Z,
,_result,0,2020-05-14T17:46:00Z,7
,_result,0,2020-05-14T17:47:00Z,
,_result,0,2020-05-14T17:48:00Z,8
,_result,0,2020-05-14T17:49:00Z,9
,_result,0,2020-05-14T17:50:00Z,10

#datatype,string,long,dateTime:RFC3339,boolean
#group,false,true,true,true
#default,,,,
,result,table,_time,flag
,_result,0,2020-05-14T17:23:00Z,false
,_result,0,2020-05-14T17:40:00Z,
,_result,0,2020-05-14T17:41:00Z,
,_result,0,2020-05-14T17:42:00Z,
,_result,0,2020-05-14T17:43:00Z,
,_result,0,2020-05-14T17:44:00Z,
,_result,0,2020-05-14T17:45:00Z,true
,_result,0,2020-05-14T17:46:00Z,false
,_result,0,2020-05-14T17:47:00Z,true
,_result,0,2020-05-14T17:48:00Z,false
,_result,0,2020-05-14T17:49:00Z,
,_result,0,2020-05-14T17:50:00Z,
"

outData = "
#datatype,string,long,dateTime:RFC3339,double,boolean
#group,false,false,false,false,false
#default,_result,,,,
,result,table,_time,value,flag
,_result,0,2020-05-14T17:23:00Z,1,false
,_result,0,2020-05-14T17:40:00Z,2,false
,_result,0,2020-05-14T17:41:00Z,3,false
,_result,0,2020-05-14T17:42:00Z,4,false
,_result,0,2020-05-14T17:43:00Z,5,false
,_result,0,2020-05-14T17:44:00Z,6,false
,_result,0,2020-05-14T17:45:00Z,,true
,_result,0,2020-05-14T17:46:00Z,7,false
,_result,0,2020-05-14T17:47:00Z,,true
,_result,0,2020-05-14T17:48:00Z,8,false
,_result,0,2020-05-14T17:49:00Z,9,false
,_result,0,2020-05-14T17:50:00Z,10,false
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
    return join(tables: {left: lhs, right: rhs}, on: ["_time"], method: "inner") |> group(columns: []) |> fill(column: "flag", usePrevious: true)
}

test _join_use_previous_test = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_join_use_previous_test})