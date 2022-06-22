package universe_test


import "csv"
import "array"
import "testing"
import "internal/debug"

a =
    csv.from(
        csv:
            "
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

b =
    csv.from(
        csv:
            "
#datatype,string,long,dateTime:RFC3339,double,string
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,key
,,0,2021-01-01T00:00:00Z,10.0,
,,0,2021-01-01T00:01:00Z,20.0,
",
    )

// left stream's, second table is missing 'key' columns
testcase missing_column_on_left_stream {
    got =
        join(tables: {a, b}, on: ["_time"])
            |> debug.slurp()

    want =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,double,double,string,string
#group,false,false,false,false,false,true,true
#default,_result,,,,,,
,result,table,_time,_value_a,_value_b,key_a,key_b
,,0,2021-01-01T00:00:00Z,1.0,10.0,foo,
,,0,2021-01-01T00:01:00Z,2.0,20.0,foo,

#datatype,string,long,dateTime:RFC3339,double,double,string
#group,false,false,false,false,false,true
#default,_result,,,,,
,result,table,_time,_value_a,_value_b,key
,,1,2021-01-01T00:00:00Z,1.5,10.0,
,,1,2021-01-01T00:01:00Z,2.5,20.0,
",
        )

    testing.diff(got, want) |> yield()
}

a1 =
    csv.from(
        csv:
            "
#datatype,string,long,dateTime:RFC3339,double
#group,false,false,false,false
#default,_result,,,
,result,table,_time,_value
,,0,2021-01-01T00:00:00Z,1.5
,,0,2021-01-01T00:01:00Z,2.5
#datatype,string,long,dateTime:RFC3339,double,string
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,key
,,1,2021-01-01T00:00:00Z,1.0,foo
,,1,2021-01-01T00:01:00Z,2.0,foo
",
    )
b1 =
    csv.from(
        csv:
            "
#datatype,string,long,dateTime:RFC3339,double,string
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,key
,,0,2021-01-01T00:00:00Z,10.0,bar
,,0,2021-01-01T00:01:00Z,20.0,bar
",
    )

// change in the result join schema on the fly as tables in left stream contains different schema
// left stream's, second table has extra column 'key'
testcase missing_column_on_left_stream_with_join_schema_change {
    got =
        join(tables: {a1, b1}, on: ["_time"])
            |> debug.slurp()
    want =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,double,double,string
#group,false,false,false,false,false,true
#default,_result,,,,,
,result,table,_time,_value_a1,_value_b1,key
,,0,2021-01-01T00:00:00Z,1.5,10.0,bar
,,0,2021-01-01T00:01:00Z,2.5,20.0,bar
#datatype,string,long,dateTime:RFC3339,double,double,string,string
#group,false,false,false,false,false,true,true
#default,_result,,,,,,
,result,table,_time,_value_a1,_value_b1,key_a1,key_b1
,,1,2021-01-01T00:00:00Z,1.0,10.0,foo,bar
,,1,2021-01-01T00:01:00Z,2.0,20.0,foo,bar
",
        )

    testing.diff(got, want) |> yield()
}

a2 =
    csv.from(
        csv:
            "
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
b2 =
    csv.from(
        csv:
            "
#datatype,string,long,dateTime:RFC3339,double,double
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,key
,,0,2021-01-01T00:00:00Z,10.0,8.0
,,0,2021-01-01T00:01:00Z,20.0,88.0
",
    )

// when a column exists on both sides but has a different type
// column 'key' is string on the left stream and double on the right stream
testcase same_column_on_both_stream_with_different_type {
    got =
        join(tables: {a2, b2}, on: ["_time"])
            |> debug.slurp()
    want =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,double,double,string,double
#group,false,false,false,false,false,true,true
#default,_result,,,,,,
,result,table,_time,_value_a2,_value_b2,key_a2,key_b2
,,0,2021-01-01T00:00:00Z,1.0,10.0,foo,8.0
,,0,2021-01-01T00:01:00Z,2.0,20.0,foo,8.0
#datatype,string,long,dateTime:RFC3339,double,double,double
#group,false,false,false,false,false,true
#default,_result,,,,,
,result,table,_time,_value_a2,_value_b2,key
,,1,2021-01-01T00:00:00Z,1.5,10.0,8.0
,,1,2021-01-01T00:01:00Z,2.5,20.0,8.0
",
        )

    testing.diff(got, want) |> yield()
}

a3 =
    csv.from(
        csv:
            "
#datatype,string,long,dateTime:RFC3339,double,string
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,key
,,0,2021-01-01T00:00:00Z,1.0,key0
,,0,2021-01-01T00:01:00Z,1.5,key0
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,key,gkey_1
,,1,2021-01-01T00:00:00Z,2.0,key1,gkey1
,,1,2021-01-01T00:01:00Z,2.5,key1,gkey1
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,key,gkey_2
,,2,2021-01-01T00:00:00Z,3.0,key2,gkey2
,,2,2021-01-01T00:01:00Z,3.5,key2,gkey2
",
    )
b3 =
    csv.from(
        csv:
            "
#datatype,string,long,dateTime:RFC3339,double,string
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,key
,,0,2021-01-01T00:00:00Z,10.0,key0
,,0,2021-01-01T00:01:00Z,10.5,key0
",
    )

// the group key is different on left and right stream
// Left Stream -                Right Stream -
// 0th table - key              0th table - key
// 1st table - key, gkey_1
// 2nd table - key, gkey_2
// Join on _time (non groupKey)
testcase different_group_key_on_left_and_right_stream_join_on_non_group_key {
    got =
        join(tables: {a3, b3}, on: ["_time"])
            |> debug.slurp()
    want =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,double,double,string,string
#group,false,false,false,false,false,true,true
#default,_result,,,,,,
,result,table,_time,_value_a3,_value_b3,key_a3,key_b3
,,0,2021-01-01T00:00:00Z,1.0,10.0,key0,key0
,,0,2021-01-01T00:01:00Z,1.5,10.5,key0,key0
#datatype,string,long,dateTime:RFC3339,double,double,string,string,string
#group,false,false,false,false,false,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value_a3,_value_b3,key_a3,key_b3,gkey_1
,,1,2021-01-01T00:00:00Z,2.0,10.0,key1,key0,gkey1
,,1,2021-01-01T00:01:00Z,2.5,10.5,key1,key0,gkey1
#datatype,string,long,dateTime:RFC3339,double,double,string,string,string
#group,false,false,false,false,false,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value_a3,_value_b3,key_a3,key_b3,gkey_2
,,2,2021-01-01T00:00:00Z,3.0,10.0,key2,key0,gkey2
,,2,2021-01-01T00:01:00Z,3.5,10.5,key2,key0,gkey2
",
        )

    testing.diff(got, want) |> yield()
}

s1 =
    array.from(rows: [{unit: "A", power: 100}, {unit: "B", power: 200}, {unit: "C", power: 300}])
        |> group(columns: ["unit"])
        |> debug.opaque()

s2 =
    union(
        tables: [
            array.from(rows: [{columnA: "valueA", unit: "A", group: "groupX"}])
                |> group(columns: ["columnA", "unit"])
                |> debug.opaque(),
            array.from(rows: [{columnB: "valueB", unit: "B", group: "groupX"}])
                |> group(columns: ["columnB", "unit"])
                |> debug.opaque(),
            array.from(rows: [{unit: "C", group: "groupX"}])
                |> group(columns: ["unit"])
                |> debug.opaque(),
        ],
    )

ra1 = array.from(rows: [{unit: "A", power: 100, group: "groupX", columnA: "valueA"}])
ra2 = array.from(rows: [{unit: "B", power: 200, group: "groupX", columnB: "valueB"}])
ra3 = array.from(rows: [{unit: "C", power: 300, group: "groupX"}])

testcase join_different_table_schemas_in_stream {
    want =
        union(
            tables: [
                ra1 |> group(columns: ["columnA", "unit"]) |> debug.opaque(),
                ra2 |> group(columns: ["columnB", "unit"]) |> debug.opaque(),
                ra3 |> group(columns: ["unit"]) |> debug.opaque(),
            ],
        )

    got = join(tables: {s1, s2}, on: ["unit"]) |> debug.opaque()

    testing.diff(want: want, got: got) |> yield()
}
