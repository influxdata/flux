package pagerduty_test


import "array"
import "pagerduty"
import "testing"

input = () => array.from(
    rows: [
        {_time: 2021-06-08T09:00:00Z, _measurement: "m0", _field: "f0", _level: "ok", host: "a", _value: 0.0},
        {_time: 2021-06-08T10:00:00Z, _measurement: "m0", _field: "f0", _level: "crit", host: "a", _value: 0.0},
        {_time: 2021-06-08T11:00:00Z, _measurement: "m0", _field: "f0", _level: "crit", host: "a", _value: 0.0},
        {_time: 2021-06-08T09:00:00Z, _measurement: "m0", _field: "f0", _level: "crit", host: "b", _value: 0.0},
        {_time: 2021-06-08T10:00:00Z, _measurement: "m0", _field: "f0", _level: "crit", host: "b", _value: 0.0},
        {_time: 2021-06-08T11:00:00Z, _measurement: "m0", _field: "f0", _level: "ok", host: "b", _value: 0.0},
    ],
)
    |> group(columns: ["_measurement", "_field", "_level", "host"])

default_tc = (start, stop) => {
    got = input()
        |> testing.load()
        |> range(start, stop)
        |> pagerduty.dedupKey()
        |> drop(columns: ["_start", "_stop", "_value"])

    want = array.from(
        rows: [
            {_time: 2021-06-08T09:00:00Z, _measurement: "m0", _field: "f0", _level: "ok", host: "a", _pagerdutyDedupKey: "4989af42de08cdfbcd8cf398809af7b87bc256ea226fbca5f528185cac9d4a7f"},
            {_time: 2021-06-08T10:00:00Z, _measurement: "m0", _field: "f0", _level: "crit", host: "a", _pagerdutyDedupKey: "4989af42de08cdfbcd8cf398809af7b87bc256ea226fbca5f528185cac9d4a7f"},
            {_time: 2021-06-08T11:00:00Z, _measurement: "m0", _field: "f0", _level: "crit", host: "a", _pagerdutyDedupKey: "4989af42de08cdfbcd8cf398809af7b87bc256ea226fbca5f528185cac9d4a7f"},
            {_time: 2021-06-08T09:00:00Z, _measurement: "m0", _field: "f0", _level: "crit", host: "b", _pagerdutyDedupKey: "d0843e5cb084696e2337732b1e9fa2b06742516a9abdb7b31cf7f1481229a5ea"},
            {_time: 2021-06-08T10:00:00Z, _measurement: "m0", _field: "f0", _level: "crit", host: "b", _pagerdutyDedupKey: "d0843e5cb084696e2337732b1e9fa2b06742516a9abdb7b31cf7f1481229a5ea"},
            {_time: 2021-06-08T11:00:00Z, _measurement: "m0", _field: "f0", _level: "ok", host: "b", _pagerdutyDedupKey: "d0843e5cb084696e2337732b1e9fa2b06742516a9abdb7b31cf7f1481229a5ea"},
        ],
    )
        |> group(columns: ["_measurement", "_field", "_level", "host"])

    return testing.diff(got, want)
}

testcase default {
    default_tc(start: 2021-06-08T09:00:00Z, stop: 2021-06-08T12:00:00Z) |> yield()
}

testcase default_larger_range {
    default_tc(start: 2021-06-08T08:00:00Z, stop: 2021-06-08T13:00:00Z) |> yield()
}

custom_exclude_tc = (start, stop) => {
    got = input()
        |> testing.load()
        |> range(start, stop)
        |> pagerduty.dedupKey(exclude: ["_start", "_stop"])
        |> drop(columns: ["_start", "_stop", "_value"])

    want = array.from(
        rows: [
            {_time: 2021-06-08T09:00:00Z, _measurement: "m0", _field: "f0", _level: "ok", host: "a", _pagerdutyDedupKey: "fd210af9b7d92b39dca3baae5e378101227f198d1aab051e6a3933e766fe52cf"},
            {_time: 2021-06-08T10:00:00Z, _measurement: "m0", _field: "f0", _level: "crit", host: "a", _pagerdutyDedupKey: "185d698c944e84fbda9bddae6796c8d3c4566af88fffd4122aa5d1b26809cfac"},
            {_time: 2021-06-08T11:00:00Z, _measurement: "m0", _field: "f0", _level: "crit", host: "a", _pagerdutyDedupKey: "185d698c944e84fbda9bddae6796c8d3c4566af88fffd4122aa5d1b26809cfac"},
            {_time: 2021-06-08T09:00:00Z, _measurement: "m0", _field: "f0", _level: "crit", host: "b", _pagerdutyDedupKey: "bc2cea845b72cfecab156610d1d5ad6fc13cabf0aa2b24df9faf53c4ca388b89"},
            {_time: 2021-06-08T10:00:00Z, _measurement: "m0", _field: "f0", _level: "crit", host: "b", _pagerdutyDedupKey: "bc2cea845b72cfecab156610d1d5ad6fc13cabf0aa2b24df9faf53c4ca388b89"},
            {_time: 2021-06-08T11:00:00Z, _measurement: "m0", _field: "f0", _level: "ok", host: "b", _pagerdutyDedupKey: "98f1f342a570d9591ffba757e9919a2a850d1c5559e9a99eae5a9f83a4467141"},
        ],
    )
        |> group(columns: ["_measurement", "_field", "_level", "host"])

    return testing.diff(got, want)
}

testcase custom_exclude {
    custom_exclude_tc(start: 2021-06-08T09:00:00Z, stop: 2021-06-08T12:00:00Z) |> yield()
}

testcase custom_exclude_larger_range {
    custom_exclude_tc(start: 2021-06-08T08:00:00Z, stop: 2021-06-08T13:00:00Z) |> yield()
}
