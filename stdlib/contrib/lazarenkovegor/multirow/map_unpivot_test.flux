package multirow_test


import "testing"
import "contrib/lazarenkovegor/multirow"

inData =
    "
#datatype,string,long,string,string,long,dateTime:RFC3339,string
#group,false,false,false,false,false,false,false
#default,_result,0,,,,2000-01-01T00:00:00Z,m0
,result,table,_field,_value,intcol3,_time,_measurement
,,,test1,test10,1,,
,,,test1,test11,,,
,,,test2,test12,3,,
,,,test2,test13,4,,
"

outData =
    "
#datatype,string,long,string,string,dateTime:RFC3339,string,long
#group,false,false,false,false,false,false,false
#default,_result,0,,m0,2000-01-01T00:00:00Z,,
,result,table,_field,_measurement,_time,_value,group_id
,,,test1,,,test10,0
,,,intcol3,,,1,0
,,,test1,,,test11,1
,,,intcol3,,,,1
,,,test2,,,test12,2
,,,intcol3,,,3,2
,,,test2,,,test13,3
,,,intcol3,,,4,3
"

t_map = (table=<-) =>
    table
        |> drop(columns: ["_start", "_stop"])
        |> multirow.map(
            fn: (row, index) =>
                [
                    {
                        _time: row._time,
                        _measurement: row._measurement,
                        _field: row._field,
                        _value: row._value,
                        group_id: index,
                    },
                    {
                        _time: row._time,
                        _measurement: row._measurement,
                        _field: "intcol3",
                        _value: string(v: row.intcol3),
                        group_id: index,
                    },
                ],
        )

test _map = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map})
