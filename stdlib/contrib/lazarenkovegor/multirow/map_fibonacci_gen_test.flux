package multirow_test


import "testing"
import "contrib/lazarenkovegor/multirow"

inData =
    "
#datatype,string,long,string,string,dateTime:RFC3339,string
#group,false,false,false,false,false,false
#default,_result,0,,,2000-01-01T00:00:00Z,m0
,result,table,_field,_value,_time,_measurement
,,,test1,test10,,
,,,test1,test11,,
,,,test2,test12,,
,,,test2,test13,,
,,,test2,test14,,
,,,test2,test15,,
,,,test2,test16,,
,,,test2,test17,,
,,,test2,test18,,
"

outData =
    "
#datatype,string,long,string,string,dateTime:RFC3339,string,long
#group,false,false,false,false,false,false,false
#default,_result,0,,m0,2000-01-01T00:00:00Z,,
,result,table,_field,_measurement,_time,_value,fibonacci
,,,test1,,,test10,0
,,,test1,,,test11,1
,,,test2,,,test12,1
,,,test2,,,test13,2
,,,test2,,,test14,3
,,,test2,,,test15,5
,,,test2,,,test16,8
,,,test2,,,test17,13
,,,test2,,,test18,21
"

t_map = (table=<-) =>
    table
        |> drop(columns: ["_start", "_stop"])
        |> multirow.map(
            fn: (previous, index, row) =>
                ({row with fibonacci: if index == 1 then 1 else previous.fibonacci + previous.prev,
                    prev: previous.fibonacci,
                }),
            init: {prev: 0, fibonacci: 0},
            virtual: ["prev"],
        )

test _map = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map})
