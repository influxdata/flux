package universe_test


import "array"
import "internal/debug"
import "testing"

testcase invalid_value_field_is_falsy {
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

testcase invalid_only_set_field_is_falsy {
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
            |> filter(fn: (r) => contains(value: 1, set: [r.unknown]))

    testing.diff(want: want, got: got) |> yield()
}

testcase invalid_mix_set_field_is_falsy {
    // Ensure we return rows that match on the valid items even though set has an
    // invalid item as well
    want =
        array.from(
            rows: [
                {
                    _time: 2018-05-22T20:00:10Z,
                    _value: 2,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a",
                },
            ],
        )
            |> debug.opaque()
    got =
        want
            |> debug.opaque()
            |> filter(fn: (r) => contains(value: 2, set: [r._value, r.unknown]))

    testing.diff(want: want, got: got) |> yield()
}

testcase invalid_mix_set_field_is_falsy2 {
    want =
        array.from(
            rows: [
                {
                    _time: 2018-05-22T20:00:10Z,
                    _value: 2,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a",
                },
            ],
        )
            |> debug.opaque()
    got =
        want
            |> debug.opaque()
            |> filter(
                fn: (r) =>
                    contains(
                        value: 2,
                        // The ordering of items in `set` differs from that of the previous
                        // test to make sure we hit each of the possible code paths.
                        set: [r.unknown, r._value],
                    ),
            )

    testing.diff(want: want, got: got) |> yield()
}
