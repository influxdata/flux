package universe_test


import "array"
import "testing"

option now = () => 2030-01-01T00:00:00Z

testcase range_nsecs_bare_last {
    input = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000001Z, _field: "foo", _value: 1.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000002Z, _field: "foo", _value: 2.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000005Z, _field: "foo", _value: 3.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000024Z, _field: "foo", _value: 11.24, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000009Z, _field: "bar", _value: 4.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000011Z, _field: "bar", _value: 5.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000015Z, _field: "bar", _value: 6.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000022Z, _field: "bar", _value: 1.2, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000031Z, _field: "bar", _value: 0.28, _measurement: "mm"},
        ],
    )
        |> group(columns: ["_field", "_measurement"])
        |> testing.load()
    want = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000005Z, _field: "foo", _value: 3.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000011Z, _field: "bar", _value: 5.0, _measurement: "mm"},
        ],
    )
        |> group(columns: ["_field", "_measurement"])
    got = input
        |> range(start: 2021-01-01T00:00:01.000000001Z, stop: 2021-01-01T00:00:01.000000015Z)
        |> last()
        |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}
testcase range_nsecs_bare_max {
    input = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000001Z, _field: "foo", _value: 1.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000002Z, _field: "foo", _value: 2.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000005Z, _field: "foo", _value: 3.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000024Z, _field: "foo", _value: 11.24, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000009Z, _field: "bar", _value: 4.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000011Z, _field: "bar", _value: 5.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000015Z, _field: "bar", _value: 6.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000022Z, _field: "bar", _value: 1.2, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000031Z, _field: "bar", _value: 0.28, _measurement: "mm"},
        ],
    )
        |> group(columns: ["_field", "_measurement"])
        |> testing.load()
    want = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000005Z, _field: "foo", _value: 3.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000015Z, _field: "bar", _value: 6.0, _measurement: "mm"},
        ],
    )
        |> group(columns: ["_field", "_measurement"])
    got = input
        |> range(start: 2021-01-01T00:00:01.000000001Z, stop: 2021-01-01T00:00:01.000000024Z)
        |> max()
        |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}
testcase range_nsecs_group_count {
    input = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000001Z, _field: "foo", _value: 1.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000002Z, _field: "foo", _value: 2.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000005Z, _field: "foo", _value: 3.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000024Z, _field: "foo", _value: 11.24, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000009Z, _field: "bar", _value: 4.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000011Z, _field: "bar", _value: 5.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000015Z, _field: "bar", _value: 6.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000022Z, _field: "bar", _value: 1.2, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000031Z, _field: "bar", _value: 0.28, _measurement: "mm"},
        ],
    )
        |> group(columns: ["_field", "_measurement"])
        |> testing.load()
    want = array.from(
        rows: [
            {_value: 4},
        ],
    )
        |> group(columns: ["_field", "_measurement"])
    got = input
        |> range(start: 2021-01-01T00:00:01.000000001Z, stop: 2021-01-01T00:00:01.000000011Z)
        |> group()
        |> count()
        |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}
testcase range_nsecs_group_max {
    input = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000001Z, _field: "foo", _value: 1.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000002Z, _field: "foo", _value: 2.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000005Z, _field: "foo", _value: 3.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000024Z, _field: "foo", _value: 11.24, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000009Z, _field: "bar", _value: 4.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000011Z, _field: "bar", _value: 5.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000015Z, _field: "bar", _value: 6.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000022Z, _field: "bar", _value: 1.2, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000031Z, _field: "bar", _value: 0.28, _measurement: "mm", section: "1a"},
        ],
    )
        |> group(columns: ["_field", "_measurement"])
        |> testing.load()
    want = array.from(
        rows: [
            {_value: 5.0, _time: 2021-01-01T00:00:01.000000011Z, section: "1a"},
            {_value: 4.0, _time: 2021-01-01T00:00:01.000000009Z, section: "2b"},
        ],
    )
        |> group(columns: ["_field", "_measurement", "section"])
    got = input
        |> range(start: 2021-01-01T00:00:01.000000001Z, stop: 2021-01-01T00:00:01.000000015Z)
        |> group(columns: ["section"])
        |> max()
        |> drop(columns: ["_start", "_stop", "_field", "_measurement"])

    testing.diff(got, want) |> yield()
}
testcase range_nsecs_group_sum {
    input = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000001Z, _field: "foo", _value: 1.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000002Z, _field: "foo", _value: 2.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000005Z, _field: "foo", _value: 3.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000024Z, _field: "foo", _value: 11.24, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000009Z, _field: "bar", _value: 4.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000011Z, _field: "bar", _value: 5.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000015Z, _field: "bar", _value: 6.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000022Z, _field: "bar", _value: 1.2, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000031Z, _field: "bar", _value: 0.28, _measurement: "mm", section: "1a"},
        ],
    )
        |> group(columns: ["_field", "_measurement"])
        |> testing.load()

    want = array.from(
        rows: [
            {_value: 9.0, section: "1a"},
            {_value: 13.2, section: "2b"},
        ],
    )
        |> group(columns: ["_field", "_measurement", "section"])

    got = input
        |> range(start: 2021-01-01T00:00:01.000000001Z, stop: 2021-01-01T00:00:01.000000024Z)
        |> group(columns: ["section"])
        |> sum()
        |> drop(columns: ["_start", "_stop", "_field", "_measurement"])

    testing.diff(got, want) |> yield()
}
testcase range_nsecs_group_last {
    input = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000001Z, _field: "foo", _value: 1.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000002Z, _field: "foo", _value: 2.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000005Z, _field: "foo", _value: 3.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000009Z, _field: "foo", _value: 4.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000011Z, _field: "foo", _value: 5.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000015Z, _field: "foo", _value: 6.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000022Z, _field: "foo", _value: 1.2, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000024Z, _field: "foo", _value: 11.24, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000031Z, _field: "foo", _value: 0.28, _measurement: "mm", section: "1a"},
        ],
    )
        |> group(columns: ["_field", "_measurement", "section"])
        |> testing.load()

    want = array.from(
        rows: [
            {_value: 5.0, _time: 2021-01-01T00:00:01.000000011Z, section: "1a", _field: "foo"},
            {_value: 1.2, _time: 2021-01-01T00:00:01.000000022Z, section: "2b", _field: "foo"},
        ],
    )
        |> group(columns: ["_field", "_measurement", "section"])

    got = input
        |> range(start: 2021-01-01T00:00:01.000000001Z, stop: 2021-01-01T00:00:01.000000024Z)
        |> group(columns: ["_field", "section"])
        |> last(column: "_value")
        |> drop(columns: ["_start", "_stop", "_measurement"])

    testing.diff(got, want) |> yield()
}
testcase range_nsecs_window_sum {
    input = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000001Z, _field: "foo", _value: 1.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000002Z, _field: "foo", _value: 2.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000005Z, _field: "foo", _value: 3.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000009Z, _field: "foo", _value: 4.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000011Z, _field: "foo", _value: 5.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000015Z, _field: "foo", _value: 6.0, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000022Z, _field: "foo", _value: 1.2, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000024Z, _field: "foo", _value: 11.24, _measurement: "mm"},
            {_time: 2021-01-01T00:00:01.000000031Z, _field: "foo", _value: 0.28, _measurement: "mm"},
        ],
    )
        |> group(columns: ["_field", "_measurement"])
        |> testing.load()

    want = array.from(
        rows: [
            {_start: 2021-01-01T00:00:01.000000005Z, _stop: 2021-01-01T00:00:01.00000001Z, _value: 7.0},
            {_start: 2021-01-01T00:00:01.00000001Z, _stop: 2021-01-01T00:00:01.00000002Z, _value: 11.0},
            {_start: 2021-01-01T00:00:01.00000002Z, _stop: 2021-01-01T00:00:01.00000003Z, _value: 12.44},
        ],
    )
        |> group(columns: ["_field", "_measurement", "_start", "_stop"])

    got = input
        |> range(start: 2021-01-01T00:00:01.000000005Z, stop: 2021-01-01T00:00:01.000000031Z)
        |> window(every: 10ns)
        |> sum()
        // removed _start and _stop columns due group key issues with array
        |> drop(columns: ["_field", "_measurement"])

    testing.diff(got, want) |> yield()
}
testcase range_nsecs_window_first {
    input = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000001Z, _field: "foo", _value: 1.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000002Z, _field: "foo", _value: 2.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000005Z, _field: "foo", _value: 3.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000009Z, _field: "foo", _value: 4.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000011Z, _field: "foo", _value: 5.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000015Z, _field: "foo", _value: 6.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000022Z, _field: "foo", _value: 1.2, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000024Z, _field: "foo", _value: 11.24, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000031Z, _field: "foo", _value: 0.28, _measurement: "mm", section: "1a"},
        ],
    )
        |> group(columns: ["_field", "_measurement", "section"])
        |> testing.load()

    want = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000001Z, _value: 1.0, section: "1a"},
            {_time: 2021-01-01T00:00:01.000000011Z, _value: 5.0, section: "1a"},
            {_time: 2021-01-01T00:00:01.000000022Z, _value: 1.2, section: "1a"},
        ],
    )
        |> group(columns: ["_field", "_measurement", "section"])

    got = input
        |> range(start: 2021-01-01T00:00:01.000000001Z, stop: 2021-01-01T00:00:01.000000031Z)
        |> window(every: 10ns)
        |> first()
        |> drop(columns: ["_start", "_stop", "_field", "_measurement"])

    testing.diff(got, want) |> yield()
}
testcase range_nsecs_window_min {
    input = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000001Z, _field: "foo", _value: 1.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000002Z, _field: "foo", _value: 2.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000005Z, _field: "foo", _value: 3.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000009Z, _field: "bar", _value: 4.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000011Z, _field: "bar", _value: 5.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000015Z, _field: "bar", _value: 6.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000022Z, _field: "bar", _value: 1.2, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000024Z, _field: "foo", _value: 11.24, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000031Z, _field: "bar", _value: 0.28, _measurement: "mm", section: "1a"},
        ],
    )
        |> group(columns: ["_field", "_measurement", "section"])
        |> testing.load()

    want = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000001Z, _value: 1.0, section: "1a", _field: "foo"},
            {_time: 2021-01-01T00:00:01.000000024Z, _value: 11.24, section: "1a", _field: "foo"},
            {_time: 2021-01-01T00:00:01.000000011Z, _value: 5.0, section: "1a", _field: "bar"},
            {_time: 2021-01-01T00:00:01.000000002Z, _value: 2.0, section: "2b", _field: "foo"},
            {_time: 2021-01-01T00:00:01.000000009Z, _value: 4.0, section: "2b", _field: "bar"},
            {_time: 2021-01-01T00:00:01.000000015Z, _value: 6.0, section: "2b", _field: "bar"},
            {_time: 2021-01-01T00:00:01.000000022Z, _value: 1.2, section: "2b", _field: "bar"},
        ],
    )
        |> group(columns: ["_field", "_measurement", "section"])

    got = input
        |> range(start: 2021-01-01T00:00:01.000000001Z, stop: 2021-01-01T00:00:01.000000031Z)
        |> window(every: 10ns)
        |> min()
        |> drop(columns: ["_start", "_stop", "_measurement"])

    testing.diff(got, want) |> yield()
}
testcase range_nsecs_agg_count {
    input = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000001Z, _field: "foo", _value: 1.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000002Z, _field: "foo", _value: 2.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000005Z, _field: "foo", _value: 3.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000009Z, _field: "bar", _value: 4.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000011Z, _field: "bar", _value: 5.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000015Z, _field: "bar", _value: 6.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000022Z, _field: "bar", _value: 1.2, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000024Z, _field: "foo", _value: 11.24, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000031Z, _field: "bar", _value: 0.28, _measurement: "mm", section: "1a"},
        ],
    )
        |> group(columns: ["_field", "_measurement", "section"])
        |> testing.load()

    want = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000005Z, _value: 1, _field: "foo", section: "1a"},
            {_time: 2021-01-01T00:00:01.00000002Z, _value: 1, _field: "foo", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000035Z, _value: 1, _field: "foo", section: "1a"},
            {_time: 2021-01-01T00:00:01.00000005Z, _value: 0, _field: "foo", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000005Z, _value: 0, _field: "bar", section: "1a"},
            {_time: 2021-01-01T00:00:01.00000002Z, _value: 1, _field: "bar", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000035Z, _value: 1, _field: "bar", section: "1a"},
            {_time: 2021-01-01T00:00:01.00000005Z, _value: 0, _field: "bar", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000005Z, _value: 1, _field: "foo", section: "2b"},
            {_time: 2021-01-01T00:00:01.00000002Z, _value: 0, _field: "foo", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000035Z, _value: 0, _field: "foo", section: "2b"},
            {_time: 2021-01-01T00:00:01.00000005Z, _value: 0, _field: "foo", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000005Z, _value: 0, _field: "bar", section: "2b"},
            {_time: 2021-01-01T00:00:01.00000002Z, _value: 2, _field: "bar", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000035Z, _value: 1, _field: "bar", section: "2b"},
            {_time: 2021-01-01T00:00:01.00000005Z, _value: 0, _field: "bar", section: "2b"},
        ],
    )
        |> group(columns: ["_field", "section"])

    got = input
        |> range(start: 2021-01-01T00:00:01Z, stop: 2021-01-01T00:00:01.00000005Z)
        |> aggregateWindow(every: 15ns, fn: count)
        |> drop(
            columns: [
                "_start",
                "_stop",
                "_measurement",
            ],
        )

    testing.diff(got, want) |> yield()
}
testcase range_nsecs_agg_last {
    input = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000001Z, _field: "foo", _value: 1.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000002Z, _field: "foo", _value: 2.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000005Z, _field: "foo", _value: 3.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000009Z, _field: "bar", _value: 4.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000011Z, _field: "bar", _value: 5.0, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000015Z, _field: "bar", _value: 6.0, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000022Z, _field: "bar", _value: 1.2, _measurement: "mm", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000024Z, _field: "foo", _value: 11.24, _measurement: "mm", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000031Z, _field: "bar", _value: 0.28, _measurement: "mm", section: "1a"},
        ],
    )
        |> group(columns: ["_field", "_measurement", "section"])
        |> testing.load()

    want = array.from(
        rows: [
            {_time: 2021-01-01T00:00:01.000000009Z, _value: 3.0, _field: "foo", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000014Z, _value: 5.0, _field: "bar", section: "1a"},
            {_time: 2021-01-01T00:00:01.000000014Z, _value: 4.0, _field: "bar", section: "2b"},
            {_time: 2021-01-01T00:00:01.000000019Z, _value: 6.0, _field: "bar", section: "2b"},
        ],
    )
        |> group(columns: ["_field", "_measurement", "section"])

    got = input
        |> range(start: 2021-01-01T00:00:01.000000005Z, stop: 2021-01-01T00:00:01.00000002Z)
        |> aggregateWindow(every: 5ns, offset: -1ns, fn: last)
        |> drop(
            columns: [
                "_start",
                "_stop",
                "_measurement",
            ],
        )

    testing.diff(got, want) |> yield()
}
