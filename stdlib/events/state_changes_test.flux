package events_test


import "testing"
import "experimental/array"
import "events"

data = array.from(
    rows: [
        {_time: 2020-10-01T00:00:00Z, _measurement: "foo", _field: "temp", state: "x", _value: 10},
        {_time: 2020-10-01T00:00:01Z, _measurement: "foo", _field: "temp", state: "x", _value: 10},
        {_time: 2020-10-01T00:00:02Z, _measurement: "foo", _field: "temp", state: "x", _value: 11},
        {_time: 2020-10-01T00:00:03Z, _measurement: "foo", _field: "temp", state: "y", _value: 12},
        {_time: 2020-10-01T00:00:04Z, _measurement: "foo", _field: "temp", state: "y", _value: 13},
        {_time: 2020-10-01T00:00:04Z, _measurement: "foo", _field: "temp", state: "y", _value: 13},
        {_time: 2020-10-01T00:00:05Z, _measurement: "foo", _field: "temp", state: "z", _value: 14},
        {_time: 2020-10-01T00:00:06Z, _measurement: "foo", _field: "temp", state: "x", _value: 15},
        {_time: 2020-10-01T00:00:07Z, _measurement: "foo", _field: "temp", state: "z", _value: 16},
    ],
)
want = array.from(
    rows: [
        {_time: 2020-10-01T00:00:00Z, _measurement: "foo", _field: "temp", state: "x", _value: 10},
        {_time: 2020-10-01T00:00:03Z, _measurement: "foo", _field: "temp", state: "y", _value: 12},
        {_time: 2020-10-01T00:00:05Z, _measurement: "foo", _field: "temp", state: "z", _value: 14},
        {_time: 2020-10-01T00:00:06Z, _measurement: "foo", _field: "temp", state: "x", _value: 15},
        {_time: 2020-10-01T00:00:07Z, _measurement: "foo", _field: "temp", state: "z", _value: 16},
    ],
)

test stateChanges = () => ({
    input: data |> testing.load(),
    want: want,
    fn: (tables=<-) => tables
        |> events.stateChanges()
        |> drop(columns: ["_start", "_stop"]),
})
