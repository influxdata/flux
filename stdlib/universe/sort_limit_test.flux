package universe_test


import "array"
import "csv"
import "internal/debug"
import "testing"

testcase sort_limit {
    got =
        array.from(
            rows: [
                {_time: 2022-01-11T00:00:00Z, _value: 10.0},
                {_time: 2022-01-11T01:00:00Z, _value: 12.0},
                {_time: 2022-01-11T02:00:00Z, _value: 18.0},
                {_time: 2022-01-11T03:00:00Z, _value: 4.0},
                {_time: 2022-01-11T04:00:00Z, _value: 8.0},
            ],
        )
            |> sort()
            |> limit(n: 3)

    want =
        array.from(
            rows: [
                {_time: 2022-01-11T03:00:00Z, _value: 4.0},
                {_time: 2022-01-11T04:00:00Z, _value: 8.0},
                {_time: 2022-01-11T00:00:00Z, _value: 10.0},
            ],
        )

    testing.diff(got: got, want: want)
}

testcase sort_limit_divergent_schemas {
    got =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,double,long
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,t0
,,0,2022-01-11T00:00:00Z,10.0,0
,,0,2022-01-11T01:00:00Z,12.0,0

#datatype,string,long,dateTime:RFC3339,double,long
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,t1
,,1,2022-01-11T00:00:00Z,18.0,1
,,1,2022-01-11T01:00:00Z,4.0,1

#datatype,string,long,dateTime:RFC3339,double,long
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,t2
,,2,2022-01-11T00:00:00Z,8.0,2
",
        )
            |> group()
            |> sort()
            |> limit(n: 3)

    want =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,double,long,long,long
#group,false,false,false,false,false,false,false
#default,_result,,,,,,
,result,table,_time,_value,t0,t1,t2
,,0,2022-01-11T01:00:00Z,4.0,,1,
,,0,2022-01-11T00:00:00Z,8.0,,,2
,,0,2022-01-11T00:00:00Z,10.0,0,,
",
        )

    testing.diff(got: got, want: want)
}

testcase sort_limit_unordered_columns {
    got =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,double,string
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,t0
,,0,2022-01-11T00:00:00Z,10.0,a
,,0,2022-01-11T01:00:00Z,12.0,a

#datatype,string,long,dateTime:RFC3339,string,double
#group,false,false,false,true,false
#default,_result,,,,
,result,table,_time,t0,_value
,,1,2022-01-11T00:00:00Z,b,18.0
,,1,2022-01-11T01:00:00Z,b,4.0

#datatype,string,long,string,double,dateTime:RFC3339
#group,false,false,true,false,false
#default,_result,,,,
,result,table,t0,_value,_time
,,2,c,8.0,2022-01-11T00:00:00Z
",
        )
            |> group()
            |> sort()
            |> limit(n: 3)

    want =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,double,string
#group,false,false,false,false,false
#default,_result,,,,
,result,table,_time,_value,t0
,,0,2022-01-11T01:00:00Z,4.0,b
,,0,2022-01-11T00:00:00Z,8.0,c
,,0,2022-01-11T00:00:00Z,10.0,a
",
        )

    testing.diff(got: got, want: want)
}

testcase sort_limit_zero_row_table {
    input =
        array.from(rows: [{foo: "bar", _value: 10}])
            |> filter(fn: (r) => r._value > 10, onEmpty: "keep")
    want = input

    got = input |> sort() |> limit(n: 5)

    testing.diff(got, want)
}

testcase sort_limit_multi_successor {
    input =
        array.from(
            rows: [
                {_time: 2022-01-11T00:00:00Z, _value: 10.0},
                {_time: 2022-01-11T01:00:00Z, _value: 12.0},
                {_time: 2022-01-11T02:00:00Z, _value: 18.0},
                {_time: 2022-01-11T03:00:00Z, _value: 4.0},
                {_time: 2022-01-11T04:00:00Z, _value: 8.0},
            ],
        )
    in0 =
        input
            |> bottom(n: 2)
    in1 =
        input
            |> top(n: 2)
    got =
        union(tables: [in0, in1])
            |> sort(columns: ["_time"])

    want =
        array.from(
            rows: [
                {_time: 2022-01-11T01:00:00Z, _value: 12.0},
                {_time: 2022-01-11T02:00:00Z, _value: 18.0},
                {_time: 2022-01-11T03:00:00Z, _value: 4.0},
                {_time: 2022-01-11T04:00:00Z, _value: 8.0},
            ],
        )

    testing.diff(got: got, want: want)
}

testcase sort_limit_empty_chunk {
    got =
        array.from(
            rows: [
                {_time: 2022-01-11T00:00:00Z, t0: "aa", _value: 12.0},
                {_time: 2022-01-11T00:00:00Z, t0: "ab", _value: 10.0},
                {_time: 2022-01-11T00:00:00Z, t0: "ba", _value: 18.0},
                {_time: 2022-01-11T00:00:00Z, t0: "bb", _value: 4.0},
            ],
        )
            |> group(columns: ["t0"])
            |> filter(fn: (r) => r.t0 =~ /b/, onEmpty: "keep")
            |> group()
            |> top(n: 2)

    want =
        array.from(
            rows: [
                {_time: 2022-01-11T00:00:00Z, t0: "ba", _value: 18.0},
                {_time: 2022-01-11T00:00:00Z, t0: "ab", _value: 10.0},
            ],
        )

    testing.diff(got: got, want: want)
}
