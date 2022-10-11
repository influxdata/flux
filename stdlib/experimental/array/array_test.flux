package array_test


import "experimental/array"
import "testing"

testcase array_concat_exp {
    got =
        array.from(
            rows:
                ["x", "y", "z"]
                    |> array.concat(v: ["a", "b", "c"])
                    |> array.map(fn: (x) => ({_value: x})),
        )

    want =
        array.from(
            rows: [
                {_value: "x"},
                {_value: "y"},
                {_value: "z"},
                {_value: "a"},
                {_value: "b"},
                {_value: "c"},
            ],
        )

    testing.diff(got, want)
}
testcase array_concat_to_empty_exp {
    got =
        array.from(
            rows:
                []
                    |> array.concat(v: ["a", "b", "c"])
                    |> array.map(fn: (x) => ({_value: x})),
        )

    want = array.from(rows: [{_value: "a"}, {_value: "b"}, {_value: "c"}])

    testing.diff(got, want)
}
testcase array_map_exp {
    got =
        array.from(
            rows:
                ["1", "2", "3"]
                    |> array.map(fn: (x) => ({_value: float(v: x)})),
        )
    want = array.from(rows: [{_value: 1.0}, {_value: 2.0}, {_value: 3.0}])

    testing.diff(want: want, got: got)
}

testcase array_filter_exp {
    got =
        array.from(
            rows:
                [
                    1,
                    2,
                    3,
                    4,
                    5,
                    6,
                    7,
                    8,
                    9,
                    10,
                ]
                    |> array.filter(fn: (x) => x > 5)
                    |> array.map(fn: (x) => ({_value: x})),
        )
    want =
        array.from(
            rows: [
                {_value: 6},
                {_value: 7},
                {_value: 8},
                {_value: 9},
                {_value: 10},
            ],
        )

    testing.diff(want: want, got: got)
}

testcase array_tobool_exp {
    got =
        array.from(
            rows:
                [1, 1, 0, 1]
                    |> array.toBool()
                    |> array.map(fn: (x) => ({_value: x})),
        )
    want = array.from(rows: [{_value: true}, {_value: true}, {_value: false}, {_value: true}])

    testing.diff(want: want, got: got)
}

testcase array_toduration_exp {
    got =
        array.from(
            rows:
                [60000000000, 90000000000, 120000000000, 150000000000]
                    |> array.toDuration()
                    |> array.map(fn: (x) => ({_value: string(v: x)})),
        )
    want = array.from(rows: [{_value: "1m"}, {_value: "1m30s"}, {_value: "2m"}, {_value: "2m30s"}])

    testing.diff(want: want, got: got)
}

testcase array_tofloat_exp {
    got =
        array.from(
            rows:
                [10, 20, 30, 40]
                    |> array.toFloat()
                    |> array.map(fn: (x) => ({_value: x})),
        )
    want = array.from(rows: [{_value: 10.0}, {_value: 20.0}, {_value: 30.0}, {_value: 40.0}])

    testing.diff(want: want, got: got)
}

testcase array_toint_exp {
    got =
        array.from(
            rows:
                [12.1, 24.2, 48.4, 96.8]
                    |> array.toInt()
                    |> array.map(fn: (x) => ({_value: x})),
        )
    want = array.from(rows: [{_value: 12}, {_value: 24}, {_value: 48}, {_value: 96}])

    testing.diff(want: want, got: got)
}

testcase array_tostring_exp {
    got =
        array.from(
            rows:
                [true, true, false, true]
                    |> array.toString()
                    |> array.map(fn: (x) => ({_value: x})),
        )
    want =
        array.from(rows: [{_value: "true"}, {_value: "true"}, {_value: "false"}, {_value: "true"}])

    testing.diff(want: want, got: got)
}

testcase array_totime_exp {
    got =
        array.from(
            rows:
                ["2022-01-01", "2022-01-02", "2022-01-03", "2022-01-04"]
                    |> array.toTime()
                    |> array.map(fn: (x) => ({_value: x})),
        )
    want =
        array.from(
            rows: [
                {_value: 2022-01-01T00:00:00Z},
                {_value: 2022-01-02T00:00:00Z},
                {_value: 2022-01-03T00:00:00Z},
                {_value: 2022-01-04T00:00:00Z},
            ],
        )

    testing.diff(want: want, got: got)
}

testcase array_touint_exp {
    got =
        array.from(
            rows:
                [-12.1, 24.2, -48.4, 96.8]
                    |> array.toUInt()
                    |> array.map(fn: (x) => ({_value: x})),
        )
    want =
        array.from(
            rows: [
                {_value: uint(v: -12.1)},
                {_value: uint(v: 24.2)},
                {_value: uint(v: -48.4)},
                {_value: uint(v: 96.8)},
            ],
        )

    testing.diff(want: want, got: got)
}
