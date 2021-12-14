package monitor_test


import "array"
import "influxdata/influxdb/monitor"
import "influxdata/influxdb/v1"
import "testing"

testcase state_changes_any_to_any_same_source_timestamp {
    input =
        array.from(
            rows: [
                {
                    _check_id: "000000000000000a",
                    _check_name: "cpu threshold check",
                    _level: "ok",
                    _measurement: "statuses",
                    _source_measurement: "cpu",
                    _time: 2018-05-22T19:54:20Z,
                    _type: "threshold",
                    cpu: "cpu-total",
                    host: "host.local",
                    usage_idle: 4.800000000000001,
                    _message: "whoa!",
                    _source_timestamp: 1527018820000000000,
                },
                {
                    _check_id: "000000000000000a",
                    _check_name: "cpu threshold check",
                    _level: "crit",
                    _measurement: "statuses",
                    _source_measurement: "cpu",
                    _time: 2018-05-22T19:54:21Z,
                    _type: "threshold",
                    cpu: "cpu-total",
                    host: "host.local",
                    usage_idle: 90.62382797849732,
                    _message: "whoa!",
                    _source_timestamp: 1527018820000000000,
                },
                {
                    _check_id: "000000000000000a",
                    _check_name: "cpu threshold check",
                    _level: "warn",
                    _measurement: "statuses",
                    _source_measurement: "cpu",
                    _time: 2018-05-22T19:54:22Z,
                    _type: "threshold",
                    cpu: "cpu-total",
                    host: "host.local",
                    usage_idle: 7.05,
                    _message: "whoa!",
                    _source_timestamp: 1527018820000000000,
                },
            ],
        )
            |> group(
                columns: [
                    "_check_id",
                    "_check_name",
                    "_level",
                    "_measurement",
                    "_source_measurement",
                    "_type",
                    "cpu",
                    "host",
                ],
            )
            |> range(start: 2018-05-22T19:54:00Z, stop: 2018-05-22T19:55:00Z)
    want =
        array.from(
            rows: [
                {
                    _check_id: "000000000000000a",
                    _check_name: "cpu threshold check",
                    _measurement: "statuses",
                    _message: "whoa!",
                    _source_measurement: "cpu",
                    _source_timestamp: 1527018820000000000,
                    _time: 2018-05-22T19:54:21Z,
                    _type: "threshold",
                    cpu: "cpu-total",
                    host: "host.local",
                    usage_idle: 90.62382797849732,
                    _level: "crit",
                },
                {
                    _check_id: "000000000000000a",
                    _check_name: "cpu threshold check",
                    _measurement: "statuses",
                    _message: "whoa!",
                    _source_measurement: "cpu",
                    _source_timestamp: 1527018820000000000,
                    _time: 2018-05-22T19:54:22Z,
                    _type: "threshold",
                    cpu: "cpu-total",
                    host: "host.local",
                    usage_idle: 7.05,
                    _level: "warn",
                },
            ],
        )
            |> group(
                columns: [
                    "_check_id",
                    "_check_name",
                    "_measurement",
                    "_source_measurement",
                    "_type",
                    "cpu",
                    "host",
                    "_level",
                ],
            )
    got =
        input
            // already pivoted
            |> monitor.stateChanges(fromLevel: "any", toLevel: "any")
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}
