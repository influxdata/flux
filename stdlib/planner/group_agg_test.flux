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

// Group + count test
// Group on one tag across fields
testcase group_one_tag_count {
    want =
        array.from(rows: [{"t0": "t0v0", "_value": 12}, {"t0": "t0v1", "_value": 12}])
            |> group(columns: ["t0"])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> group(columns: ["t0"])
            |> count()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want)
}

testcase group_all_filter_field_count {
    want = array.from(rows: [{"_value": 12}])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> filter(fn: (r) => r._field == "f0")
            |> group()
            |> count()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase group_one_tag_filter_field_count {
    want =
        array.from(rows: [{"t0": "t0v0", "_value": 6}, {"t0": "t0v1", "_value": 6}])
            |> group(columns: ["t0"])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> filter(fn: (r) => r._field == "f0")
            |> group(columns: ["t0"])
            |> count()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase group_two_tag_filter_field_count {
    want =
        array.from(
            rows: [
                {"t0": "t0v0", "t1": "t1v0", "_value": 3},
                {"t0": "t0v0", "t1": "t1v1", "_value": 3},
                {"t0": "t0v1", "t1": "t1v0", "_value": 3},
                {"t0": "t0v1", "t1": "t1v1", "_value": 3},
            ],
        )
            |> group(columns: ["t0", "t1"])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> filter(fn: (r) => r._field == "f0")
            |> group(columns: ["t0", "t1"])
            |> count()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

// Group + sum tests
testcase group_one_tag_sum {
    want =
        array.from(rows: [{"t0": "t0v0", "_value": 19}, {"t0": "t0v1", "_value": 22}])
            |> group(columns: ["t0"])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> group(columns: ["t0"])
            |> sum()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase group_all_filter_field_sum {
    want = array.from(rows: [{"_value": 25}])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> filter(fn: (r) => r._field == "f0")
            |> group()
            |> sum()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase group_one_tag_filter_field_sum {
    want =
        array.from(rows: [{"t0": "t0v0", "_value": 12}, {"t0": "t0v1", "_value": 13}])
            |> group(columns: ["t0"])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> filter(fn: (r) => r._field == "f0")
            |> group(columns: ["t0"])
            |> sum()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase group_two_tag_filter_field_sum {
    want =
        array.from(
            rows: [
                {"t0": "t0v0", "t1": "t1v0", "_value": 4},
                {"t0": "t0v0", "t1": "t1v1", "_value": 8},
                {"t0": "t0v1", "t1": "t1v0", "_value": 5},
                {"t0": "t0v1", "t1": "t1v1", "_value": 8},
            ],
        )
            |> group(columns: ["t0", "t1"])
    got =
        testing.load(tables: inData)
            |> range(start: -100y)
            |> filter(fn: (r) => r._field == "f0")
            |> group(columns: ["t0", "t1"])
            |> sum()
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}
