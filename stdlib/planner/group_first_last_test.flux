package planner_test


import "array"
import "testing"
import "csv"

// two fields, two tags keys, with three rows in each combo
inData =
    array.from(
        rows: [
            {
                _field: "f0",
                _measurement: "m0",
                t0: "t0v0",
                t1: "t1v0",
                _time: 2021-07-06T23:06:30Z,
                _value: 3,
            },
            {
                _field: "f0",
                _measurement: "m0",
                t0: "t0v0",
                t1: "t1v0",
                _time: 2021-07-06T23:06:40Z,
                _value: 1,
            },
            {
                _field: "f0",
                _measurement: "m0",
                t0: "t0v0",
                t1: "t1v0",
                _time: 2021-07-06T23:06:50Z,
                _value: 0,
            },
            {
                _field: "f0",
                _measurement: "m0",
                t0: "t0v0",
                t1: "t1v1",
                _time: 2021-07-06T23:06:30Z,
                _value: 4,
            },
            {
                _field: "f0",
                _measurement: "m0",
                t0: "t0v0",
                t1: "t1v1",
                _time: 2021-07-06T23:06:40Z,
                _value: 3,
            },
            {
                _field: "f0",
                _measurement: "m0",
                t0: "t0v0",
                t1: "t1v1",
                _time: 2021-07-06T23:06:50Z,
                _value: 1,
            },
            {
                _field: "f0",
                _measurement: "m0",
                t0: "t0v1",
                t1: "t1v0",
                _time: 2021-07-06T23:06:30Z,
                _value: 1,
            },
            {
                _field: "f0",
                _measurement: "m0",
                t0: "t0v1",
                t1: "t1v0",
                _time: 2021-07-06T23:06:40Z,
                _value: 0,
            },
            {
                _field: "f0",
                _measurement: "m0",
                t0: "t0v1",
                t1: "t1v0",
                _time: 2021-07-06T23:06:50Z,
                _value: 4,
            },
            {
                _field: "f0",
                _measurement: "m0",
                t0: "t0v1",
                t1: "t1v1",
                _time: 2021-07-06T23:06:30Z,
                _value: 4,
            },
            {
                _field: "f0",
                _measurement: "m0",
                t0: "t0v1",
                t1: "t1v1",
                _time: 2021-07-06T23:06:40Z,
                _value: 0,
            },
            {
                _field: "f0",
                _measurement: "m0",
                t0: "t0v1",
                t1: "t1v1",
                _time: 2021-07-06T23:06:50Z,
                _value: 4,
            },
            {
                _field: "f1",
                _measurement: "m0",
                t0: "t0v0",
                t1: "t1v0",
                _time: 2021-07-06T23:06:30Z,
                _value: 0,
            },
            {
                _field: "f1",
                _measurement: "m0",
                t0: "t0v0",
                t1: "t1v0",
                _time: 2021-07-06T23:06:40Z,
                _value: 0,
            },
            {
                _field: "f1",
                _measurement: "m0",
                t0: "t0v0",
                t1: "t1v0",
                _time: 2021-07-06T23:06:50Z,
                _value: 0,
            },
            {
                _field: "f1",
                _measurement: "m0",
                t0: "t0v0",
                t1: "t1v1",
                _time: 2021-07-06T23:06:30Z,
                _value: 0,
            },
            {
                _field: "f1",
                _measurement: "m0",
                t0: "t0v0",
                t1: "t1v1",
                _time: 2021-07-06T23:06:40Z,
                _value: 4,
            },
            {
                _field: "f1",
                _measurement: "m0",
                t0: "t0v0",
                t1: "t1v1",
                _time: 2021-07-06T23:06:50Z,
                _value: 3,
            },
            {
                _field: "f1",
                _measurement: "m0",
                t0: "t0v1",
                t1: "t1v0",
                _time: 2021-07-06T23:06:30Z,
                _value: 3,
            },
            {
                _field: "f1",
                _measurement: "m0",
                t0: "t0v1",
                t1: "t1v0",
                _time: 2021-07-06T23:06:40Z,
                _value: 2,
            },
            {
                _field: "f1",
                _measurement: "m0",
                t0: "t0v1",
                t1: "t1v0",
                _time: 2021-07-06T23:06:50Z,
                _value: 1,
            },
            {
                _field: "f1",
                _measurement: "m0",
                t0: "t0v1",
                t1: "t1v1",
                _time: 2021-07-06T23:06:30Z,
                _value: 1,
            },
            {
                _field: "f1",
                _measurement: "m0",
                t0: "t0v1",
                t1: "t1v1",
                _time: 2021-07-06T23:06:40Z,
                _value: 0,
            },
            {
                _field: "f1",
                _measurement: "m0",
                t0: "t0v1",
                t1: "t1v1",
                _time: 2021-07-06T23:06:50Z,
                _value: 2,
            },
        ],
    )
        |> group(columns: ["_measurement", "_field", "t0", "t1"])

// Group + first test
// Group on one tag across fields
testcase group_one_tag_first {
    want =
        array.from(
            rows: [
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v0",
                    "t1": "t1v0",
                    "_value": 3,
                    _time: 2021-07-06T23:06:30Z,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v1",
                    "t1": "t1v0",
                    "_value": 1,
                    _time: 2021-07-06T23:06:30Z,
                },
            ],
        )
            |> group(columns: ["t0"])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> group(columns: ["t0"])
            // Sort to make comparison more reliable when using storage (where
            // row ordering can vary).
            |> sort(columns: ["_time", "_field", "t0", "t1"])
            |> first()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase group_all_filter_field_first {
    want =
        array.from(
            rows: [
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v0",
                    "t1": "t1v0",
                    "_value": 3,
                    _time: 2021-07-06T23:06:30Z,
                },
            ],
        )
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> filter(fn: (r) => r._field == "f0")
            |> group()
            // Sort to make comparison more reliable when using storage (where
            // row ordering can vary).
            |> sort(columns: ["_time", "_field", "t0", "t1"])
            |> first()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase group_one_tag_filter_field_first {
    want =
        array.from(
            rows: [
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v0",
                    "t1": "t1v0",
                    "_value": 3,
                    _time: 2021-07-06T23:06:30Z,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v1",
                    "t1": "t1v0",
                    "_value": 1,
                    _time: 2021-07-06T23:06:30Z,
                },
            ],
        )
            |> group(columns: ["t0"])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> filter(fn: (r) => r._field == "f0")
            |> group(columns: ["t0"])
            // Sort to make comparison more reliable when using storage (where
            // row ordering can vary).
            |> sort(columns: ["_time", "_field", "t0", "t1"])
            |> first()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase group_two_tag_filter_field_first {
    want =
        array.from(
            rows: [
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v0",
                    "t1": "t1v0",
                    "_value": 3,
                    _time: 2021-07-06T23:06:30Z,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v0",
                    "t1": "t1v1",
                    "_value": 4,
                    _time: 2021-07-06T23:06:30Z,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v1",
                    "t1": "t1v0",
                    "_value": 1,
                    _time: 2021-07-06T23:06:30Z,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v1",
                    "t1": "t1v1",
                    "_value": 4,
                    _time: 2021-07-06T23:06:30Z,
                },
            ],
        )
            |> group(columns: ["t0", "t1"])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> filter(fn: (r) => r._field == "f0")
            |> group(columns: ["t0", "t1"])
            |> first()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

// Group + last tests
testcase group_one_tag_last {
    want =
        array.from(
            rows: [
                {
                    _measurement: "m0",
                    _field: "f1",
                    "t0": "t0v0",
                    "t1": "t1v1",
                    _time: 2021-07-06T23:06:50Z,
                    _value: 3,
                },
                {
                    _measurement: "m0",
                    _field: "f1",
                    "t0": "t0v1",
                    "t1": "t1v1",
                    _time: 2021-07-06T23:06:50Z,
                    _value: 2,
                },
            ],
        )
            |> group(columns: ["t0"])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> group(columns: ["t0"])
            // Sort to make comparison more reliable when using storage (where
            // row ordering can vary).
            |> sort(columns: ["_time", "_field", "t0", "t1"])
            |> last()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase group_all_filter_field_last {
    want =
        array.from(
            rows: [
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v1",
                    "t1": "t1v1",
                    _time: 2021-07-06T23:06:50Z,
                    _value: 4,
                },
            ],
        )
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> filter(fn: (r) => r._field == "f0")
            |> group()
            // Sort to make comparison more reliable when using storage (where
            // row ordering can vary).
            |> sort(columns: ["_time", "_field", "t0", "t1"])
            |> last()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase group_one_tag_filter_field_last {
    want =
        array.from(
            rows: [
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v0",
                    "t1": "t1v1",
                    _time: 2021-07-06T23:06:50Z,
                    _value: 1,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v1",
                    "t1": "t1v1",
                    _time: 2021-07-06T23:06:50Z,
                    _value: 4,
                },
            ],
        )
            |> group(columns: ["t0"])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> filter(fn: (r) => r._field == "f0")
            |> group(columns: ["t0"])
            // Sort to make comparison more reliable when using storage (where
            // row ordering can vary).
            |> sort(columns: ["_time", "_field", "t0", "t1"])
            |> last()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase group_two_tag_filter_field_last {
    want =
        array.from(
            rows: [
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v0",
                    "t1": "t1v0",
                    _time: 2021-07-06T23:06:50Z,
                    _value: 0,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v0",
                    "t1": "t1v1",
                    _time: 2021-07-06T23:06:50Z,
                    _value: 1,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v1",
                    "t1": "t1v0",
                    _time: 2021-07-06T23:06:50Z,
                    _value: 4,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    "t0": "t0v1",
                    "t1": "t1v1",
                    _time: 2021-07-06T23:06:50Z,
                    _value: 4,
                },
            ],
        )
            |> group(columns: ["t0", "t1"])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> filter(fn: (r) => r._field == "f0")
            |> group(columns: ["t0", "t1"])
            |> last()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}
