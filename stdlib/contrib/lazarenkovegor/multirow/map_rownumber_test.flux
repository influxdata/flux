package multirow_test


import "testing"
import "contrib/lazarenkovegor/multirow"

inData =
    "
#datatype,string,long,string,string,dateTime:RFC3339,long
#group,false,false,true,false,false,false
#default,_result,0,,m0,2000-01-01T00:00:00Z,
,result,table,_field,_measurement,_time,_value
,,,test1,,,0
,,,test1,,,1
,,,test1,,,1
,,,test1,,,2
,,,test1,,,3
,,,test1,,,5
,,,test2,,,8
,,,test2,,,13
,,,test2,,,21
"

outData =
    "
#datatype,string,long,string,long
#group,false,false,true,false
#default,_result,0,,
,result,table,_field,row_number
,,0,test1,0
,,0,test1,1
,,0,test1,2
,,0,test1,3
,,0,test1,4
,,0,test1,5
,,1,test2,0
,,1,test2,1
,,1,test2,2
"

t_map = (table=<-) =>
    table
        |> drop(columns: ["_start", "_stop"])
        |> multirow.map(fn: (index) => index, column: "row_number")

test _map = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map})
