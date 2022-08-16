package universe_test


import "array"
import "csv"
import "testing"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_measurement,user,_field
,,0,2018-05-22T19:53:26Z,0,CPU,user1,a
,,0,2018-05-22T19:53:36Z,1,CPU,user1,a
,,1,2018-05-22T19:53:26Z,4,CPU,user2,a
,,1,2018-05-22T19:53:36Z,20,CPU,user2,a
,,1,2018-05-22T19:53:46Z,7,CPU,user2,a
,,1,2018-05-22T19:53:56Z,10,CPU,user2,a
,,2,2018-05-22T19:53:26Z,1,RAM,user1,b
,,2,2018-05-22T19:53:36Z,2,RAM,user1,b
,,2,2018-05-22T19:53:46Z,3,RAM,user1,b
,,2,2018-05-22T19:53:56Z,5,RAM,user1,b
,,3,2018-05-22T19:53:26Z,2,RAM,user2,b
,,3,2018-05-22T19:53:36Z,4,RAM,user2,b
,,3,2018-05-22T19:53:46Z,4,RAM,user2,b
,,3,2018-05-22T19:53:56Z,0,RAM,user2,b
,,3,2018-05-22T19:54:06Z,2,RAM,user2,b
,,3,2018-05-22T19:54:16Z,10,RAM,user2,b
"
outData =
    "
#datatype,string,long,string,dateTime:RFC3339,double,double,string,string,string
#group,false,false,true,false,false,false,true,false,true
#default,_result,,,,,,,,
,result,table,_measurement,_time,_value_left,_value_right,user_left,user_right,_field
,,0,CPU,2018-05-22T19:53:26Z,0,4,user1,user2,a
,,0,CPU,2018-05-22T19:53:36Z,1,20,user1,user2,b
,,1,RAM,2018-05-22T19:53:26Z,1,2,user1,user2,b
,,1,RAM,2018-05-22T19:53:36Z,2,4,user1,user2,b
,,1,RAM,2018-05-22T19:53:46Z,3,4,user1,user2,b
,,1,RAM,2018-05-22T19:53:56Z,5,0,user1,user2,b
"

testcase join_base {
    input =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:55:00Z)
            |> drop(columns: ["_start", "_stop"])
    want = csv.from(csv: outData)
    left =
        input
            |> filter(fn: (r) => r.user == "user1")
            |> group(columns: ["user"])
    right =
        input
            |> filter(fn: (r) => r.user == "user2")
            |> group(columns: ["_measurement", "_field"])

    got = join(tables: {left: left, right: right}, on: ["_time", "_measurement", "_field"])

    testing.diff(want: want, got: got)
}

testcase join_repro_4692 {
    // Running `findRecord` repeatedly on the same stream can clone the
    // inputs to a join, mutating the original spec each time to refer to more
    // and more input streams each time. This leads to a panic.
    //
    // This tests verifies that repeated runs of `findRecord` on a `join` is
    // "safe" to do in terms of "it no longer panics."
    // Refs: <https://github.com/influxdata/flux/issues/4692>
    xs = array.from(rows: [{id: 1, x: 1}, {id: 2, x: 2}, {id: 3, x: 3}])
    ys = array.from(rows: [{id: 1, y: 1}, {id: 2, y: 2}, {id: 3, y: 3}])
    zs = join(tables: {xs: xs, ys: ys}, on: ["id"])

    getById = (id) => {
        r = zs |> filter(fn: (r) => r.id == id) |> findRecord(fn: (key) => true, idx: 0)

        return {x: r.x, y: r.y}
    }

    got =
        array.from(
            rows: [
                {_value: 1},
                {_value: 2},
                // repeated lookups for id=2 should be okay...
                {_value: 2},
                {_value: 3},
            ],
        )
            |> map(fn: (r) => getById(id: r._value))

    want = array.from(rows: [{x: 1, y: 1}, {x: 2, y: 2}, {x: 2, y: 2}, {x: 3, y: 3}])

    testing.diff(want: want, got: got)
}

testcase join_repro_4692_2 {
    xs = array.from(rows: [{id: 1, x: 1, v: "hi"}, {id: 2, x: 2, v: "hi"}, {id: 3, x: 3, v: "hi"}])
    ys = array.from(rows: [{id: 1, y: 1}, {id: 2, y: 2}, {id: 3, y: 3}])

    getById = (id) => {
        zs =
            if id == 2 then
                join(tables: {xs: xs |> map(fn: (r) => ({r with v: "hey"})), ys: ys}, on: ["id"])
            else
                join(tables: {xs: xs, ys: ys}, on: ["id"])

        r = zs |> filter(fn: (r) => r.id == id) |> findRecord(fn: (key) => true, idx: 0)

        return {x: r.x, y: r.y, v: r.v}
    }

    got =
        array.from(
            rows: [
                {_value: 1},
                {_value: 2},
                // repeated lookups for id=2 should be okay...
                {_value: 2},
                {_value: 3},
            ],
        )
            |> map(fn: (r) => getById(id: r._value))

    want =
        array.from(
            rows: [
                {x: 1, y: 1, v: "hi"},
                {x: 2, y: 2, v: "hey"},
                {x: 2, y: 2, v: "hey"},
                {x: 3, y: 3, v: "hi"},
            ],
        )

    testing.diff(want: want, got: got)
}

testcase join_repro_4692_3 {
    xs = array.from(rows: [{id: 1, x: 1, v: "hi"}, {id: 2, x: 2, v: "hi"}, {id: 3, x: 3, v: "hi"}])
    ys = array.from(rows: [{id: 1, y: 1}, {id: 2, y: 2}, {id: 3, y: 3}])

    getById = (id, a, b) => {
        zs =
            if id == 2 then
                a
            else
                b

        r = zs |> filter(fn: (r) => r.id == id) |> findRecord(fn: (key) => true, idx: 0)

        return {x: r.x, y: r.y, v: r.v}
    }

    got =
        array.from(
            rows: [
                {_value: 1},
                {_value: 2},
                // repeated lookups for id=2 should be okay...
                {_value: 2},
                {_value: 3},
            ],
        )
            |> map(
                fn: (r) => {
                    a =
                        join(
                            tables: {xs: xs |> map(fn: (r) => ({r with v: "hey"})), ys: ys},
                            on: ["id"],
                        )
                    b = join(tables: {xs: xs, ys: ys}, on: ["id"])

                    return getById(id: r._value, a: a, b: b)
                },
            )

    want =
        array.from(
            rows: [
                {x: 1, y: 1, v: "hi"},
                {x: 2, y: 2, v: "hey"},
                {x: 2, y: 2, v: "hey"},
                {x: 3, y: 3, v: "hi"},
            ],
        )

    testing.diff(want: want, got: got)
}
