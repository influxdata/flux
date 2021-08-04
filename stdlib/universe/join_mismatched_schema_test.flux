package universe_test


import "csv"
import "testing"
import "internal/debug"

a = csv.from(
    csv: "
#datatype,string,long,dateTime:RFC3339,double,string
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,key
,,0,2021-01-01T00:00:00Z,1.0,foo
,,0,2021-01-01T00:01:00Z,2.0,foo

#datatype,string,long,dateTime:RFC3339,double
#group,false,false,false,false
#default,_result,,,
,result,table,_time,_value
,,1,2021-01-01T00:00:00Z,1.5
,,1,2021-01-01T00:01:00Z,2.5
",
)

b = csv.from(
    csv: "
#datatype,string,long,dateTime:RFC3339,double,string
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,key
,,0,2021-01-01T00:00:00Z,10.0,
,,0,2021-01-01T00:01:00Z,20.0,
",
)

testcase normal {
    got = join(tables: {a, b}, on: ["_time"])
        |> debug.slurp()

    want = csv.from(
        csv: "
#datatype,string,long,dateTime:RFC3339,double,double,string,string
#group,false,false,false,false,false,true,true
#default,_result,,,,,,
,result,table,_time,_value_a,_value_b,key_a,key_b
,,0,2021-01-01T00:00:00Z,1.0,10.0,foo,
,,0,2021-01-01T00:01:00Z,2.0,20.0,foo,

#datatype,string,long,dateTime:RFC3339,double,double,string,string
#group,false,false,false,false,false,false,true
#default,_result,,,,,,
,result,table,_time,_value_a,_value_b,key_a,key_b
,,1,2021-01-01T00:00:00Z,1.5,10.0,,
,,1,2021-01-01T00:01:00Z,2.5,20.0,,
",
    )

    testing.diff(got, want) |> yield()
}
