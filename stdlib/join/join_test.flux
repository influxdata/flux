package join_test


import "join"
import "array"
import "csv"
import "testing"

right =
    array.from(
        rows: [
            {_time: 2022-06-01T00:00:00Z, _value: 1, id: "a", key: 1},
            {_time: 2022-06-01T00:00:01Z, _value: 2, id: "a", key: 1},
            {_time: 2022-06-01T00:00:02Z, _value: 3, id: "a", key: 1},
            {_time: 2022-06-01T00:00:03Z, _value: 4, id: "a", key: 1},
            {_time: 2022-06-01T00:00:00Z, _value: 5, id: "b", key: 1},
            {_time: 2022-06-01T00:00:01Z, _value: 6, id: "b", key: 1},
            {_time: 2022-06-01T00:00:02Z, _value: 7, id: "b", key: 1},
            {_time: 2022-06-01T00:00:03Z, _value: 8, id: "b", key: 1},
            {_time: 2022-06-01T00:00:00Z, _value: 9, id: "a", key: 2},
            {_time: 2022-06-01T00:00:01Z, _value: 10, id: "a", key: 2},
            {_time: 2022-06-01T00:00:02Z, _value: 11, id: "a", key: 2},
            {_time: 2022-06-01T00:00:03Z, _value: 12, id: "a", key: 2},
            {_time: 2022-06-01T00:00:00Z, _value: 13, id: "b", key: 2},
            {_time: 2022-06-01T00:00:01Z, _value: 14, id: "b", key: 2},
            {_time: 2022-06-01T00:00:02Z, _value: 15, id: "b", key: 2},
            {_time: 2022-06-01T00:00:03Z, _value: 16, id: "b", key: 2},
        ],
    )
        |> group(columns: ["key"])

left =
    array.from(
        rows: [
            {_time: 2022-06-01T00:00:00Z, _value: 12.34, label: "a", key: 1},
            {_time: 2022-06-01T00:00:01Z, _value: 73.01, label: "a", key: 1},
            {_time: 2022-06-01T00:00:02Z, _value: 56.85, label: "a", key: 1},
            {_time: 2022-06-01T00:00:03Z, _value: 21.28, label: "a", key: 1},
            {_time: 2022-06-01T00:00:00Z, _value: 12.34, label: "c", key: 1},
            {_time: 2022-06-01T00:00:01Z, _value: 73.01, label: "c", key: 1},
            {_time: 2022-06-01T00:00:02Z, _value: 56.85, label: "c", key: 1},
            {_time: 2022-06-01T00:00:03Z, _value: 21.28, label: "c", key: 1},
            {_time: 2022-06-01T00:00:00Z, _value: 12.34, label: "a", key: 2},
            {_time: 2022-06-01T00:00:01Z, _value: 73.01, label: "a", key: 2},
            {_time: 2022-06-01T00:00:02Z, _value: 56.85, label: "a", key: 2},
            {_time: 2022-06-01T00:00:03Z, _value: 21.28, label: "a", key: 2},
            {_time: 2022-06-01T00:00:00Z, _value: 12.34, label: "c", key: 2},
            {_time: 2022-06-01T00:00:01Z, _value: 73.01, label: "c", key: 2},
            {_time: 2022-06-01T00:00:02Z, _value: 56.85, label: "c", key: 2},
            {_time: 2022-06-01T00:00:03Z, _value: 21.28, label: "c", key: 2},
        ],
    )
        |> group(columns: ["key"])

testcase inner_join {
    want =
        array.from(
            rows: [
                {
                    _time: 2022-06-01T00:00:00Z,
                    label: "a",
                    intv: 1,
                    floatv: 12.34,
                    key: 1,
                },
                {
                    _time: 2022-06-01T00:00:01Z,
                    label: "a",
                    intv: 2,
                    floatv: 73.01,
                    key: 1,
                },
                {
                    _time: 2022-06-01T00:00:02Z,
                    label: "a",
                    intv: 3,
                    floatv: 56.85,
                    key: 1,
                },
                {
                    _time: 2022-06-01T00:00:03Z,
                    label: "a",
                    intv: 4,
                    floatv: 21.28,
                    key: 1,
                },
                {
                    _time: 2022-06-01T00:00:00Z,
                    label: "a",
                    intv: 9,
                    floatv: 12.34,
                    key: 2,
                },
                {
                    _time: 2022-06-01T00:00:01Z,
                    label: "a",
                    intv: 10,
                    floatv: 73.01,
                    key: 2,
                },
                {
                    _time: 2022-06-01T00:00:02Z,
                    label: "a",
                    intv: 11,
                    floatv: 56.85,
                    key: 2,
                },
                {
                    _time: 2022-06-01T00:00:03Z,
                    label: "a",
                    intv: 12,
                    floatv: 21.28,
                    key: 2,
                },
            ],
        )
            |> group(columns: ["key"])

    got =
        join.tables(
            left: left,
            right: right,
            on: (l, r) => l.label == r.id and l._time == r._time,
            as: (l, r) =>
                ({
                    label: l.label,
                    intv: r._value,
                    floatv: l._value,
                    _time: l._time,
                    key: l.key,
                }),
            method: "inner",
        )

    testing.diff(want: want, got: got)
}

testcase full_outer_join {
    wantData =
        "
#datatype,string,long,string,long,double,dateTime:RFC3339,long
#group,false,false,false,false,false,false,false
#default,_result,,,,,,
,result,table,label,intv,floatv,_time,key
,,0,a,1,12.34,2022-06-01T00:00:00Z,1
,,0,a,2,73.01,2022-06-01T00:00:01Z,1
,,0,a,3,56.85,2022-06-01T00:00:02Z,1
,,0,a,4,21.28,2022-06-01T00:00:03Z,1
,,0,a,9,12.34,2022-06-01T00:00:00Z,2
,,0,a,10,73.01,2022-06-01T00:00:01Z,2
,,0,a,11,56.85,2022-06-01T00:00:02Z,2
,,0,a,12,21.28,2022-06-01T00:00:03Z,2
,,0,b,5,,2022-06-01T00:00:00Z,1
,,0,b,6,,2022-06-01T00:00:01Z,1
,,0,b,7,,2022-06-01T00:00:02Z,1
,,0,b,8,,2022-06-01T00:00:03Z,1
,,0,b,13,,2022-06-01T00:00:00Z,2
,,0,b,14,,2022-06-01T00:00:01Z,2
,,0,b,15,,2022-06-01T00:00:02Z,2
,,0,b,16,,2022-06-01T00:00:03Z,2
,,0,c,,12.34,2022-06-01T00:00:00Z,1
,,0,c,,73.01,2022-06-01T00:00:01Z,1
,,0,c,,56.85,2022-06-01T00:00:02Z,1
,,0,c,,21.28,2022-06-01T00:00:03Z,1
,,0,c,,12.34,2022-06-01T00:00:00Z,2
,,0,c,,73.01,2022-06-01T00:00:01Z,2
,,0,c,,56.85,2022-06-01T00:00:02Z,2
,,0,c,,21.28,2022-06-01T00:00:03Z,2
"
    want = csv.from(csv: wantData) |> group(columns: ["key"])

    got =
        join.tables(
            left: left,
            right: right,
            on: (l, r) => l.label == r.id and l._time == r._time,
            as: (l, r) => {
                label = if exists l.label then l.label else r.id
                time = if exists l._time then l._time else r._time

                return {
                    label: label,
                    intv: r._value,
                    floatv: l._value,
                    _time: time,
                    key: r.key,
                }
            },
            method: "full",
        )

    testing.diff(want: want, got: got)
}

testcase left_outer_join {
    wantData =
        "
#datatype,string,long,string,long,double,dateTime:RFC3339,long
#group,false,false,false,false,false,false,false
#default,_result,,,,,,
,result,table,label,intv,floatv,_time,key
,,0,a,1,12.34,2022-06-01T00:00:00Z,1
,,0,a,2,73.01,2022-06-01T00:00:01Z,1
,,0,a,3,56.85,2022-06-01T00:00:02Z,1
,,0,a,4,21.28,2022-06-01T00:00:03Z,1
,,0,a,9,12.34,2022-06-01T00:00:00Z,2
,,0,a,10,73.01,2022-06-01T00:00:01Z,2
,,0,a,11,56.85,2022-06-01T00:00:02Z,2
,,0,a,12,21.28,2022-06-01T00:00:03Z,2
,,0,c,,12.34,2022-06-01T00:00:00Z,1
,,0,c,,73.01,2022-06-01T00:00:01Z,1
,,0,c,,56.85,2022-06-01T00:00:02Z,1
,,0,c,,21.28,2022-06-01T00:00:03Z,1
,,0,c,,12.34,2022-06-01T00:00:00Z,2
,,0,c,,73.01,2022-06-01T00:00:01Z,2
,,0,c,,56.85,2022-06-01T00:00:02Z,2
,,0,c,,21.28,2022-06-01T00:00:03Z,2
"
    want = csv.from(csv: wantData) |> group(columns: ["key"])

    got =
        join.tables(
            left: left,
            right: right,
            on: (l, r) => l.label == r.id and l._time == r._time,
            as: (l, r) =>
                ({
                    label: l.label,
                    intv: r._value,
                    floatv: l._value,
                    _time: l._time,
                    key: r.key,
                }),
            method: "left",
        )

    testing.diff(want: want, got: got)
}

testcase right_outer_join {
    wantData =
        "
#datatype,string,long,string,long,double,dateTime:RFC3339,long
#group,false,false,false,false,false,false,false
#default,_result,,,,,,
,result,table,label,intv,floatv,_time,key
,,0,a,1,12.34,2022-06-01T00:00:00Z,1
,,0,a,2,73.01,2022-06-01T00:00:01Z,1
,,0,a,3,56.85,2022-06-01T00:00:02Z,1
,,0,a,4,21.28,2022-06-01T00:00:03Z,1
,,0,a,9,12.34,2022-06-01T00:00:00Z,2
,,0,a,10,73.01,2022-06-01T00:00:01Z,2
,,0,a,11,56.85,2022-06-01T00:00:02Z,2
,,0,a,12,21.28,2022-06-01T00:00:03Z,2
,,0,b,5,,2022-06-01T00:00:00Z,1
,,0,b,6,,2022-06-01T00:00:01Z,1
,,0,b,7,,2022-06-01T00:00:02Z,1
,,0,b,8,,2022-06-01T00:00:03Z,1
,,0,b,13,,2022-06-01T00:00:00Z,2
,,0,b,14,,2022-06-01T00:00:01Z,2
,,0,b,15,,2022-06-01T00:00:02Z,2
,,0,b,16,,2022-06-01T00:00:03Z,2
"
    want = csv.from(csv: wantData) |> group(columns: ["key"])

    got =
        join.tables(
            left: left,
            right: right,
            on: (l, r) => l.label == r.id and l._time == r._time,
            as: (l, r) => {
                return {
                    label: r.id,
                    intv: r._value,
                    floatv: l._value,
                    _time: r._time,
                    key: l.key,
                }
            },
            method: "right",
        )

    testing.diff(want: want, got: got)
}

testcase exclusive_group_keys1 {
    tbl1 =
        array.from(rows: [{_time: 2022-06-01T00:00:00Z, _value: 1, label: "a", key: 1}])
            |> group(columns: ["key"])
    tbl2 =
        array.from(rows: [{_time: 2022-06-01T00:00:00Z, _value: 0.1, id: "a", key: 2}])
            |> group(columns: ["key"])

    wantOutput =
        "
#datatype,string,long,string,long,double,dateTime:RFC3339,long
#group,false,false,false,false,false,false,false
#default,_result,,,,,,
,result,table,label,v_left,v_right,_time,key
,,0,a,1,,2022-06-01T00:00:00Z,1
,,0,a,,0.1,2022-06-01T00:00:00Z,2
"
    want = csv.from(csv: wantOutput) |> group(columns: ["key"])

    got =
        join.tables(
            method: "full",
            left: tbl1,
            right: tbl2,
            on: (l, r) => l.label == r.id and r._time == l._time,
            as: (l, r) => {
                time = if exists l._time then l._time else r._time
                label = if exists l.label then l.label else r.id

                return {
                    _time: time,
                    label: label,
                    v_left: l._value,
                    v_right: r._value,
                    key: r.key,
                }
            },
        )

    testing.diff(want: want, got: got)
}

testcase exclusive_group_keys2 {
    tbl1 =
        array.from(
            rows: [
                {
                    _time: 2022-06-01T00:00:00Z,
                    _value: 1,
                    label: "a",
                    key: 1,
                    group: "one",
                },
            ],
        )
            |> group(columns: ["key", "group"])
    tbl2 =
        array.from(rows: [{_time: 2022-06-01T00:00:00Z, _value: 0.1, id: "a", key: 2}])
            |> group(columns: ["key"])

    wantOutput =
        "
#datatype,string,long,string,long,double,dateTime:RFC3339,long,string
#group,false,false,false,false,false,false,true,true
#default,_result,,,,,,,
,result,table,label,v_left,v_right,_time,key,group
,,0,a,1,,2022-06-01T00:00:00Z,1,one

#datatype,string,long,string,long,double,dateTime:RFC3339,long,string
#group,false,false,false,false,false,false,true,false
#default,_result,,,,,,,
,result,table,label,v_left,v_right,_time,key,group
,,0,a,,0.1,2022-06-01T00:00:00Z,2,
"
    want = csv.from(csv: wantOutput)

    got =
        join.tables(
            method: "full",
            left: tbl1,
            right: tbl2,
            on: (l, r) => l.label == r.id and r._time == l._time,
            as: (l, r) => {
                time = if exists l._time then l._time else r._time
                label = if exists l.label then l.label else r.id

                return {
                    _time: time,
                    label: label,
                    v_left: l._value,
                    v_right: r._value,
                    key: r.key,
                    group: l.group,
                }
            },
        )

    testing.diff(want: want, got: got)
}
