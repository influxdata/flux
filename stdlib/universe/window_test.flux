package universe_test


import "array"
import "testing"

input = () => array.from(
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

testcase window_period_gaps {
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
    got = input()
        |> window(every: 10s, period: 5s)

    testing.diff(got, want) |> yield()
}
