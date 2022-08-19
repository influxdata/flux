package universe_test


import "array"
import "testing"

// The intent with these tests is to ensure coverage for vectorized conditional
// expressions. Only a subset of flux types work due to specifics of the
// underlying arrow array storage used during the vectorize optimization.
//
// The results of these tests are lackluster on their face. The outcome should
// be plain and straight forward and should even work as-is when NOT vectorized.
// Ideally, we'd be able to assert that the function was in fact vectorized, but
// this will not be possible until <https://github.com/influxdata/flux/issues/4739>
// is completed.
// For now, the verification of actually hitting the vectorized code path can be
// done manually with a debugger, or by running these cases as individual scripts
// on the CLI with tracing enabled.
fn = (r) => ({v: if r.cond then r.a else r.b})

testcase vec_conditional_time {
    want = array.from(rows: [{v: 2022-01-01T00:00:00Z}, {v: 2022-01-04T00:00:00Z}])

    got =
        array.from(
            rows: [
                {cond: true, a: 2022-01-01T00:00:00Z, b: 2022-01-02T00:00:00Z},
                {cond: false, a: 2022-01-03T00:00:00Z, b: 2022-01-04T00:00:00Z},
            ],
        )
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_conditional_time_repeat {
    want = array.from(rows: [{v: 1955-11-12T00:00:00Z}, {v: 1985-10-26T00:00:00Z}])

    got =
        array.from(rows: [{cond: true}, {cond: false}])
            |> map(fn: (r) => ({v: if r.cond then 1955-11-12T00:00:00Z else 1985-10-26T00:00:00Z}))

    testing.diff(want: want, got: got)
}

testcase vec_conditional_int {
    want = array.from(rows: [{v: -1}, {v: 2}])

    got =
        array.from(rows: [{cond: true, a: -1, b: 0}, {cond: false, a: 1, b: 2}])
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_conditional_int_repeat {
    want = array.from(rows: [{v: 10}, {v: 20}])

    got =
        array.from(rows: [{cond: true}, {cond: false}])
            |> map(fn: (r) => ({v: if r.cond then 10 else 20}))

    testing.diff(want: want, got: got)
}

testcase vec_conditional_float {
    want = array.from(rows: [{v: -1.0}, {v: 2.0}])

    got =
        array.from(rows: [{cond: true, a: -1.0, b: 0.0}, {cond: false, a: 1.0, b: 2.0}])
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_conditional_float_repeat {
    want = array.from(rows: [{v: 10.0}, {v: 20.0}])

    got =
        array.from(rows: [{cond: true}, {cond: false}])
            |> map(fn: (r) => ({v: if r.cond then 10.0 else 20.0}))

    testing.diff(want: want, got: got)
}

// FIXME: can't do a vec repeat version for uint until we vectorize uint
testcase vec_conditional_uint {
    want = array.from(rows: [{v: uint(v: 0)}, {v: uint(v: 3)}])

    got =
        array.from(
            rows: [
                {cond: true, a: uint(v: 0), b: uint(v: 1)},
                {cond: false, a: uint(v: 2), b: uint(v: 3)},
            ],
        )
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_conditional_string {
    want = array.from(rows: [{v: "a"}, {v: "d"}])

    got =
        array.from(rows: [{cond: true, a: "a", b: "b"}, {cond: false, a: "c", b: "d"}])
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_conditional_string_repeat {
    want = array.from(rows: [{v: "yes"}, {v: "no"}])

    got =
        array.from(rows: [{cond: true}, {cond: false}])
            |> map(fn: (r) => ({v: if r.cond then "yes" else "no"}))

    testing.diff(want: want, got: got)
}

testcase vec_conditional_bool {
    want = array.from(rows: [{v: true}, {v: false}])

    got =
        array.from(rows: [{cond: true, a: true, b: false}, {cond: false, a: true, b: false}])
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_conditional_bool_repeat {
    want = array.from(rows: [{v: true}, {v: false}])

    got =
        array.from(rows: [{cond: true}, {cond: false}])
            |> map(fn: (r) => ({v: if r.cond then true else false}))

    testing.diff(want: want, got: got)
}
