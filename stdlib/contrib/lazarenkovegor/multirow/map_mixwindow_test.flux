package multirow_test


import "testing"
import "contrib/lazarenkovegor/multirow"

inData =
    "
#datatype,string,long,string,long,dateTime:RFC3339,string
#group,false,false,false,false,false,false
#default,_result,0,,,2000-01-01T00:00:00Z,m0
,result,table,_field,_value,_time,_measurement
,,,test1,1,2020-08-02T17:22:00Z,
,,,test1,2,2020-08-02T17:22:00Z,
,,,test2,3,2020-08-02T17:22:01Z,
,,,test2,4,2020-08-02T17:22:01Z,
,,,test2,5,2020-08-02T17:22:01Z,
,,,test2,6,2020-08-02T17:22:02Z,
,,,test2,7,2020-08-02T17:22:03Z,
,,,test2,8,2020-08-02T17:22:03Z,
,,,test2,9,2020-08-02T17:22:04Z,
"

outData =
    "
#datatype,string,long,string,long,dateTime:RFC3339
#group,false,false,false,false,false
#default,_result,0,,,2000-01-01T00:00:00Z
,result,table,_field,_value,_time
,,,test1,2,2020-08-02T17:22:00Z
,,,test1,3,2020-08-02T17:22:00Z
,,,test2,4,2020-08-02T17:22:01Z
,,,test2,5,2020-08-02T17:22:01Z
,,,test2,6,2020-08-02T17:22:01Z
,,,test2,5,2020-08-02T17:22:02Z
,,,test2,3,2020-08-02T17:22:03Z
,,,test2,4,2020-08-02T17:22:03Z
,,,test2,3,2020-08-02T17:22:04Z
"

t_map = (table=<-) =>
    table
        |> drop(columns: ["_start", "_stop"])
        |> multirow.map(
            left: 1s,
            right: 1,
            fn: (window, row) => window |> count() |> map(fn: (r) => ({r with _time: row._time, _field: row._field})),
        )

test _map = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map})
