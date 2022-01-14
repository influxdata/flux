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
"

outData =
    "
#datatype,string,long,string,string,dateTime:RFC3339,string
#group,false,false,false,false,false,false
#default,_result,0,,m0,2000-01-01T00:00:00Z,
,result,table,_field,_measurement,_time,_value
,,,test1,,,test11
,,,test2,,,test13
"

t_map = (table=<-) =>
    table
        |> drop(columns: ["_start", "_stop"])
        |> multirow.map(fn: (index, row) => if index % 2 > 0 then [row] else [])

test _map = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map})
