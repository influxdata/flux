package universe_test


import "array"
import "testing"

testcase window_period_gaps {
    input = array.from(
        rows: [
            {_time: 2018-05-22T19:53:00Z, _measurement: "diskio", _field: "io_time", host: "host.local", name: "disk0", _value: 15204688},
            {_time: 2018-05-22T19:53:03Z, _measurement: "diskio", _field: "io_time", host: "host.local", name: "disk0", _value: 15204894},
            {_time: 2018-05-22T19:53:06Z, _measurement: "diskio", _field: "io_time", host: "host.local", name: "disk0", _value: 15205102},
            {_time: 2018-05-22T19:53:09Z, _measurement: "diskio", _field: "io_time", host: "host.local", name: "disk0", _value: 15205226},
            {_time: 2018-05-22T19:53:12Z, _measurement: "diskio", _field: "io_time", host: "host.local", name: "disk0", _value: 15205499},
            {_time: 2018-05-22T19:53:15Z, _measurement: "diskio", _field: "io_time", host: "host.local", name: "disk0", _value: 15205755},
            {_time: 2018-05-22T19:53:18Z, _measurement: "diskio", _field: "io_time", host: "host.local", name: "disk0", _value: 16205923},
        ],
    )
        |> group(columns: ["_measurement", "_field", "host", "name"])
        |> testing.load()
        |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:53:30Z)
    want = array.from(
        rows: [
            {_time: 2018-05-22T19:53:00Z, _measurement: "diskio", _field: "io_time", _start: 2018-05-22T19:53:00Z, _stop: 2018-05-22T19:53:05Z, host: "host.local", name: "disk0", _value: 15204688},
            {_time: 2018-05-22T19:53:03Z, _measurement: "diskio", _field: "io_time", _start: 2018-05-22T19:53:00Z, _stop: 2018-05-22T19:53:05Z, host: "host.local", name: "disk0", _value: 15204894},
            {_time: 2018-05-22T19:53:12Z, _measurement: "diskio", _field: "io_time", _start: 2018-05-22T19:53:10Z, _stop: 2018-05-22T19:53:15Z, host: "host.local", name: "disk0", _value: 15205499},
        ],
    )
        |> group(
            columns: [
                "_measurement",
                "_field",
                "_start",
                "_stop",
                "host",
                "name",
            ],
        )
    got = input
        |> window(every: 10s, period: 5s)

    testing.diff(got, want) |> yield()
}
testcase window_offset {
    option now = () => 2030-01-01T00:00:00Z

    input = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: 15204688, _field: "io_time", _measurement: "diskio", host: "host.local", name: "disk0"},
            {_time: 2018-05-22T19:53:36Z, _value: 15204894, _field: "io_time", _measurement: "diskio", host: "host.local", name: "disk0"},
            {_time: 2018-05-22T19:53:46Z, _value: 15205102, _field: "io_time", _measurement: "diskio", host: "host.local", name: "disk0"},
            {_time: 2018-05-22T19:53:56Z, _value: 15205226, _field: "io_time", _measurement: "diskio", host: "host.local", name: "disk0"},
            {_time: 2018-05-22T19:54:06Z, _value: 15205499, _field: "io_time", _measurement: "diskio", host: "host.local", name: "disk0"},
            {_time: 2018-05-22T19:54:16Z, _value: 15205755, _field: "io_time", _measurement: "diskio", host: "host.local", name: "disk0"},
            {_time: 2018-05-22T19:53:26Z, _value: 648, _field: "io_time", _measurement: "diskio", host: "host.local", name: "disk2"},
            {_time: 2018-05-22T19:53:36Z, _value: 648, _field: "io_time", _measurement: "diskio", host: "host.local", name: "disk2"},
            {_time: 2018-05-22T19:53:46Z, _value: 648, _field: "io_time", _measurement: "diskio", host: "host.local", name: "disk2"},
            {_time: 2018-05-22T19:53:56Z, _value: 648, _field: "io_time", _measurement: "diskio", host: "host.local", name: "disk2"},
            {_time: 2018-05-22T19:54:06Z, _value: 648, _field: "io_time", _measurement: "diskio", host: "host.local", name: "disk2"},
            {_time: 2018-05-22T19:54:16Z, _value: 648, _field: "io_time", _measurement: "diskio", host: "host.local", name: "disk2"},
        ],
    )
        |> group(columns: ["_field", "_measurement", "host", "name"])
        |> testing.load()
        |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:55:00Z)
        |> group(columns: ["_measurement"])
    want = array.from(
        rows: [
            {_start: 2018-05-22T19:53:00Z, _stop: 2018-05-22T19:55:00Z, _measurement: "diskio", _time: 2018-05-22T19:53:26Z, _value: 7602668.0},
            {_start: 2018-05-22T19:53:00Z, _stop: 2018-05-22T19:55:00Z, _measurement: "diskio", _time: 2018-05-22T19:53:36Z, _value: 7602771.0},
            {_start: 2018-05-22T19:53:00Z, _stop: 2018-05-22T19:55:00Z, _measurement: "diskio", _time: 2018-05-22T19:53:46Z, _value: 7602875.0},
            {_start: 2018-05-22T19:53:00Z, _stop: 2018-05-22T19:55:00Z, _measurement: "diskio", _time: 2018-05-22T19:53:56Z, _value: 7602937.0},
            {_start: 2018-05-22T19:53:00Z, _stop: 2018-05-22T19:55:00Z, _measurement: "diskio", _time: 2018-05-22T19:54:06Z, _value: 7603073.5},
            {_start: 2018-05-22T19:53:00Z, _stop: 2018-05-22T19:55:00Z, _measurement: "diskio", _time: 2018-05-22T19:54:16Z, _value: 7603201.5},
        ],
    )
        |> group(columns: ["_measurement", "_start", "_stop"])
    got = input
        |> window(every: 1s, offset: 2s)
        |> mean()
        |> duplicate(column: "_start", as: "_time")
        |> window(every: inf)

    testing.diff(got, want) |> yield()
}
testcase window_negative_offset {
    input = array.from(
        rows: [
            {_time: 2020-04-10T12:00:00Z, _value: 1, _field: "data", _measurement: "test1"},
            {_time: 2020-04-20T12:00:00Z, _value: 2, _field: "data", _measurement: "test1"},
            {_time: 2020-05-10T12:00:00Z, _value: 3, _field: "data", _measurement: "test1"},
            {_time: 2020-05-20T12:00:00Z, _value: 4, _field: "data", _measurement: "test1"},
        ],
    )
        |> group(columns: ["_measurement", "_field"])
        |> testing.load()
        |> range(start: 2020-03-15T00:00:00Z, stop: 2020-06-01T15:02:25Z)
    want = array.from(
        rows: [
            {_start: 2020-03-31T22:00:00Z, _stop: 2020-04-30T22:00:00Z, _time: 2020-04-10T12:00:00Z, _value: 1, _field: "data", _measurement: "test1"},
            {_start: 2020-03-31T22:00:00Z, _stop: 2020-04-30T22:00:00Z, _time: 2020-04-20T12:00:00Z, _value: 2, _field: "data", _measurement: "test1"},
            {_start: 2020-04-30T22:00:00Z, _stop: 2020-05-30T22:00:00Z, _time: 2020-05-10T12:00:00Z, _value: 3, _field: "data", _measurement: "test1"},
            {_start: 2020-04-30T22:00:00Z, _stop: 2020-05-30T22:00:00Z, _time: 2020-05-20T12:00:00Z, _value: 4, _field: "data", _measurement: "test1"},
        ],
    )
        |> group(
            columns: [
                "_measurement",
                "_field",
                "_start",
                "_stop",
            ],
        )
    got = input
        |> window(period: 1mo, offset: -2h)

    testing.diff(got, want) |> yield()
}
