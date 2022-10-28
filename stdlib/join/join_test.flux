package join_test


import "join"
import "array"
import "csv"
import "testing"
import "internal/debug"

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

testcase multi_join {
    intermediate =
        join.tables(
            left: left,
            right: right,
            on: (l, r) => l.label == r.id and l._time == r._time,
            as: (l, r) => {
                return {
                    label: l.label,
                    _time: r._time,
                    key: l.key,
                    _value: l._value + float(v: r._value),
                }
            },
            method: "inner",
        )
    got =
        join.tables(
            left: left,
            right: intermediate,
            on: (l, r) => l.label == r.label and l._time == r._time,
            as: (l, r) => {
                return {label: l.label, _time: r._time, key: l.key, _value: r._value - l._value}
            },
            method: "inner",
        )
    want =
        right
            |> filter(fn: (r) => r.id == "a")
            |> map(
                fn: (r) => ({key: r.key, _time: r._time, _value: float(v: r._value), label: r.id}),
            )

    testing.diff(want: want, got: got)
}

testcase join_empty_table {
    // TODO Enable/fix in https://github.com/influxdata/flux/issues/5307
    option testing.tags = ["skip"]

    something = array.from(rows: [{_value: 1, id: "a"}])

    nothing =
        array.from(rows: [{_value: 0.6, id: "b"}])
            |> filter(fn: (r) => r.id == "empty table")

    fn = () =>
        join.tables(
            method: "full",
            left: something,
            right: nothing,
            on: (l, r) => l.id == r.id,
            as: (l, r) => {
                id = if exists l.id then l.id else r.id

                return {id: id, v_left: l._value, v_right: r._value}
            },
        )
    want = /error preparing right sight of join: cannot join on empty table/

    testing.shouldError(fn, want)
}

testcase large_input_in_join {
    rTbl =
        array.from(
            rows: [
                {_time: 2022-09-28T12:24:30Z, bool_value: true},
                {_time: 2022-09-28T17:00:00Z, bool_value: false},
                {_time: 2022-09-29T07:00:00Z, bool_value: true},
                {_time: 2022-09-29T17:00:00Z, bool_value: false},
                {_time: 2022-09-30T08:00:00Z, bool_value: true},
                {_time: 2022-09-30T16:00:00Z, bool_value: false},
                {_time: 2022-10-03T07:00:00Z, bool_value: true},
                {_time: 2022-10-03T17:00:00Z, bool_value: false},
                {_time: 2022-10-04T07:00:00Z, bool_value: true},
                {_time: 2022-10-04T17:00:00Z, bool_value: false},
                {_time: 2022-10-05T07:00:00Z, bool_value: true},
                {_time: 2022-10-05T12:24:30Z, bool_value: true},
                {_time: 2022-10-05T12:24:30Z, bool_value: true},
            ],
        )
            |> debug.opaque()
    lTbl =
        array.from(
            rows: [
                {_time: 2022-10-05T10:26:38Z, _value: "d", id: "id1"},
                {_time: 2022-09-28T13:57:12Z, _value: "d", id: "id2"},
                {_time: 2022-09-28T13:57:24Z, _value: "d", id: "id3"},
                {_time: 2022-09-28T13:57:27Z, _value: "d", id: "id4"},
                {_time: 2022-09-28T13:57:39Z, _value: "d", id: "id5"},
                {_time: 2022-09-28T13:57:41Z, _value: "d", id: "id6"},
                {_time: 2022-09-28T13:57:44Z, _value: "d", id: "id7"},
                {_time: 2022-09-28T13:57:47Z, _value: "d", id: "id8"},
                {_time: 2022-09-29T06:35:08Z, _value: "d", id: "id9"},
                {_time: 2022-09-29T06:35:26Z, _value: "ip", id: "id10"},
                {_time: 2022-09-28T13:01:45Z, _value: "pp", id: "id11"},
                {_time: 2022-09-28T15:42:52Z, _value: "a", id: "id12"},
                {_time: 2022-09-29T10:47:35Z, _value: "d", id: "id13"},
                {_time: 2022-09-30T10:54:23Z, _value: "d", id: "id14"},
                {_time: 2022-10-04T15:43:50Z, _value: "d", id: "id15"},
                {_time: 2022-10-05T08:43:58Z, _value: "d", id: "id16"},
                {_time: 2022-10-05T08:44:07Z, _value: "d", id: "id17"},
                {_time: 2022-10-05T08:44:29Z, _value: "d", id: "id18"},
                {_time: 2022-10-05T08:44:20Z, _value: "d", id: "id19"},
                {_time: 2022-09-29T13:53:56Z, _value: "d", id: "id20"},
                {_time: 2022-10-03T15:12:30Z, _value: "d", id: "id21"},
                {_time: 2022-10-05T08:44:36Z, _value: "d", id: "id22"},
                {_time: 2022-09-28T15:57:40Z, _value: "d", id: "id23"},
                {_time: 2022-09-29T14:24:25Z, _value: "d", id: "id24"},
                {_time: 2022-09-29T14:27:06Z, _value: "d", id: "id25"},
                {_time: 2022-10-04T09:56:55Z, _value: "d", id: "id26"},
                {_time: 2022-10-05T08:44:44Z, _value: "d", id: "id27"},
                {_time: 2022-10-05T09:02:50Z, _value: "d", id: "id28"},
                {_time: 2022-09-29T10:47:23Z, _value: "r", id: "id29"},
                {_time: 2022-10-05T09:03:32Z, _value: "d", id: "id30"},
                {_time: 2022-09-29T12:41:00Z, _value: "d", id: "id31"},
                {_time: 2022-10-04T14:43:00Z, _value: "d", id: "id32"},
                {_time: 2022-09-29T12:41:14Z, _value: "f", id: "id33"},
                {_time: 2022-10-05T08:45:15Z, _value: "d", id: "id34"},
                {_time: 2022-09-29T15:53:57Z, _value: "r", id: "id35"},
                {_time: 2022-09-29T15:54:14Z, _value: "r", id: "id36"},
                {_time: 2022-10-05T09:14:18Z, _value: "d", id: "id37"},
                {_time: 2022-09-30T11:54:37Z, _value: "r", id: "id38"},
                {_time: 2022-10-05T09:03:52Z, _value: "d", id: "id39"},
                {_time: 2022-10-05T09:02:30Z, _value: "d", id: "id40"},
                {_time: 2022-10-05T09:08:21Z, _value: "d", id: "id41"},
                {_time: 2022-10-03T07:31:56Z, _value: "d", id: "id42"},
                {_time: 2022-10-03T15:06:16Z, _value: "d", id: "id43"},
                {_time: 2022-10-03T07:54:57Z, _value: "r", id: "id44"},
                {_time: 2022-10-04T08:30:25Z, _value: "d", id: "id45"},
                {_time: 2022-10-05T08:45:44Z, _value: "d", id: "id46"},
                {_time: 2022-10-05T08:45:58Z, _value: "d", id: "id47"},
                {_time: 2022-10-05T08:46:06Z, _value: "d", id: "id48"},
                {_time: 2022-10-03T13:30:01Z, _value: "b", id: "id49"},
                {_time: 2022-10-03T14:18:00Z, _value: "b", id: "id50"},
                {_time: 2022-10-03T14:18:07Z, _value: "b", id: "id51"},
                {_time: 2022-10-03T14:18:58Z, _value: "d", id: "id52"},
                {_time: 2022-10-03T14:21:45Z, _value: "f", id: "id53"},
                {_time: 2022-10-03T14:33:13Z, _value: "d", id: "id54"},
                {_time: 2022-10-03T14:33:34Z, _value: "f", id: "id55"},
                {_time: 2022-10-04T07:32:29Z, _value: "d", id: "id56"},
                {_time: 2022-10-04T07:32:52Z, _value: "f", id: "id57"},
                {_time: 2022-10-04T09:06:14Z, _value: "r", id: "id58"},
                {_time: 2022-10-04T11:49:50Z, _value: "f", id: "id59"},
                {_time: 2022-10-04T12:12:05Z, _value: "d", id: "id60"},
                {_time: 2022-10-04T14:38:21Z, _value: "r", id: "id61"},
                {_time: 2022-10-04T14:52:05Z, _value: "r", id: "id62"},
                {_time: 2022-10-04T15:11:23Z, _value: "f", id: "id63"},
                {_time: 2022-10-04T15:14:30Z, _value: "d", id: "id64"},
                {_time: 2022-10-05T12:12:11Z, _value: "r", id: "id65"},
                {_time: 2022-10-05T12:20:34Z, _value: "av", id: "id66"},
                {_time: 2022-10-05T12:21:09Z, _value: "d", id: "id67"},
                {_time: 2022-10-05T12:20:30Z, _value: "f", id: "id68"},
                {_time: 2022-09-28T14:18:25Z, _value: "d", id: "id69"},
                {_time: 2022-09-28T14:18:35Z, _value: "d", id: "id70"},
                {_time: 2022-09-28T14:18:37Z, _value: "d", id: "id71"},
                {_time: 2022-10-05T09:15:09Z, _value: "pd", id: "id72"},
                {_time: 2022-10-05T07:06:07Z, _value: "pd", id: "id73"},
                {_time: 2022-10-03T08:08:53Z, _value: "rec", id: "id74"},
                {_time: 2022-10-03T08:08:38Z, _value: "rec", id: "id75"},
                {_time: 2022-10-05T07:09:44Z, _value: "pd", id: "id76"},
                {_time: 2022-10-05T07:06:08Z, _value: "pd", id: "id77"},
                {_time: 2022-10-05T09:14:32Z, _value: "pd", id: "id78"},
                {_time: 2022-09-29T13:12:37Z, _value: "pd", id: "id79"},
                {_time: 2022-09-29T15:54:15Z, _value: "o", id: "id80"},
                {_time: 2022-09-30T14:43:29Z, _value: "c", id: "id81"},
                {_time: 2022-09-29T13:54:20Z, _value: "fp", id: "id82"},
                {_time: 2022-09-29T15:53:59Z, _value: "o", id: "id83"},
            ],
        )
            |> debug.opaque()

    mappedRight =
        rTbl
            |> set(key: "on", value: "x")

    mappedLeft =
        lTbl
            |> set(key: "on", value: "x")
    want =
        array.from(
            rows: [
                {_time: 2022-09-28T12:24:30Z, bool_value: true, id: "id83", on: "x"},
                {_time: 2022-09-28T17:00:00Z, bool_value: false, id: "id83", on: "x"},
                {_time: 2022-09-29T07:00:00Z, bool_value: true, id: "id83", on: "x"},
                {_time: 2022-09-29T17:00:00Z, bool_value: false, id: "id83", on: "x"},
                {_time: 2022-09-30T08:00:00Z, bool_value: true, id: "id83", on: "x"},
                {_time: 2022-09-30T16:00:00Z, bool_value: false, id: "id83", on: "x"},
                {_time: 2022-10-03T07:00:00Z, bool_value: true, id: "id83", on: "x"},
                {_time: 2022-10-03T17:00:00Z, bool_value: false, id: "id83", on: "x"},
                {_time: 2022-10-04T07:00:00Z, bool_value: true, id: "id83", on: "x"},
                {_time: 2022-10-04T17:00:00Z, bool_value: false, id: "id83", on: "x"},
                {_time: 2022-10-05T07:00:00Z, bool_value: true, id: "id83", on: "x"},
                {_time: 2022-10-05T12:24:30Z, bool_value: true, id: "id83", on: "x"},
                {_time: 2022-10-05T12:24:30Z, bool_value: true, id: "id83", on: "x"},
            ],
        )

    got =
        join.inner(
            left: mappedLeft,
            right: mappedRight,
            on: (l, r) => l.on == r.on,
            as: (l, r) => {
                return {r with id: l.id}
            },
        )
            |> tail(n: 13)

    testing.diff(want: want, got: got)
}
