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
#datatype,string,long,string,double
#group,false,false,true,false
#default,_result,0,,
,result,table,_field,_value
,,0,test1,0.6666666666666666
,,0,test1,1.3333333333333333
,,0,test1,2
,,0,test1,3.3333333333333335
,,0,test1,4
,,1,test2,10.5
,,1,test2,14
,,1,test2,17
"

t_map = (table=<-) =>
    table
        |> drop(columns: ["_start", "_stop"])
        |> multirow.map(left: 1, right: 1, limit: -5, fn: (window) => window |> mean())

test _map = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map})
