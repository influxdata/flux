package universe_test


import "array"
import "internal/debug"
import "testing"

testcase invalid_field_is_falsy {
    want = array.from(rows: [{}]) |> filter(fn: (r) => false) |> debug.opaque()
    got =
        array.from(
            rows: [
                {
                    _time: 2018-05-22T20:00:10Z,
                    _value: 0.336,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a",
                },
            ],
        )
            |> debug.opaque()
            |> filter(fn: (r) => contains(value: r.unknown, set: [1, 2, 3]))

    testing.diff(want: want, got: got) |> yield()
}
