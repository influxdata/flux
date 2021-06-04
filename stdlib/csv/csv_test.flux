package csv_test


import "array"
import "strings"
import "testing"
import "csv"
import "math"

testcase from_raw {
    input = "
time,float,int,uint,bool,string
2021-03-12T13:58:59Z,42.69,-67,152,false,hello world
2021-03-12T13:58:59Z,42.69,-67,152,false,hello world
2021-03-12T13:58:59Z,42.69,-67,152,false,hello world
2021-03-12T13:58:59Z,42.69,-67,152,false,hello world
"
    want = array.from(
        rows: [
            {time: "2021-03-12T13:58:59Z", float: "42.69", int: "-67", uint: "152", bool: "false", string: "hello world"},
            {time: "2021-03-12T13:58:59Z", float: "42.69", int: "-67", uint: "152", bool: "false", string: "hello world"},
            {time: "2021-03-12T13:58:59Z", float: "42.69", int: "-67", uint: "152", bool: "false", string: "hello world"},
            {time: "2021-03-12T13:58:59Z", float: "42.69", int: "-67", uint: "152", bool: "false", string: "hello world"},
        ],
    )

    // Using raw mode so all columns are of type string
    result = csv.from(csv: input, mode: "raw")

    testing.diff(got: result, want: want)
}
testcase from_annotations {
    input = "
#datatype,string,long,dateTime:RFC3339,double,long,unsignedLong,boolean,string
#group,false,false,false,false,false,false,false,false
#default,_result,,,,,,,
,result,table,time,float,int,uint,bool,string
,,0,2021-03-12T13:58:59Z,42.69,-67,152,false,hello world
,,0,2021-03-12T13:58:59Z,42.69,-67,152,false,hello world
,,0,2021-03-12T13:58:59Z,+Inf,-67,152,false,hello world
,,0,2021-03-12T13:58:59Z,-Inf,-67,152,false,hello world
,,0,2021-03-12T13:58:59Z,NaN,-67,152,false,hello world
"
    want = array.from(
        rows: [
            {time: 2021-03-12T13:58:59Z, float: 42.69, int: -67, uint: uint(v: 152), bool: false, string: "hello world"},
            {time: 2021-03-12T13:58:59Z, float: 42.69, int: -67, uint: uint(v: 152), bool: false, string: "hello world"},
            {time: 2021-03-12T13:58:59Z, float: math.mInf(sign: 1), int: -67, uint: uint(v: 152), bool: false, string: "hello world"},
            {time: 2021-03-12T13:58:59Z, float: math.mInf(sign: -1), int: -67, uint: uint(v: 152), bool: false, string: "hello world"},
            {time: 2021-03-12T13:58:59Z, float: math.NaN(), int: -67, uint: uint(v: 152), bool: false, string: "hello world"},
        ],
    )

    // Using annotations mode so all columns can be specific types
    result = csv.from(csv: input, mode: "annotations")

    testing.diff(got: result, want: want, nansEqual: true)
}
testcase from_multiple_tables {
    input = "
#datatype,string,long,dateTime:RFC3339,double,long,unsignedLong,boolean,string
#group,false,false,false,false,false,false,false,true
#default,_result,,,,,,,
,result,table,time,float,int,uint,bool,string
,,0,2021-03-12T13:58:59Z,42.69,-67,152,false,hello world
,,0,2021-03-12T13:58:59Z,42.69,-67,152,false,hello world
,,0,2021-03-12T13:58:59Z,42.69,-67,152,false,hello world
,,0,2021-03-12T13:58:59Z,42.69,-67,152,false,hello world

#datatype,string,long,dateTime:RFC3339,double,long,unsignedLong,boolean,string
#group,false,false,false,false,false,false,false,true
#default,_result,,,,,,,
,result,table,time,float,int,uint,bool,string
,,1,2021-03-12T13:58:59Z,42.69,-67,152,false,bye world
,,1,2021-03-12T13:58:59Z,42.69,-67,152,false,bye world
,,1,2021-03-12T13:58:59Z,42.69,-67,152,false,bye world
,,1,2021-03-12T13:58:59Z,42.69,-67,152,false,bye world
"
    want = array.from(
        rows: [
            {time: 2021-03-12T13:58:59Z, float: 42.69, int: -67, uint: uint(v: 152), bool: false, string: "hello world"},
            {time: 2021-03-12T13:58:59Z, float: 42.69, int: -67, uint: uint(v: 152), bool: false, string: "hello world"},
            {time: 2021-03-12T13:58:59Z, float: 42.69, int: -67, uint: uint(v: 152), bool: false, string: "hello world"},
            {time: 2021-03-12T13:58:59Z, float: 42.69, int: -67, uint: uint(v: 152), bool: false, string: "hello world"},
            {time: 2021-03-12T13:58:59Z, float: 42.69, int: -67, uint: uint(v: 152), bool: false, string: "bye world"},
            {time: 2021-03-12T13:58:59Z, float: 42.69, int: -67, uint: uint(v: 152), bool: false, string: "bye world"},
            {time: 2021-03-12T13:58:59Z, float: 42.69, int: -67, uint: uint(v: 152), bool: false, string: "bye world"},
            {time: 2021-03-12T13:58:59Z, float: 42.69, int: -67, uint: uint(v: 152), bool: false, string: "bye world"},
        ],
    )
        |> group(columns: ["string"])

    // Using annotations mode so all columns can be specific types
    result = csv.from(csv: input, mode: "annotations")

    testing.diff(got: result, want: want)
}
testcase from_large {
    input = "
#datatype,string,long,string,double
#group,false,false,true,false
#default,_result,,,
,result,table,tag,_value
${strings.repeat(v: ",,0,a,5\n", i: 1111)}

#datatype,string,long,string,double
#group,false,false,true,false
#default,_result,,,
,result,table,tag,_value
${strings.repeat(v: ",,0,b,5\n", i: 1100)}
"
    want = array.from(
        rows: [
            {tag: "a", _value: 1111},
            {tag: "b", _value: 1100},
        ],
    )

    // Using annotations mode so all columns can be specific types
    result = csv.from(csv: input, mode: "annotations")
        |> count()
        |> group()
        |> sort(columns: ["string"])

    testing.diff(got: result, want: want)
}
