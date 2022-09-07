package experimental_test


import "testing"
import "array"
import "experimental"
import "internal/debug"

testcase match {
    rows = [{_measurement: "m0", _field: "f0", _time: 2022-07-12T00:00:00Z, _value: 2.0}]

    want =
        array.from(rows: rows)
            |> group(columns: ["_measurement", "_field"])
    got =
        array.from(rows: rows)
            |> group(columns: ["_measurement", "_field"])

    experimental.diff(want, got)
        |> testing.assertEmpty()
}

testcase mismatch {
    want =
        array.from(
            rows: [{_measurement: "m0", _field: "f0", _time: 2022-07-12T00:00:00Z, _value: 2.0}],
        )
            |> group(columns: ["_measurement", "_field"])

    got =
        array.from(
            rows: [{_measurement: "m0", _field: "f0", _time: 2022-07-12T00:00:00Z, _value: 3.0}],
        )
            |> group(columns: ["_measurement", "_field"])

    exp =
        array.from(
            rows: [
                {
                    diff: "-",
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T00:00:00Z,
                    _value: 2.0,
                },
                {
                    diff: "+",
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T00:00:00Z,
                    _value: 3.0,
                },
            ],
        )
            |> group(columns: ["_measurement", "_field"])

    experimental.diff(want, got)
        |> rename(columns: {_diff: "diff"})
        |> testing.diff(want: exp)
}

testcase partial_match {
    want =
        array.from(
            rows: [
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T00:00:00Z, _value: 2.0},
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T01:00:00Z, _value: 3.0},
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T02:00:00Z, _value: 4.0},
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T03:00:00Z, _value: 5.0},
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T04:00:00Z, _value: 6.0},
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T05:00:00Z, _value: 7.0},
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T07:00:00Z, _value: 9.0},
            ],
        )
            |> group(columns: ["_measurement", "_field"])

    got =
        array.from(
            rows: [
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T00:00:00Z, _value: 2.0},
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T03:00:00Z, _value: 5.0},
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T04:00:00Z, _value: 6.0},
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T05:00:00Z, _value: 7.0},
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T06:00:00Z, _value: 8.0},
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T07:00:00Z, _value: 10.0},
            ],
        )
            |> group(columns: ["_measurement", "_field"])

    exp =
        array.from(
            rows: [
                {
                    diff: "",
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T00:00:00Z,
                    _value: 2.0,
                },
                {
                    diff: "-",
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T01:00:00Z,
                    _value: 3.0,
                },
                {
                    diff: "-",
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T02:00:00Z,
                    _value: 4.0,
                },
                {
                    diff: "",
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T03:00:00Z,
                    _value: 5.0,
                },
                {
                    diff: "",
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T04:00:00Z,
                    _value: 6.0,
                },
                {
                    diff: "",
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T05:00:00Z,
                    _value: 7.0,
                },
                {
                    diff: "-",
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T07:00:00Z,
                    _value: 9.0,
                },
                {
                    diff: "+",
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T06:00:00Z,
                    _value: 8.0,
                },
                {
                    diff: "+",
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T07:00:00Z,
                    _value: 10.0,
                },
            ],
        )
            |> group(columns: ["_measurement", "_field"])

    experimental.diff(want, got)
        |> rename(columns: {_diff: "diff"})
        |> testing.diff(want: exp)
}

testcase empty_want {
    want =
        array.from(
            rows: [{_measurement: "m0", _field: "f0", _time: 2022-07-12T00:00:00Z, _value: 2.0}],
        )
            |> group(columns: ["_measurement", "_field"])

    got =
        array.from(
            rows: [
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T00:00:00Z, _value: 2.0},
                {_measurement: "m0", _field: "f1", _time: 2022-07-12T00:00:00Z, _value: 3.0},
            ],
        )
            |> group(columns: ["_measurement", "_field"])

    exp =
        array.from(
            rows: [
                {
                    diff: "+",
                    _measurement: "m0",
                    _field: "f1",
                    _time: 2022-07-12T00:00:00Z,
                    _value: 3.0,
                },
            ],
        )
            |> group(columns: ["_measurement", "_field"])

    experimental.diff(want, got)
        |> rename(columns: {_diff: "diff"})
        |> testing.diff(want: exp)
}

testcase empty_got {
    want =
        array.from(
            rows: [
                {_measurement: "m0", _field: "f0", _time: 2022-07-12T00:00:00Z, _value: 2.0},
                {_measurement: "m0", _field: "f1", _time: 2022-07-12T00:00:00Z, _value: 3.0},
            ],
        )
            |> group(columns: ["_measurement", "_field"])

    got =
        array.from(
            rows: [{_measurement: "m0", _field: "f0", _time: 2022-07-12T00:00:00Z, _value: 2.0}],
        )
            |> group(columns: ["_measurement", "_field"])

    exp =
        array.from(
            rows: [
                {
                    diff: "-",
                    _measurement: "m0",
                    _field: "f1",
                    _time: 2022-07-12T00:00:00Z,
                    _value: 3.0,
                },
            ],
        )
            |> group(columns: ["_measurement", "_field"])

    experimental.diff(want, got)
        |> rename(columns: {_diff: "diff"})
        |> testing.diff(want: exp)
}

testcase mismatch_non_string_group_key {
    want =
        array.from(
            rows: [
                {
                    _measurement: "m0",
                    _field: "f0",
                    _start: 2022-07-12T00:00:00Z,
                    _time: 2022-07-12T00:00:00Z,
                    _value: 2.0,
                },
            ],
        )
            |> group(columns: ["_measurement", "_field", "_start"])

    got =
        array.from(
            rows: [
                {
                    _measurement: "m0",
                    _field: "f0",
                    _start: 2022-07-12T00:00:00Z,
                    _time: 2022-07-12T00:00:00Z,
                    _value: 3.0,
                },
            ],
        )
            |> group(columns: ["_measurement", "_field", "_start"])

    exp =
        array.from(
            rows: [
                {
                    diff: "-",
                    _measurement: "m0",
                    _field: "f0",
                    _start: 2022-07-12T00:00:00Z,
                    _time: 2022-07-12T00:00:00Z,
                    _value: 2.0,
                },
                {
                    diff: "+",
                    _measurement: "m0",
                    _field: "f0",
                    _start: 2022-07-12T00:00:00Z,
                    _time: 2022-07-12T00:00:00Z,
                    _value: 3.0,
                },
            ],
        )
            |> group(columns: ["_measurement", "_field", "_start"])

    experimental.diff(want, got)
        |> rename(columns: {_diff: "diff"})
        |> testing.diff(want: exp)
}

testcase mismatch_null {
    want =
        array.from(
            rows: [{_measurement: "m0", _field: "f0", _time: 2022-07-12T00:00:00Z, _value: 2.0}],
        )
            |> group(columns: ["_measurement", "_field"])

    got =
        array.from(
            rows: [
                {
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T00:00:00Z,
                    _value: debug.null(type: "float"),
                },
            ],
        )
            |> group(columns: ["_measurement", "_field"])

    exp =
        array.from(
            rows: [
                {
                    diff: "-",
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T00:00:00Z,
                    _value: 2.0,
                },
                {
                    diff: "+",
                    _measurement: "m0",
                    _field: "f0",
                    _time: 2022-07-12T00:00:00Z,
                    _value: debug.null(type: "float"),
                },
            ],
        )
            |> group(columns: ["_measurement", "_field"])

    experimental.diff(want, got)
        |> rename(columns: {_diff: "diff"})
        |> testing.diff(want: exp)
}
