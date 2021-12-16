package experimental_test


import "testing"
import "experimental"
import "array"

testcase addDuration_to_time {
    option now = () => 2021-12-09T00:00:00Z

    cases =
        array.from(
            rows: [
                {d: int(v: 1h), to: now()},
                {d: int(v: 2h), to: now()},
                {d: int(v: 2s), to: now()},
                {d: int(v: 2h), to: 2020-01-01T00:00:00Z},
                {d: int(v: 3d), to: 2020-01-01T00:00:00Z},
            ],
        )
    want =
        array.from(
            rows: [
                {_value: 2021-12-09T01:00:00Z},
                {_value: 2021-12-09T02:00:00Z},
                {_value: 2021-12-09T00:00:02Z},
                {_value: 2020-01-01T02:00:00Z},
                {_value: 2020-01-04T00:00:00Z},
            ],
        )

    got = cases |> map(fn: (r) => ({_value: experimental.addDuration(d: duration(v: r.d), to: r.to)}))

    testing.diff(want: want, got: got)
}
testcase addDuration_to_time_mixed {
    //TODO(nathaniel): When we either have the ability to assert on scalar values or can represent durations in tables
    // we can fold this test case into the test case above
    option now = () => 2021-12-09T00:00:00Z

    want = array.from(rows: [{_value: 2022-03-09T00:00:00Z}])

    got = array.from(rows: [{_value: experimental.addDuration(d: 3mo, to: now())}])

    testing.diff(want: want, got: got)
}
testcase addDuration_to_duration_as_time {
    option now = () => 2021-12-09T00:00:00Z

    cases =
        array.from(
            rows: [
                {d: int(v: 1h), to: int(v: -1h)},
                {d: int(v: 1h), to: int(v: -1d)},
                {d: int(v: 1h), to: int(v: -1w)},
                {d: int(v: 1d), to: int(v: -1h)},
                {d: int(v: 1s), to: int(v: -1d)},
                {d: int(v: 1ms), to: int(v: -1w)},
            ],
        )
    want =
        array.from(
            rows: [
                {_value: 2021-12-09T00:00:00Z},
                {_value: 2021-12-08T01:00:00Z},
                {_value: 2021-12-02T01:00:00Z},
                {_value: 2021-12-09T23:00:00Z},
                {_value: 2021-12-08T00:00:01Z},
                {_value: 2021-12-02T00:00:00.001Z},
            ],
        )

    got = cases |> map(fn: (r) => ({_value: experimental.addDuration(d: duration(v: r.d), to: duration(v: r.to))}))

    testing.diff(want: want, got: got)
}
testcase addDuration_to_duration_mixed {
    //TODO(nathaniel): When we either have the ability to assert on scalar values or can represent durations in tables
    // we can fold this test case into the test case above
    option now = () => 2021-12-09T00:00:00Z

    want = array.from(rows: [{_value: 2022-01-08T12:00:00Z}])

    got = array.from(rows: [{_value: experimental.addDuration(d: 1mo, to: -12h)}])

    testing.diff(want: want, got: got)
}

testcase subDuration_to_time {
    option now = () => 2021-12-09T00:00:00Z

    cases =
        array.from(
            rows: [
                {d: int(v: 1h), from: now()},
                {d: int(v: 2h), from: now()},
                {d: int(v: 2s), from: now()},
                {d: int(v: 2h), from: 2020-01-01T00:00:00Z},
                {d: int(v: 3d), from: 2020-01-01T00:00:00Z},
            ],
        )
    want =
        array.from(
            rows: [
                {_value: 2021-12-08T23:00:00Z},
                {_value: 2021-12-08T22:00:00Z},
                {_value: 2021-12-08T23:59:58Z},
                {_value: 2019-12-31T22:00:00Z},
                {_value: 2019-12-29T00:00:00Z},
            ],
        )

    got = cases |> map(fn: (r) => ({_value: experimental.subDuration(d: duration(v: r.d), from: r.from)}))

    testing.diff(want: want, got: got)
}
testcase subDuration_to_time_mixed {
    //TODO(nathaniel): When we either have the ability to assert on scalar values or can represent durations in tables
    // we can fold this test case into the test case above
    option now = () => 2021-12-09T00:00:00Z

    want = array.from(rows: [{_value: 2020-12-09T00:00:00Z}])

    got = array.from(rows: [{_value: experimental.subDuration(d: 1y, from: now())}])

    testing.diff(want: want, got: got)
}
testcase subDuration_to_duration_as_time {
    option now = () => 2021-12-09T00:00:00Z

    cases =
        array.from(
            rows: [
                {d: int(v: 1h), from: int(v: -1h)},
                {d: int(v: 1h), from: int(v: -1d)},
                {d: int(v: 1h), from: int(v: -1w)},
                {d: int(v: 1d), from: int(v: -1h)},
                {d: int(v: 1s), from: int(v: -1d)},
                {d: int(v: 1ms), from: int(v: -1w)},
            ],
        )
    want =
        array.from(
            rows: [
                {_value: 2021-12-08T22:00:00Z},
                {_value: 2021-12-07T23:00:00Z},
                {_value: 2021-12-01T23:00:00Z},
                {_value: 2021-12-07T23:00:00Z},
                {_value: 2021-12-07T23:59:59Z},
                {_value: 2021-12-01T23:59:59.999Z},
            ],
        )

    got = cases |> map(fn: (r) => ({_value: experimental.subDuration(d: duration(v: r.d), from: duration(v: r.from))}))

    testing.diff(want: want, got: got)
}
testcase subDuration_to_duration_mixed {
    //TODO(nathaniel): When we either have the ability to assert on scalar values or can represent durations in tables
    // we can fold this test case into the test case above
    option now = () => 2021-12-09T00:00:00Z

    want = array.from(rows: [{_value: 2019-12-08T12:00:00Z}])

    got = array.from(rows: [{_value: experimental.subDuration(d: 2y, from: -12h)}])

    testing.diff(want: want, got: got)
}
