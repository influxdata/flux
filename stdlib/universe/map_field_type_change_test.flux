package universe_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,m,f,2018-01-01T00:00:00Z,2
"

outData = "
#datatype,string,long,string,string,dateTime:RFC3339,string
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,m,f,2018-01-01T00:00:00Z,hello
"

t_map_field_type_change = (table=<-) =>
    (table
        |> range(start: 2018-01-01T00:00:00Z)
        |> drop(columns: ["_start", "_stop"])
        |> map(fn: (r) => ({r with _value: 2.0})) // establish _value as a double column in output
        |> map(fn: (r) => ({r with _value: "hello"})) // convert to a string
        |> filter(fn: (r) => r._value == "hello") // previously this would produce an error
    )

test _map_field_type_change = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map_field_type_change})
