package array_test


import "testing"
import "array"

testcase fromElementTest {
    fromElement = (v) => array.from(rows: [{v: v}])
    want = fromElement(v: 123)
    got = fromElement(v: 123)

    testing.diff(want, got)
        |> yield()
}

testcase array_concat {
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
testcase array_concat_to_empty {
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
testcase array_map {
    got =
        array.from(
            rows:
                ["1", "2", "3"]
                    |> array.map(fn: (x) => ({_value: float(v: x)})),
        )
    want = array.from(rows: [{_value: 1.0}, {_value: 2.0}, {_value: 3.0}])

    testing.diff(want: want, got: got)
}

testcase array_filter {
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
