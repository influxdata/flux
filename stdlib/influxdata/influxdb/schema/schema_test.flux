package schema_test


import "internal/debug"
import "array"
import "testing"
import "csv"
import "influxdata/influxdb/schema"

// Range of the input data
start = 2018-05-22T19:53:16Z
stop = 2018-05-22T19:53:26Z

// cpu:
//      tags: host,cpu
//      fields: usage_idle,usage_user
cpu =
    array.from(
        rows: [
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu0",
                _field: "usage_idle",
                _value: 1.83,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu1",
                _field: "usage_idle",
                _value: 1.72,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu0",
                _field: "usage_user",
                _value: 1.83,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu1",
                _field: "usage_user",
                _value: 1.72,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu0",
                _field: "usage_idle",
                _value: 1.83,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu1",
                _field: "usage_idle",
                _value: 1.72,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu0",
                _field: "usage_user",
                _value: 1.83,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu1",
                _field: "usage_user",
                _value: 1.72,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu0",
                _field: "usage_idle",
                _value: 1.98,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu1",
                _field: "usage_idle",
                _value: 1.97,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu0",
                _field: "usage_user",
                _value: 1.98,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu1",
                _field: "usage_user",
                _value: 1.97,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu0",
                _field: "usage_idle",
                _value: 1.98,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu1",
                _field: "usage_idle",
                _value: 1.97,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu0",
                _field: "usage_user",
                _value: 1.98,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu1",
                _field: "usage_user",
                _value: 1.97,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu0",
                _field: "usage_idle",
                _value: 1.95,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu1",
                _field: "usage_idle",
                _value: 1.92,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu0",
                _field: "usage_user",
                _value: 1.95,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu1",
                _field: "usage_user",
                _value: 1.92,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu0",
                _field: "usage_idle",
                _value: 1.95,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu1",
                _field: "usage_idle",
                _value: 1.92,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu0",
                _field: "usage_user",
                _value: 1.95,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu1",
                _field: "usage_user",
                _value: 1.92,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu0",
                _field: "usage_idle",
                _value: 1.83,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu1",
                _field: "usage_idle",
                _value: 1.72,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu0",
                _field: "usage_user",
                _value: 1.83,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu1",
                _field: "usage_user",
                _value: 1.72,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu0",
                _field: "usage_idle",
                _value: 1.83,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu1",
                _field: "usage_idle",
                _value: 1.72,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu0",
                _field: "usage_user",
                _value: 1.83,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu1",
                _field: "usage_user",
                _value: 1.72,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu0",
                _field: "usage_idle",
                _value: 1.98,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu1",
                _field: "usage_idle",
                _value: 1.97,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu0",
                _field: "usage_user",
                _value: 1.98,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu1",
                _field: "usage_user",
                _value: 1.97,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu0",
                _field: "usage_idle",
                _value: 1.98,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu1",
                _field: "usage_idle",
                _value: 1.97,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu0",
                _field: "usage_user",
                _value: 1.98,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu1",
                _field: "usage_user",
                _value: 1.97,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu0",
                _field: "usage_idle",
                _value: 1.95,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu1",
                _field: "usage_idle",
                _value: 1.92,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu0",
                _field: "usage_user",
                _value: 1.95,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.local",
                cpu: "cpu1",
                _field: "usage_user",
                _value: 1.92,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu0",
                _field: "usage_idle",
                _value: 1.95,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu1",
                _field: "usage_idle",
                _value: 1.92,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu0",
                _field: "usage_user",
                _value: 1.95,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "cpu",
                host: "host.global",
                cpu: "cpu1",
                _field: "usage_user",
                _value: 1.92,
            },
        ],
    )
        |> group(columns: ["_measurement", "_field", "host", "cpu"])

// swap:
//      tags: host,partition
//      fields: used_percent
swap =
    array.from(
        rows: [
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "swap",
                host: "host.local",
                partition: "/dev/sda1",
                _field: "used_percent",
                _value: 66.98,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "swap",
                host: "host.local",
                partition: "/dev/sdb1",
                _field: "used_percent",
                _value: 66.98,
            },
            {
                _time: 2018-05-22T19:53:26Z,
                _measurement: "swap",
                host: "host.local",
                partition: "/dev/sda1",
                _field: "used_percent",
                _value: 37.59,
            },
            {
                _time: 2018-05-22T19:53:26Z,
                _measurement: "swap",
                host: "host.local",
                partition: "/dev/sdb1",
                _field: "used_percent",
                _value: 37.59,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "swap",
                host: "host.global",
                partition: "/dev/sda1",
                _field: "used_percent",
                _value: 72.98,
            },
            {
                _time: 2018-05-22T19:53:16Z,
                _measurement: "swap",
                host: "host.global",
                partition: "/dev/sdb1",
                _field: "used_percent",
                _value: 72.98,
            },
            {
                _time: 2018-05-22T19:53:26Z,
                _measurement: "swap",
                host: "host.global",
                partition: "/dev/sda1",
                _field: "used_percent",
                _value: 88.59,
            },
            {
                _time: 2018-05-22T19:53:26Z,
                _measurement: "swap",
                host: "host.global",
                partition: "/dev/sdb1",
                _field: "used_percent",
                _value: 88.59,
            },
        ],
    )
        |> group(columns: ["_measurement", "_field", "host", "partition"])

option schema._from = (bucket) =>
    union(
        tables: [
            cpu
                |> debug.opaque(),
            swap
                |> debug.opaque(),
        ],
    )
        |> testing.load()

testcase tagValues {
    want = array.from(rows: [{_value: "host.global"}, {_value: "host.local"}])

    got =
        schema.tagValues(bucket: "bucket", tag: "host", start: start, stop: stop)
            |> sort()

    testing.diff(want: want, got: got) |> yield()
}
testcase tagValues_with_predicate {
    want = array.from(rows: [{_value: "cpu0"}, {_value: "cpu1"}])

    got =
        schema.tagValues(
            bucket: "bucket",
            tag: "cpu",
            predicate: (r) => r._measurement == "cpu",
            start: start,
            stop: stop,
        )
            |> sort()

    testing.diff(want: want, got: got) |> yield()
}
testcase tagValues_empty {
    want =
        array.from(rows: [{_value: "foo"}])
            |> filter(fn: (r) => false)

    got =
        schema.tagValues(
            bucket: "bucket",
            tag: "cpu",
            // Predicate will filter out everything.
            // The cpu measurement doesn't have the used_percent field so we get nothing back.
            // This is a regression test for a panic that would occur here.
            predicate: (r) => r._measurement == "cpu" and r._field == "used_percent",
            start: start,
            stop: stop,
        )
            |> sort()

    testing.diff(want: want, got: got) |> yield()
}
testcase measurementTagValues {
    want = array.from(rows: [{_value: "cpu0"}, {_value: "cpu1"}])

    got =
        schema.measurementTagValues(
            bucket: "bucket",
            tag: "cpu",
            measurement: "cpu",
            start: start,
            stop: stop,
        )
            |> sort()

    testing.diff(want: want, got: got) |> yield()
}
testcase tagKeys {
    want =
        array.from(
            rows: [
                {_value: "_field"},
                {_value: "_measurement"},
                {_value: "cpu"},
                {_value: "host"},
                {_value: "partition"},
            ],
        )

    got =
        schema.tagKeys(bucket: "bucket", start: start, stop: stop)
            |> filter(fn: (r) => r._value != "_start" and r._value != "_stop")
            |> sort()

    testing.diff(want: want, got: got) |> yield()
}
testcase tagKeys_with_predicate {
    want =
        array.from(
            rows: [{_value: "_field"}, {_value: "_measurement"}, {_value: "cpu"}, {_value: "host"}],
        )

    got =
        schema.tagKeys(
            bucket: "bucket",
            predicate: (r) => r._measurement == "cpu",
            start: start,
            stop: stop,
        )
            |> filter(fn: (r) => r._value != "_start" and r._value != "_stop")
            |> sort()

    testing.diff(want: want, got: got) |> yield()
}
testcase measurementTagKeys {
    want =
        array.from(
            rows: [{_value: "_field"}, {_value: "_measurement"}, {_value: "cpu"}, {_value: "host"}],
        )

    got =
        schema.measurementTagKeys(bucket: "bucket", measurement: "cpu", start: start, stop: stop)
            |> filter(fn: (r) => r._value != "_start" and r._value != "_stop")
            |> sort()

    testing.diff(want: want, got: got) |> yield()
}
testcase fieldKeys {
    want =
        array.from(rows: [{_value: "usage_idle"}, {_value: "usage_user"}, {_value: "used_percent"}])

    got =
        schema.fieldKeys(bucket: "bucket", start: start, stop: stop)
            |> sort()

    testing.diff(want: want, got: got) |> yield()
}

testcase fieldKeys_with_predicate {
    want = array.from(rows: [{_value: "usage_idle"}, {_value: "usage_user"}])

    got =
        schema.fieldKeys(
            bucket: "bucket",
            predicate: (r) => r._measurement == "cpu" and r.host == "host.local",
            start: start,
            stop: stop,
        )
            |> sort()

    testing.diff(want: want, got: got) |> yield()
}
testcase measurementFieldKeys {
    want = array.from(rows: [{_value: "usage_idle"}, {_value: "usage_user"}])

    got =
        schema.measurementFieldKeys(bucket: "bucket", measurement: "cpu", start: start, stop: stop)
            |> sort()

    testing.diff(want: want, got: got) |> yield()
}
testcase measurements {
    want = array.from(rows: [{_value: "cpu"}, {_value: "swap"}])

    got =
        schema.measurements(bucket: "bucket", start: start, stop: stop)
            |> sort()

    testing.diff(want: want, got: got) |> yield()
}
