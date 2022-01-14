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
#datatype,string,long,string,string,dateTime:RFC3339,string,string,long,long
#group,false,false,true,false,false,false,false,false,false
#default,_result,0,,m0,2000-01-01T00:00:00Z,,,,
,result,table,_field,_measurement,_time,concat,_value,intcol3,val
,,,newGroup,,,test10,test10,1,99
,,,newGroup,,,test10->test11,test11,,97
,,,newGroup,,,test12,test12,3,99
,,,newGroup,,,test12->test13,test13,4,97
"

t_map = (table=<-) =>
    table
        |> drop(columns: ["_start", "_stop"])
        |> group(columns: ["_field"])
        |> multirow.map(
            fn: (previous, row) => {
                x = previous.x_col * 2 - 1

                return {row with _field: "newGroup",
                    concat: (if exists previous.concat then previous.concat + "->" else "") + row._value,
                    x_col: x,
                    val: x % 100,
                }
            },
            init: {x_col: 100},
            virtual: ["x_col"],
        )

test _map = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map})
