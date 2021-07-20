package universe_test


import "array"
import "testing"
import "internal/debug"

inData = [
    {_time: 2018-05-22T00:00:00Z, _value: 30, _measurement: "disk", _field: "used_percent"},
    {_time: 2018-05-22T00:00:10Z, _value: 35, _measurement: "disk", _field: "used_percent"},
    {_time: 2018-05-22T00:00:20Z, _value: 40, _measurement: "disk", _field: "used_percent"},
    {_time: 2018-05-22T00:00:30Z, _value: 45, _measurement: "disk", _field: "used_percent"},
    {_time: 2018-05-22T00:00:40Z, _value: 50, _measurement: "disk", _field: "used_percent"},
    {_time: 2018-05-22T00:00:50Z, _value: 55, _measurement: "disk", _field: "used_percent"},
]

// We use debug.slurp to ensure that moving average always produces tables where
// the column sizes are the same. Moving average doesn't check this but debug.slurp does.
runTest = (n) => array.from(rows: inData)
    |> group(columns: ["_measurement", "_field"])
    |> testing.load()
    |> range(start: 2018-05-22T00:00:00Z, stop: 2018-05-22T00:01:00Z)
    |> limit(n: n)
    |> movingAverage(n: 3)
    |> debug.slurp()
    |> drop(columns: ["_start", "_stop"])

testcase normal {
    got = runTest(n: 6)
    want = array.from(
        rows: [
            {_time: 2018-05-22T00:00:20Z, _value: 35.0, _measurement: "disk", _field: "used_percent"},
            {_time: 2018-05-22T00:00:30Z, _value: 40.0, _measurement: "disk", _field: "used_percent"},
            {_time: 2018-05-22T00:00:40Z, _value: 45.0, _measurement: "disk", _field: "used_percent"},
            {_time: 2018-05-22T00:00:50Z, _value: 50.0, _measurement: "disk", _field: "used_percent"},
        ],
    )
        |> group(columns: ["_measurement", "_field"])

    testing.diff(got, want) |> yield()
}

testcase unfilled {
    got = runTest(n: 1)
    want = array.from(
        rows: [
            {_time: 2018-05-22T00:00:00Z, _value: 30.0, _measurement: "disk", _field: "used_percent"},
        ],
    )
        |> group(columns: ["_measurement", "_field"])

    testing.diff(got, want) |> yield()
}

testcase exact {
    got = runTest(n: 3)
    want = array.from(
        rows: [
            {_time: 2018-05-22T00:00:20Z, _value: 35.0, _measurement: "disk", _field: "used_percent"},
        ],
    )
        |> group(columns: ["_measurement", "_field"])

    testing.diff(got, want) |> yield()
}
