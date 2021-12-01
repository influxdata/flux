package universe_test


import "array"
import "testing"

testcase join_mismatched_columns {
    observations = array.from(
        rows: [
            {
                Alias: "SIM-SAM-M169",
                Device: 1.0,
                SerialNumber: 1234567890,
                _field: "Pitch",
                _value: 8.4,
                _time: 2021-11-25T00:00:00Z,
            },
            {
                Alias: "SIM-SAM-M169",
                Device: 1.0,
                SerialNumber: 1234567890,
                _field: "Angle",
                _value: 1.2,
                _time: 2021-11-25T00:00:00Z,
            },
            {
                Alias: "SIM-SAM-M169",
                Device: 2.0,
                SerialNumber: 13579,
                _field: "Pitch",
                _value: 9.3,
                _time: 2021-11-25T00:00:00Z,
            },
            {
                Alias: "SIM-SAM-M169",
                Device: 2.0,
                SerialNumber: 13579,
                _field: "Angle",
                _value: 5.6,
                _time: 2021-11-25T00:00:00Z,
            },
            {
                Alias: "SIM-SAM-M169",
                Device: 2.0,
                SerialNumber: 13579,
                _field: "Gauge",
                _value: 9.3,
                _time: 2021-11-25T00:00:00Z,
            },
            {
                Alias: "SIM-SAM-M169",
                Device: 2.0,
                SerialNumber: 13579,
                _field: "Angle",
                _value: 5.6,
                _time: 2021-11-25T00:00:00Z,
            },
        ],
    )
    weather = observations
        |> range(start: 2021-11-24T00:00:00Z, stop: 2021-11-26T00:00:00Z)
        |> filter(fn: (r) => r["Alias"] == "SIM-SAM-M169")
        |> group(columns: ["Alias", "Device", "SerialNumber", "_time"])
        |> pivot(columnKey: ["_field"], rowKey: ["_time"], valueColumn: "_value")

    lkx = observations
        |> range(start: 2021-11-24T00:00:00Z, stop: 2021-11-26T00:00:00Z)
        |> filter(fn: (r) => r["Alias"] == "SIM-SAM-M169")
        |> filter(fn: (r) => r["_field"] == "Pitch")
        |> group(columns: ["Alias", "Device", "SerialNumber", "_time"])
        |> pivot(columnKey: ["_field"], rowKey: ["_time"], valueColumn: "_value")

    want = array.from(
        rows: [
            {
                Alias: "SIM-SAM-M169",
                Device: 1.0,
                SerialNumber: 1234567890,
                Pitch: 8.4,
                Angle: 1.2,
                _time: 2021-11-25T00:00:00Z,
            },
            {
                Alias: "SIM-SAM-M169",
                Device: 2.0,
                SerialNumber: 13579,
                Pitch: 9.3,
                Angle: 5.6,
                _time: 2021-11-25T00:00:00Z,
            },
        ],
    )

    got = join(tables: {d0: weather, d1: lkx}, on: ["_time", "Alias", "Device", "SerialNumber"])
        |> group()

    testing.diff(want: want, got: got)
}
