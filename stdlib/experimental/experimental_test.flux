package experimental_test


import "array"
import "csv"
import "experimental"
import "internal/debug"
import "testing"

testcase addDuration_to_time {
    option now = () => 2021-12-09T00:00:00Z

    cases =
        array.from(
            rows: [
                {d: int(v: 1h), to: now()},
                {d: int(v: 2h), to: now()},
                {d: int(v: 2s), to: now()},
                {d: int(v: 2h), to: 2020-01-01T00:00:00Z},
                {d: int(v: 3d), to: 2020-01-01T00:00:00Z},
            ],
        )
    want =
        array.from(
            rows: [
                {_value: 2021-12-09T01:00:00Z},
                {_value: 2021-12-09T02:00:00Z},
                {_value: 2021-12-09T00:00:02Z},
                {_value: 2020-01-01T02:00:00Z},
                {_value: 2020-01-04T00:00:00Z},
            ],
        )

    got =
        cases |> map(fn: (r) => ({_value: experimental.addDuration(d: duration(v: r.d), to: r.to)}))

    testing.diff(want: want, got: got)
}
testcase addDuration_to_time_mixed {
    //TODO(nathaniel): When we either have the ability to assert on scalar values or can represent durations in tables
    // we can fold this test case into the test case above
    option now = () => 2021-12-09T00:00:00Z

    want = array.from(rows: [{_value: 2022-03-09T00:00:00Z}])

    got = array.from(rows: [{_value: experimental.addDuration(d: 3mo, to: now())}])

    testing.diff(want: want, got: got)
}
testcase addDuration_to_duration_as_time {
    option now = () => 2021-12-09T00:00:00Z

    cases =
        array.from(
            rows: [
                {d: int(v: 1h), to: int(v: -1h)},
                {d: int(v: 1h), to: int(v: -1d)},
                {d: int(v: 1h), to: int(v: -1w)},
                {d: int(v: 1d), to: int(v: -1h)},
                {d: int(v: 1s), to: int(v: -1d)},
                {d: int(v: 1ms), to: int(v: -1w)},
            ],
        )
    want =
        array.from(
            rows: [
                {_value: 2021-12-09T00:00:00Z},
                {_value: 2021-12-08T01:00:00Z},
                {_value: 2021-12-02T01:00:00Z},
                {_value: 2021-12-09T23:00:00Z},
                {_value: 2021-12-08T00:00:01Z},
                {_value: 2021-12-02T00:00:00.001Z},
            ],
        )

    got =
        cases
            |> map(
                fn: (r) =>
                    ({
                        _value:
                            experimental.addDuration(d: duration(v: r.d), to: duration(v: r.to)),
                    }),
            )

    testing.diff(want: want, got: got)
}
testcase addDuration_to_duration_mixed {
    //TODO(nathaniel): When we either have the ability to assert on scalar values or can represent durations in tables
    // we can fold this test case into the test case above
    option now = () => 2021-12-09T00:00:00Z

    want = array.from(rows: [{_value: 2022-01-08T12:00:00Z}])

    got = array.from(rows: [{_value: experimental.addDuration(d: 1mo, to: -12h)}])

    testing.diff(want: want, got: got)
}

testcase subDuration_to_time {
    option now = () => 2021-12-09T00:00:00Z

    cases =
        array.from(
            rows: [
                {d: int(v: 1h), from: now()},
                {d: int(v: 2h), from: now()},
                {d: int(v: 2s), from: now()},
                {d: int(v: 2h), from: 2020-01-01T00:00:00Z},
                {d: int(v: 3d), from: 2020-01-01T00:00:00Z},
            ],
        )
    want =
        array.from(
            rows: [
                {_value: 2021-12-08T23:00:00Z},
                {_value: 2021-12-08T22:00:00Z},
                {_value: 2021-12-08T23:59:58Z},
                {_value: 2019-12-31T22:00:00Z},
                {_value: 2019-12-29T00:00:00Z},
            ],
        )

    got =
        cases
            |> map(
                fn: (r) => ({_value: experimental.subDuration(d: duration(v: r.d), from: r.from)}),
            )

    testing.diff(want: want, got: got)
}
testcase subDuration_to_time_mixed {
    //TODO(nathaniel): When we either have the ability to assert on scalar values or can represent durations in tables
    // we can fold this test case into the test case above
    option now = () => 2021-12-09T00:00:00Z

    want = array.from(rows: [{_value: 2020-12-09T00:00:00Z}])

    got = array.from(rows: [{_value: experimental.subDuration(d: 1y, from: now())}])

    testing.diff(want: want, got: got)
}
testcase subDuration_to_duration_as_time {
    option now = () => 2021-12-09T00:00:00Z

    cases =
        array.from(
            rows: [
                {d: int(v: 1h), from: int(v: -1h)},
                {d: int(v: 1h), from: int(v: -1d)},
                {d: int(v: 1h), from: int(v: -1w)},
                {d: int(v: 1d), from: int(v: -1h)},
                {d: int(v: 1s), from: int(v: -1d)},
                {d: int(v: 1ms), from: int(v: -1w)},
            ],
        )
    want =
        array.from(
            rows: [
                {_value: 2021-12-08T22:00:00Z},
                {_value: 2021-12-07T23:00:00Z},
                {_value: 2021-12-01T23:00:00Z},
                {_value: 2021-12-07T23:00:00Z},
                {_value: 2021-12-07T23:59:59Z},
                {_value: 2021-12-01T23:59:59.999Z},
            ],
        )

    got =
        cases
            |> map(
                fn: (r) =>
                    ({
                        _value:
                            experimental.subDuration(
                                d: duration(v: r.d),
                                from: duration(v: r.from),
                            ),
                    }),
            )

    testing.diff(want: want, got: got)
}
testcase subDuration_to_duration_mixed {
    //TODO(nathaniel): When we either have the ability to assert on scalar values or can represent durations in tables
    // we can fold this test case into the test case above
    option now = () => 2021-12-09T00:00:00Z

    want = array.from(rows: [{_value: 2019-12-08T12:00:00Z}])

    got = array.from(rows: [{_value: experimental.subDuration(d: 2y, from: -12h)}])

    testing.diff(want: want, got: got)
}

inData =
    "
#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,iZquGj,ei77f8T,2018-12-18T20:52:33Z,-61.68790887989735
,,0,iZquGj,ei77f8T,2018-12-18T20:52:43Z,-6.3173755351186465
,,0,iZquGj,ei77f8T,2018-12-18T20:52:53Z,-26.049728557657513
,,0,iZquGj,ei77f8T,2018-12-18T20:53:03Z,114.285955884979
,,0,iZquGj,ei77f8T,2018-12-18T20:53:13Z,16.140262630578995
,,0,iZquGj,ei77f8T,2018-12-18T20:53:23Z,29.50336437998469

#datatype,string,long,string,string,dateTime:RFC3339,long
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,1,iZquGj,ucyoZ,2018-12-18T20:52:33Z,-66
,,1,iZquGj,ucyoZ,2018-12-18T20:52:43Z,59
,,1,iZquGj,ucyoZ,2018-12-18T20:52:53Z,64
,,1,iZquGj,ucyoZ,2018-12-18T20:53:03Z,84
,,1,iZquGj,ucyoZ,2018-12-18T20:53:13Z,68
,,1,iZquGj,ucyoZ,2018-12-18T20:53:23Z,49
"

testcase unpivot_pivot_roundtrip {
    want =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-12-18T20:00:00Z, stop: 2018-12-18T21:00:00Z)

    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-12-18T20:00:00Z, stop: 2018-12-18T21:00:00Z)
            |> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
            |> experimental.unpivot()

    testing.diff(got: got, want: want)
        |> yield()
}

inDataUnpivoted =
    "
#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,iZquGj,f,2018-12-18T20:52:33Z,-61.68790887989735
,,0,iZquGj,f,2018-12-18T20:52:43Z,-6.3173755351186465

#datatype,string,long,string,string,dateTime:RFC3339,long
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,1,iZquGj,i,2018-12-18T20:52:33Z,-66
,,1,iZquGj,i,2018-12-18T20:52:43Z,3


#datatype,string,long,string,string,dateTime:RFC3339,string
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,2,iZquGj,s,2018-12-18T20:52:33Z,abc
,,2,iZquGj,s,2018-12-18T20:52:43Z,123
"

inDataPivoted =
    "
#datatype,string,long,string,dateTime:RFC3339,double,long,string
#group,false,false,true,false,false,false,false
#default,_result,,,,,,
,result,table,_measurement,_time,f,i,s
,,0,iZquGj,2018-12-18T20:52:33Z,-61.68790887989735,-66,abc
,,0,iZquGj,2018-12-18T20:52:43Z,-6.3173755351186465,3,123
"

testcase unpivot_3_columns {
    want =
        csv.from(csv: inDataUnpivoted)
            |> testing.load()
            |> range(start: 2018-12-18T20:00:00Z, stop: 2018-12-18T21:00:00Z)

    got =
        csv.from(csv: inDataPivoted)
            |> range(start: 2018-12-18T20:00:00Z, stop: 2018-12-18T21:00:00Z)
            |> experimental.unpivot()

    testing.diff(got: got, want: want)
        |> yield()
}

testcase unpivot_with_nulls {
    input =
        array.from(
            rows: [
                {_time: 2018-12-18T20:52:33Z, a: 1.0, b: debug.null(type: "string")},
                {_time: 2018-12-18T20:52:33Z, a: debug.null(type: "float"), b: "abc"},
            ],
        )

    want =
        array.from(rows: [{_time: 2018-12-18T20:52:33Z, _field: "a", _value: 1.0}])
            |> group(columns: ["_field"])

    got =
        input
            |> experimental.unpivot()
            |> filter(fn: (r) => r._field == "a")

    testing.diff(want, got)
}

testcase unpivot_with_nulls_2 {
    input =
        array.from(
            rows: [
                {_time: 2018-12-18T20:52:33Z, a: debug.null(type: "float"), b: "abc"},
                {_time: 2018-12-18T20:52:33Z, a: 1.0, b: debug.null(type: "string")},
            ],
        )

    want =
        array.from(rows: [{_time: 2018-12-18T20:52:33Z, _field: "a", _value: 1.0}])
            |> group(columns: ["_field"])

    got =
        input
            |> experimental.unpivot()
            |> filter(fn: (r) => r._field == "a")

    testing.diff(want, got)
}

testcase unpivot_other_columns {
    got =
        array.from(
            rows: [
                {
                    _measurement: "m",
                    tag: "t1",
                    f0: 10.1,
                    f1: 10.2,
                    _time: 2018-12-01T00:00:00Z,
                },
                {
                    _measurement: "m",
                    tag: "t1",
                    f0: 20.1,
                    f1: 20.2,
                    _time: 2018-12-01T00:00:10Z,
                },
            ],
        )
            |> group(columns: ["_measurement"])
            |> experimental.unpivot(otherColumns: ["_time", "tag"])

    want =
        array.from(
            rows: [
                {
                    _measurement: "m",
                    tag: "t1",
                    _field: "f0",
                    _value: 10.1,
                    _time: 2018-12-01T00:00:00Z,
                },
                {
                    _measurement: "m",
                    tag: "t1",
                    _field: "f1",
                    _value: 10.2,
                    _time: 2018-12-01T00:00:00Z,
                },
                {
                    _measurement: "m",
                    tag: "t1",
                    _field: "f0",
                    _value: 20.1,
                    _time: 2018-12-01T00:00:10Z,
                },
                {
                    _measurement: "m",
                    tag: "t1",
                    _field: "f1",
                    _value: 20.2,
                    _time: 2018-12-01T00:00:10Z,
                },
            ],
        )
            |> group(columns: ["_measurement", "_field"])

    testing.diff(want, got)
}

testcase unpivot_field_in_other_columns {
    got =
        array.from(
            rows: [
                {
                    _measurement: "m",
                    tag: "t1",
                    f0: 10.1,
                    f1: 10.2,
                    _time: 2018-12-01T00:00:00Z,
                    _field: "load1",
                },
                {
                    _measurement: "m",
                    tag: "t1",
                    f0: 20.1,
                    f1: 20.2,
                    _time: 2018-12-01T00:00:10Z,
                    _field: "load1",
                },
            ],
        )
            |> group(columns: ["_measurement"])
            |> experimental.unpivot(otherColumns: ["_time", "tag", "_field"])

    want =
        array.from(
            rows: [
                {
                    _measurement: "m",
                    tag: "t1",
                    _field: "f0",
                    _value: 10.1,
                    _time: 2018-12-01T00:00:00Z,
                },
                {
                    _measurement: "m",
                    tag: "t1",
                    _field: "f0",
                    _value: 20.1,
                    _time: 2018-12-01T00:00:10Z,
                },
                {
                    _measurement: "m",
                    tag: "t1",
                    _field: "f1",
                    _value: 10.2,
                    _time: 2018-12-01T00:00:00Z,
                },
                {
                    _measurement: "m",
                    tag: "t1",
                    _field: "f1",
                    _value: 20.2,
                    _time: 2018-12-01T00:00:10Z,
                },
            ],
        )
            |> group(columns: ["_measurement"])

    testing.diff(want, got)
}

testcase unpivot_other_columns_nulls {
    got =
        array.from(
            rows: [
                {
                    _measurement: "m",
                    tag0: "foo",
                    tag1: "bar",
                    f0: 10.1,
                    f1: debug.null(type: "float"),
                    _time: 2018-12-01T00:00:00Z,
                },
                {
                    _measurement: "m",
                    tag0: "foz",
                    tag1: "baz",
                    f0: 20.1,
                    f1: 20.2,
                    _time: 2018-12-01T00:00:10Z,
                },
            ],
        )
            |> group(columns: ["_measurement"])
            |> experimental.unpivot(otherColumns: ["_time", "tag0", "tag1"])

    want =
        array.from(
            rows: [
                {
                    _measurement: "m",
                    tag0: "foo",
                    tag1: "bar",
                    _field: "f0",
                    _value: 10.1,
                    _time: 2018-12-01T00:00:00Z,
                },
                {
                    _measurement: "m",
                    tag0: "foz",
                    tag1: "baz",
                    _field: "f0",
                    _value: 20.1,
                    _time: 2018-12-01T00:00:10Z,
                },
                {
                    _measurement: "m",
                    tag0: "foz",
                    tag1: "baz",
                    _field: "f1",
                    _value: 20.2,
                    _time: 2018-12-01T00:00:10Z,
                },
            ],
        )
            |> group(columns: ["_measurement", "_field"])

    testing.diff(want, got)
}

testcase unpivot_other_cols_error {
    fn = () =>
        array.from(rows: [{foo: "bar", v0: 10, _time: 2020-01-01T00:00:00Z}])
            |> experimental.unpivot(otherColumns: ["does not exist"])
            |> tableFind(fn: (key) => true)

    testing.shouldError(fn: fn, want: /unpivot could not find column named "does not exist"/)
}
