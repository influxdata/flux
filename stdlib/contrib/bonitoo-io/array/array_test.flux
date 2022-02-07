package array_test


import "array"
import barray "contrib/bonitoo-io/array"
import "testing"

testcase array_concat {
    input =
        array.from(
            rows: [
                {
                    _time: 2018-05-22T19:53:26Z,
                    _value: 15204688,
                    _field: "io_time",
                    _measurement: "diskio",
                    host: "host.local",
                    name: "disk0",
                },
                {
                    _time: 2018-05-22T19:53:36Z,
                    _value: 15204894,
                    _field: "io_time",
                    _measurement: "diskio",
                    host: "host.local",
                    name: "disk0",
                },
                {
                    _time: 2018-05-22T19:53:46Z,
                    _value: 15205102,
                    _field: "io_time",
                    _measurement: "diskio",
                    host: "host.local",
                    name: "disk0",
                },
            ],
        )
            |> group(columns: ["_measurement", "_field", "host", "name"])
            |> testing.load()
            |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:54:00Z)
    want =
        array.from(
            rows: [
                {_measurement: "diskio", _time: 2018-05-22T19:53:26Z, _value: 15204688},
                {_measurement: "diskio", _time: 2018-05-22T19:53:36Z, _value: 15204894},
                {_measurement: "diskio", _time: 2018-05-22T19:53:46Z, _value: 15205102},
            ],
        )
            |> group(columns: ["_measurement"])

    cols = ["_measurement"]
    got =
        input
            |> keep(columns: barray.concat(arr: ["_time", "_value"], v: cols))

    testing.diff(got, want) |> yield()
}
testcase array_concat_empty {
    input =
        array.from(
            rows: [
                {
                    _time: 2018-05-22T19:53:26Z,
                    _value: 15204688,
                    _field: "io_time",
                    _measurement: "diskio",
                    host: "host.local",
                    name: "disk0",
                },
                {
                    _time: 2018-05-22T19:53:36Z,
                    _value: 15204894,
                    _field: "io_time",
                    _measurement: "diskio",
                    host: "host.local",
                    name: "disk0",
                },
                {
                    _time: 2018-05-22T19:53:46Z,
                    _value: 15205102,
                    _field: "io_time",
                    _measurement: "diskio",
                    host: "host.local",
                    name: "disk0",
                },
            ],
        )
            |> group(columns: ["_measurement", "_field", "host", "name"])
            |> testing.load()
            |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:54:00Z)
    want =
        array.from(
            rows: [
                {_measurement: "diskio", _time: 2018-05-22T19:53:26Z, _value: 15204688},
                {_measurement: "diskio", _time: 2018-05-22T19:53:36Z, _value: 15204894},
                {_measurement: "diskio", _time: 2018-05-22T19:53:46Z, _value: 15205102},
            ],
        )
            |> group(columns: ["_measurement"])

    cols = ["_measurement", "_time", "_value"]
    got =
        input
            |> keep(columns: barray.concat(arr: cols, v: []))

    testing.diff(got, want) |> yield()
}
testcase array_concat_to_empty {
    input =
        array.from(
            rows: [
                {
                    _time: 2018-05-22T19:53:26Z,
                    _value: 15204688,
                    _field: "io_time",
                    _measurement: "diskio",
                    host: "host.local",
                    name: "disk0",
                },
                {
                    _time: 2018-05-22T19:53:36Z,
                    _value: 15204894,
                    _field: "io_time",
                    _measurement: "diskio",
                    host: "host.local",
                    name: "disk0",
                },
                {
                    _time: 2018-05-22T19:53:46Z,
                    _value: 15205102,
                    _field: "io_time",
                    _measurement: "diskio",
                    host: "host.local",
                    name: "disk0",
                },
            ],
        )
            |> group(columns: ["_measurement", "_field", "host", "name"])
            |> testing.load()
            |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:54:00Z)
    want =
        array.from(
            rows: [
                {_measurement: "diskio", _time: 2018-05-22T19:53:26Z, _value: 15204688},
                {_measurement: "diskio", _time: 2018-05-22T19:53:36Z, _value: 15204894},
                {_measurement: "diskio", _time: 2018-05-22T19:53:46Z, _value: 15205102},
            ],
        )
            |> group(columns: ["_measurement"])

    cols = ["_measurement", "_time", "_value"]
    got =
        input
            |> keep(columns: barray.concat(arr: [], v: cols))

    testing.diff(got, want) |> yield()
}
testcase array_map {
    input =
        array.from(
            rows: [
                {
                    _time: 2018-05-22T19:53:26Z,
                    _value: 15204688,
                    _field: "io_time",
                    _measurement: "diskio",
                    host: "host.local",
                    name: "disk0",
                },
                {
                    _time: 2018-05-22T19:53:36Z,
                    _value: 15204894,
                    _field: "io_time",
                    _measurement: "diskio",
                    host: "host.local",
                    name: "disk0",
                },
                {
                    _time: 2018-05-22T19:53:46Z,
                    _value: 15205102,
                    _field: "io_time",
                    _measurement: "diskio",
                    host: "host.local",
                    name: "disk0",
                },
                {
                    _time: 2018-05-22T19:53:56Z,
                    _value: -1,
                    _field: "io_time",
                    _measurement: "diskio",
                    host: "host.local",
                    name: "disk0",
                },
            ],
        )
            |> group(columns: ["_measurement", "_field", "host", "name"])
            |> testing.load()
            |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:54:00Z)
    want =
        array.from(
            rows: [
                {_measurement: "diskio", _time: 2018-05-22T19:53:26Z, _value: 15204688},
                {_measurement: "diskio", _time: 2018-05-22T19:53:36Z, _value: 15204894},
                {_measurement: "diskio", _time: 2018-05-22T19:53:46Z, _value: 15205102},
            ],
        )
            |> group(columns: ["_measurement"])

    invalids = ["-1"]
    fx = (x) => int(v: x)

    got =
        input
            |> filter(fn: (r) => not contains(value: r._value, set: barray.map(arr: invalids, fn: fx)))
            |> keep(columns: ["_measurement", "_time", "_value"])

    testing.diff(got, want) |> yield()
}
