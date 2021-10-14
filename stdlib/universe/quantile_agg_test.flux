package universe_test


import "array"
import "testing"

inData = array.from(
    rows: [
        {_time: 2018-05-22T20:00:00Z, _value: 6.05, t0: "a"},
        {_time: 2018-05-22T20:00:10Z, _value: 9.41, t0: "a"},
        {_time: 2018-05-22T20:00:20Z, _value: 6.65, t0: "a"},
        {_time: 2018-05-22T20:00:30Z, _value: 4.37, t0: "a"},
        {_time: 2018-05-22T20:00:40Z, _value: 4.25, t0: "a"},
        {_time: 2018-05-22T20:00:00Z, _value: 6.87, t0: "b"},
        {_time: 2018-05-22T20:00:10Z, _value: 0.66, t0: "b"},
        {_time: 2018-05-22T20:00:20Z, _value: 1.57, t0: "b"},
        {_time: 2018-05-22T20:00:30Z, _value: 0.97, t0: "b"},
        {_time: 2018-05-22T20:00:40Z, _value: 3.01, t0: "b"},
        {_time: 2018-05-22T20:00:00Z, _value: 7.85, t0: "c"},
        {_time: 2018-05-22T20:00:10Z, _value: 3.52, t0: "c"},
        {_time: 2018-05-22T20:00:20Z, _value: 4.21, t0: "c"},
        {_time: 2018-05-22T20:00:30Z, _value: 8.96, t0: "c"},
        {_time: 2018-05-22T20:00:40Z, _value: 4.24, t0: "c"},
        {_time: 2018-05-22T20:00:00Z, _value: 3.43, t0: "d"},
        {_time: 2018-05-22T20:00:10Z, _value: 1.11, t0: "d"},
        {_time: 2018-05-22T20:00:20Z, _value: 0.48, t0: "d"},
        {_time: 2018-05-22T20:00:30Z, _value: 6.35, t0: "d"},
        {_time: 2018-05-22T20:00:40Z, _value: 5.23, t0: "d"},
    ],
)

testcase quantile_with_group {
    want = array.from(
        rows: [
            {_value: 7.34, t0: "a"},
            {_value: 3.975, t0: "b"},
            {_value: 8.1275, t0: "c"},
            {_value: 5.51, t0: "d"},
        ],
    )
        |> group(columns: ["t0"])

    got = inData
        |> range(start: 2018-05-22T20:00:00Z, stop: 2018-05-22T20:01:00Z)
        |> group(columns: ["t0"])
        |> quantile(q: 0.75, method: "estimate_tdigest")
        |> drop(columns: ["_start", "_stop"])

    testing.diff(want: want, got: got) |> yield()
}

testcase quantile_without_group {
    want = array.from(
        rows: [
            {_value: 6.5},
        ],
    )

    got = inData
        |> range(start: 2018-05-22T20:00:00Z, stop: 2018-05-22T20:01:00Z)
        |> quantile(q: 0.75, method: "estimate_tdigest")
        |> drop(columns: ["_start", "_stop", "t0"])

    testing.diff(want: want, got: got) |> yield()
}
