package influxdb_test


import "array"
import "math"
import "testing"

testcase to_roundtrip_floats {
    data = array.from(
        rows: [
            {_time: 2021-03-12T13:58:54Z, _measurement: "m", t0: "a", t1: "b", _field: "float", _value: 42.69},
            {_time: 2021-03-12T13:58:55Z, _measurement: "m", t0: "a", t1: "b", _field: "float", _value: math.pi},
            {_time: 2021-03-12T13:58:56Z, _measurement: "m", t0: "a", t1: "b", _field: "float", _value: math.maxfloat},
            {_time: 2021-03-12T13:58:57Z, _measurement: "m", t0: "a", t1: "b", _field: "float", _value: math.smallestNonzeroFloat},
        // Current storage does not support writing +/-Inf or NaN values
        //{_time: 2021-03-12T13:58:58Z, _measurement: "m", t0: "a", t1: "b", _field: "float", _value: math.mInf(sign:  1)},
        //{_time: 2021-03-12T13:58:59Z, _measurement: "m", t0: "a", t1: "b", _field: "float", _value: math.mInf(sign: -1)},
        //{_time: 2021-03-12T13:59:00Z, _measurement: "m", t0: "a", t1: "b", _field: "float", _value:       math.NaN()},
        ],
    )
        |> group(columns: ["_measurement", "t0", "t1", "_field"])
    want = data
        |> range(start: 2021-03-01T00:00:00Z, stop: 2021-04-01T00:00:00Z)
    got = data
        |> testing.load()
        |> range(start: 2021-03-01T00:00:00Z, stop: 2021-04-01T00:00:00Z)

    testing.diff(got: got, want: want, nansEqual: true)
}
testcase to_roundtrip_ints {
    data = array.from(
        rows: [
            {_time: 2021-03-12T13:58:54Z, _measurement: "m", t0: "a", t1: "b", _field: "int", _value: -67},
            {_time: 2021-03-12T13:58:55Z, _measurement: "m", t0: "a", t1: "b", _field: "int", _value: -67},
            {_time: 2021-03-12T13:58:56Z, _measurement: "m", t0: "a", t1: "b", _field: "int", _value: -67},
            {_time: 2021-03-12T13:58:57Z, _measurement: "m", t0: "a", t1: "b", _field: "int", _value: math.minint},
            {_time: 2021-03-12T13:58:58Z, _measurement: "m", t0: "a", t1: "b", _field: "int", _value: math.maxint},
        ],
    )
        |> group(columns: ["_measurement", "t0", "t1", "_field"])
    want = data
        |> range(start: 2021-03-01T00:00:00Z, stop: 2021-04-01T00:00:00Z)
    got = data
        |> testing.load()
        |> range(start: 2021-03-01T00:00:00Z, stop: 2021-04-01T00:00:00Z)

    testing.diff(got: got, want: want)
}
testcase to_roundtrip_uints {
    data = array.from(
        rows: [
            {_time: 2021-03-12T13:58:54Z, _measurement: "m", t0: "a", t1: "b", _field: "uint", _value: uint(v: 0)},
            {_time: 2021-03-12T13:58:55Z, _measurement: "m", t0: "a", t1: "b", _field: "uint", _value: uint(v: 152)},
            {_time: 2021-03-12T13:58:56Z, _measurement: "m", t0: "a", t1: "b", _field: "uint", _value: uint(v: 152)},
            {_time: 2021-03-12T13:58:57Z, _measurement: "m", t0: "a", t1: "b", _field: "uint", _value: uint(v: 152)},
            {_time: 2021-03-12T13:58:58Z, _measurement: "m", t0: "a", t1: "b", _field: "uint", _value: math.maxuint},
        ],
    )
        |> group(columns: ["_measurement", "t0", "t1", "_field"])
    want = data
        |> range(start: 2021-03-01T00:00:00Z, stop: 2021-04-01T00:00:00Z)
    got = data
        |> testing.load()
        |> range(start: 2021-03-01T00:00:00Z, stop: 2021-04-01T00:00:00Z)

    testing.diff(got: got, want: want)
}
testcase to_roundtrip_bools {
    data = array.from(
        rows: [
            {_time: 2021-03-12T13:58:54Z, _measurement: "m", t0: "a", t1: "b", _field: "bool", _value: true},
            {_time: 2021-03-12T13:58:55Z, _measurement: "m", t0: "a", t1: "b", _field: "bool", _value: false},
            {_time: 2021-03-12T13:58:56Z, _measurement: "m", t0: "a", t1: "b", _field: "bool", _value: true},
            {_time: 2021-03-12T13:58:57Z, _measurement: "m", t0: "a", t1: "b", _field: "bool", _value: false},
            {_time: 2021-03-12T13:58:58Z, _measurement: "m", t0: "a", t1: "b", _field: "bool", _value: true},
        ],
    )
        |> group(columns: ["_measurement", "t0", "t1", "_field"])
    want = data
        |> range(start: 2021-03-01T00:00:00Z, stop: 2021-04-01T00:00:00Z)
    got = data
        |> testing.load()
        |> range(start: 2021-03-01T00:00:00Z, stop: 2021-04-01T00:00:00Z)

    testing.diff(got: got, want: want)
}
testcase to_roundtrip_strings {
    data = array.from(
        rows: [
            {_time: 2021-03-12T13:58:54Z, _measurement: "m", t0: "a", t1: "w", _field: "string", _value: "hello world"},
            {_time: 2021-03-12T13:58:55Z, _measurement: "m", t0: "a", t1: "x", _field: "string", _value: "你好，世界"},
            {_time: 2021-03-12T13:58:56Z, _measurement: "m", t0: "b", t1: "y", _field: "string", _value: "여보세요 세계"},
            {_time: 2021-03-12T13:58:57Z, _measurement: "m", t0: "b", t1: "z", _field: "string", _value: "hello world"},
        ],
    )
        |> group(columns: ["_measurement", "t0", "t1", "_field"])
    want = data
        |> range(start: 2021-03-01T00:00:00Z, stop: 2021-04-01T00:00:00Z)
    got = data
        |> testing.load()
        |> range(start: 2021-03-01T00:00:00Z, stop: 2021-04-01T00:00:00Z)

    testing.diff(got: got, want: want)
}
