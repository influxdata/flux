package universe_test


import "array"
import "testing"

inData = array.from(
    rows: [
        {_time: 2018-05-22T20:00:00Z, _value: 6.05, _measurement: "m0", _field: "f0", t0: "a"},
        {_time: 2018-05-22T20:00:10Z, _value: 9.41, _measurement: "m0", _field: "f0", t0: "a"},
        {_time: 2018-05-22T20:00:20Z, _value: 6.65, _measurement: "m0", _field: "f0", t0: "a"},
        {_time: 2018-05-22T20:00:30Z, _value: 4.37, _measurement: "m0", _field: "f0", t0: "a"},
        {_time: 2018-05-22T20:00:40Z, _value: 4.25, _measurement: "m0", _field: "f0", t0: "a"},
        {_time: 2018-05-22T20:00:00Z, _value: 6.87, _measurement: "m0", _field: "f0", t0: "b"},
        {_time: 2018-05-22T20:00:10Z, _value: 0.66, _measurement: "m0", _field: "f0", t0: "b"},
        {_time: 2018-05-22T20:00:20Z, _value: 1.57, _measurement: "m0", _field: "f0", t0: "b"},
        {_time: 2018-05-22T20:00:30Z, _value: 0.97, _measurement: "m0", _field: "f0", t0: "b"},
        {_time: 2018-05-22T20:00:40Z, _value: 3.01, _measurement: "m0", _field: "f0", t0: "b"},
    ]
)
    |> group(columns: ["_measurement", "_field", "t0"])

testcase default {
    want = array.from(
        rows: [
            {_time: 2018-05-22T20:00:10Z, _value: 0.336, _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:20Z, _value: -0.276, _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:30Z, _value: -0.228, _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:40Z, _value: -0.012, _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:10Z, _value: -0.621, _measurement: "m0", _field: "f0", t0: "b"},
            {_time: 2018-05-22T20:00:20Z, _value: 0.091, _measurement: "m0", _field: "f0", t0: "b"},
            {_time: 2018-05-22T20:00:30Z, _value: -0.06, _measurement: "m0", _field: "f0", t0: "b"},
            {_time: 2018-05-22T20:00:40Z, _value: 0.204, _measurement: "m0", _field: "f0", t0: "b"},
        ]
    )
        |> group(columns: ["_measurement", "_field", "t0"])

    got = inData
        |> range(start: 2018-05-22T20:00:00Z, stop: 2018-05-22T20:01:00Z)
        |> derivative()
        |> drop(columns: ["_start", "_stop"])

    testing.diff(want: want, got: got) |> yield()
}

testcase non_negative {
    want = array.from(
        rows: [
            {_time: 2018-05-22T20:00:10Z, _value: 0.336, _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:20Z, _value: 0.091, _measurement: "m0", _field: "f0", t0: "b"},
            {_time: 2018-05-22T20:00:40Z, _value: 0.204, _measurement: "m0", _field: "f0", t0: "b"},
        ]
    )
        |> group(columns: ["_measurement", "_field", "t0"])

    got = inData
        |> range(start: 2018-05-22T20:00:00Z, stop: 2018-05-22T20:01:00Z)
        |> derivative(nonNegative: true)
        |> filter(fn: (r) => exists r._value)
        |> drop(columns: ["_start", "_stop"])

    testing.diff(want: want, got: got) |> yield()
}

// This test ensures that non-numeric types that go through derivative get
// passed through. That's the current behavior, but it's also likely a bug.
// Remove this if we choose to fix https://github.com/influxdata/flux/issues/914.
testcase passthrough {
    input = array.from(
        rows: [
            {_time: 2018-05-22T20:00:00Z, _value: "a", _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:10Z, _value: "b", _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:20Z, _value: "c", _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:30Z, _value: "d", _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:40Z, _value: "e", _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:00Z, _value: "f", _measurement: "m0", _field: "f0", t0: "b"},
            {_time: 2018-05-22T20:00:10Z, _value: "g", _measurement: "m0", _field: "f0", t0: "b"},
            {_time: 2018-05-22T20:00:20Z, _value: "h", _measurement: "m0", _field: "f0", t0: "b"},
            {_time: 2018-05-22T20:00:30Z, _value: "i", _measurement: "m0", _field: "f0", t0: "b"},
            {_time: 2018-05-22T20:00:40Z, _value: "j", _measurement: "m0", _field: "f0", t0: "b"},
        ]
    )
        |> group(columns: ["_measurement", "_field", "t0"])

    want = array.from(
        rows: [
            {_time: 2018-05-22T20:00:10Z, _value: "b", _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:20Z, _value: "c", _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:30Z, _value: "d", _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:40Z, _value: "e", _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:10Z, _value: "g", _measurement: "m0", _field: "f0", t0: "b"},
            {_time: 2018-05-22T20:00:20Z, _value: "h", _measurement: "m0", _field: "f0", t0: "b"},
            {_time: 2018-05-22T20:00:30Z, _value: "i", _measurement: "m0", _field: "f0", t0: "b"},
            {_time: 2018-05-22T20:00:40Z, _value: "j", _measurement: "m0", _field: "f0", t0: "b"},
        ]
    )
        |> group(columns: ["_measurement", "_field", "t0"])

    got = input
        |> range(start: 2018-05-22T20:00:00Z, stop: 2018-05-22T20:01:00Z)
        |> derivative()
        |> drop(columns: ["_start", "_stop"])

    testing.diff(want: want, got: got) |> yield()
}

testcase duplicate_times {
    want = array.from(
        rows: [
            {_time: 2018-05-22T20:00:10Z, _value: 0.336, _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:20Z, _value: -0.276, _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:30Z, _value: -0.228, _measurement: "m0", _field: "f0", t0: "a"},
            {_time: 2018-05-22T20:00:40Z, _value: -0.012, _measurement: "m0", _field: "f0", t0: "a"},
        ]
    )
        |> group(columns: ["_measurement", "_field"])

    got = inData
        |> range(start: 2018-05-22T20:00:00Z, stop: 2018-05-22T20:01:00Z)
        |> group(columns: ["_measurement", "_field"])
        |> sort(columns: ["_time", "t0"])
        |> derivative()
        |> drop(columns: ["_start", "_stop"])

    testing.diff(want: want, got: got) |> yield()
}
