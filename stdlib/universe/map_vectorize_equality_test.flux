package universe_test


import "array"
import "testing"

// The intent with these tests is to ensure coverage for vectorized equality
// operators. Only a subset of flux types work due to specifics of the
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
//
fn = (r) =>
    ({r with
        eq: r.a == r.b,
        neq: r.a != r.b,
        lt: r.a < r.b,
        lte: r.a <= r.b,
        gt: r.a > r.b,
        gte: r.a >= r.b,
    })

testcase vec_equality_time {
    want =
        array.from(
            rows: [
                {
                    a: 2022-01-01T00:00:00Z,
                    b: 2022-01-01T00:00:00Z,
                    eq: true,
                    neq: false,
                    lt: false,
                    lte: true,
                    gt: false,
                    gte: true,
                },
                {
                    a: 2022-01-01T02:00:00Z,
                    b: 2022-01-01T00:00:00Z,
                    eq: false,
                    neq: true,
                    lt: false,
                    lte: false,
                    gt: true,
                    gte: true,
                },
                {
                    a: 2022-01-01T00:00:00Z,
                    b: 2022-01-01T02:00:00Z,
                    eq: false,
                    neq: true,
                    lt: true,
                    lte: true,
                    gt: false,
                    gte: false,
                },
            ],
        )

    got =
        array.from(
            rows: [
                {a: 2022-01-01T00:00:00Z, b: 2022-01-01T00:00:00Z},
                {a: 2022-01-01T02:00:00Z, b: 2022-01-01T00:00:00Z},
                {a: 2022-01-01T00:00:00Z, b: 2022-01-01T02:00:00Z},
            ],
        )
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_equality_time_repeat {
    want =
        array.from(
            rows: [
                {
                    a: 2022-01-01T00:00:00Z,
                    eq: false,
                    neq: true,
                    lt: true,
                    lte: true,
                    gt: false,
                    gte: false,
                },
                {
                    a: 2022-01-01T01:00:00Z,
                    eq: true,
                    neq: false,
                    lt: false,
                    lte: true,
                    gt: false,
                    gte: true,
                },
                {
                    a: 2022-01-01T02:00:00Z,
                    eq: false,
                    neq: true,
                    lt: false,
                    lte: false,
                    gt: true,
                    gte: true,
                },
            ],
        )

    got =
        array.from(rows: [{a: 2022-01-01T00:00:00Z}, {a: 2022-01-01T01:00:00Z}, {a: 2022-01-01T02:00:00Z}])
            |> map(
                fn: (r) =>
                    ({r with
                        eq: r.a == 2022-01-01T01:00:00Z,
                        neq: r.a != 2022-01-01T01:00:00Z,
                        lt: r.a < 2022-01-01T01:00:00Z,
                        lte: r.a <= 2022-01-01T01:00:00Z,
                        gt: r.a > 2022-01-01T01:00:00Z,
                        gte: r.a >= 2022-01-01T01:00:00Z,
                    }),
            )

    testing.diff(want: want, got: got)
}

testcase vec_equality_int {
    want =
        array.from(
            rows: [
                {
                    a: 0,
                    b: 0,
                    eq: true,
                    neq: false,
                    lt: false,
                    lte: true,
                    gt: false,
                    gte: true,
                },
                {
                    a: 2,
                    b: 0,
                    eq: false,
                    neq: true,
                    lt: false,
                    lte: false,
                    gt: true,
                    gte: true,
                },
                {
                    a: 0,
                    b: 2,
                    eq: false,
                    neq: true,
                    lt: true,
                    lte: true,
                    gt: false,
                    gte: false,
                },
            ],
        )

    got =
        array.from(rows: [{a: 0, b: 0}, {a: 2, b: 0}, {a: 0, b: 2}])
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_equality_int_repeat {
    want =
        array.from(
            rows: [
                {
                    a: -1,
                    eq: false,
                    neq: true,
                    lt: true,
                    lte: true,
                    gt: false,
                    gte: false,
                },
                {
                    a: 0,
                    eq: true,
                    neq: false,
                    lt: false,
                    lte: true,
                    gt: false,
                    gte: true,
                },
                {
                    a: 1,
                    eq: false,
                    neq: true,
                    lt: false,
                    lte: false,
                    gt: true,
                    gte: true,
                },
            ],
        )

    got =
        array.from(rows: [{a: -1}, {a: 0}, {a: 1}])
            |> map(
                fn: (r) =>
                    ({r with
                        eq: r.a == 0,
                        neq: r.a != 0,
                        lt: r.a < 0,
                        lte: r.a <= 0,
                        gt: r.a > 0,
                        gte: r.a >= 0,
                    }),
            )

    testing.diff(want: want, got: got)
}

testcase vec_equality_float {
    want =
        array.from(
            rows: [
                {
                    a: 0.0,
                    b: 0.0,
                    eq: true,
                    neq: false,
                    lt: false,
                    lte: true,
                    gt: false,
                    gte: true,
                },
                {
                    a: 2.0,
                    b: 0.0,
                    eq: false,
                    neq: true,
                    lt: false,
                    lte: false,
                    gt: true,
                    gte: true,
                },
                {
                    a: 0.0,
                    b: 2.0,
                    eq: false,
                    neq: true,
                    lt: true,
                    lte: true,
                    gt: false,
                    gte: false,
                },
            ],
        )
    got =
        array.from(rows: [{a: 0.0, b: 0.0}, {a: 2.0, b: 0.0}, {a: 0.0, b: 2.0}])
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_equality_float_repeat {
    want =
        array.from(
            rows: [
                {
                    a: -1.0,
                    eq: false,
                    neq: true,
                    lt: true,
                    lte: true,
                    gt: false,
                    gte: false,
                },
                {
                    a: 0.0,
                    eq: true,
                    neq: false,
                    lt: false,
                    lte: true,
                    gt: false,
                    gte: true,
                },
                {
                    a: 1.0,
                    eq: false,
                    neq: true,
                    lt: false,
                    lte: false,
                    gt: true,
                    gte: true,
                },
            ],
        )

    got =
        array.from(rows: [{a: -1.0}, {a: 0.0}, {a: 1.0}])
            |> map(
                fn: (r) =>
                    ({r with
                        eq: r.a == 0.0,
                        neq: r.a != 0.0,
                        lt: r.a < 0.0,
                        lte: r.a <= 0.0,
                        gt: r.a > 0.0,
                        gte: r.a >= 0.0,
                    }),
            )

    testing.diff(want: want, got: got)
}

// FIXME: can't do a vec repeat version for uint until we vectorize uint
testcase vec_equality_uint {
    want =
        array.from(
            rows: [
                {
                    a: uint(v: 0),
                    b: uint(v: 0),
                    eq: true,
                    neq: false,
                    lt: false,
                    lte: true,
                    gt: false,
                    gte: true,
                },
                {
                    a: uint(v: 2),
                    b: uint(v: 0),
                    eq: false,
                    neq: true,
                    lt: false,
                    lte: false,
                    gt: true,
                    gte: true,
                },
                {
                    a: uint(v: 0),
                    b: uint(v: 2),
                    eq: false,
                    neq: true,
                    lt: true,
                    lte: true,
                    gt: false,
                    gte: false,
                },
            ],
        )

    got =
        array.from(
            rows: [{a: uint(v: 0), b: uint(v: 0)}, {a: uint(v: 2), b: uint(v: 0)}, {a: uint(v: 0), b: uint(v: 2)}],
        )
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_equality_string {
    want =
        array.from(
            rows: [
                {
                    a: "x",
                    b: "x",
                    eq: true,
                    neq: false,
                    lt: false,
                    lte: true,
                    gt: false,
                    gte: true,
                },
                {
                    a: "z",
                    b: "x",
                    eq: false,
                    neq: true,
                    lt: false,
                    lte: false,
                    gt: true,
                    gte: true,
                },
                {
                    a: "x",
                    b: "z",
                    eq: false,
                    neq: true,
                    lt: true,
                    lte: true,
                    gt: false,
                    gte: false,
                },
            ],
        )

    got =
        array.from(rows: [{a: "x", b: "x"}, {a: "z", b: "x"}, {a: "x", b: "z"}])
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_equality_string_repeat {
    want =
        array.from(
            rows: [
                {
                    a: "x",
                    eq: false,
                    neq: true,
                    lt: true,
                    lte: true,
                    gt: false,
                    gte: false,
                },
                {
                    a: "y",
                    eq: true,
                    neq: false,
                    lt: false,
                    lte: true,
                    gt: false,
                    gte: true,
                },
                {
                    a: "z",
                    eq: false,
                    neq: true,
                    lt: false,
                    lte: false,
                    gt: true,
                    gte: true,
                },
            ],
        )

    got =
        array.from(rows: [{a: "x"}, {a: "y"}, {a: "z"}])
            |> map(
                fn: (r) =>
                    ({r with
                        eq: r.a == "y",
                        neq: r.a != "y",
                        lt: r.a < "y",
                        lte: r.a <= "y",
                        gt: r.a > "y",
                        gte: r.a >= "y",
                    }),
            )

    testing.diff(want: want, got: got)
}

// FIXME: can't do a vec repeat for bool until bool literals can vectorize
testcase vec_equality_bool {
    want =
        array.from(
            rows: [
                {a: false, b: false, eq: true, neq: false},
                {a: true, b: false, eq: false, neq: true},
                {a: false, b: true, eq: false, neq: true},
                {a: true, b: true, eq: true, neq: false},
            ],
        )

    got =
        array.from(rows: [{a: false, b: false}, {a: true, b: false}, {a: false, b: true}, {a: true, b: true}])
            // N.b. bool currently only supports `Equatable` but not `Comparable`
            // so we can only test for eq/neq at this time.
            // In the future bool may become `Comparable` in which case we can
            // rewrite this test using the `fn` defined at the top of this file.
            |> map(fn: (r) => ({r with eq: r.a == r.b, neq: r.a != r.b}))

    testing.diff(want: want, got: got)
}
