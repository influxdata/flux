package universe_test


import "array"
import "testing"
import "testing/expect"

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
    expect.planner(rules: ["vectorizeMapRule": 1])

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
    expect.planner(rules: ["vectorizeMapRule": 1])

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
        array.from(
            rows: [{a: 2022-01-01T00:00:00Z}, {a: 2022-01-01T01:00:00Z}, {a: 2022-01-01T02:00:00Z}],
        )
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
    expect.planner(rules: ["vectorizeMapRule": 1])

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
    expect.planner(rules: ["vectorizeMapRule": 1])

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
    expect.planner(rules: ["vectorizeMapRule": 1])

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
    expect.planner(rules: ["vectorizeMapRule": 1])

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
    expect.planner(rules: ["vectorizeMapRule": 1])

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
            rows: [
                {a: uint(v: 0), b: uint(v: 0)},
                {a: uint(v: 2), b: uint(v: 0)},
                {a: uint(v: 0), b: uint(v: 2)},
            ],
        )
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_equality_string {
    expect.planner(rules: ["vectorizeMapRule": 1])

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
    expect.planner(rules: ["vectorizeMapRule": 1])

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

testcase vec_equality_bool {
    expect.planner(rules: ["vectorizeMapRule": 1])

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
        array.from(
            rows: [
                {a: false, b: false},
                {a: true, b: false},
                {a: false, b: true},
                {a: true, b: true},
            ],
        )
            // N.b. bool currently only supports `Equatable` but not `Comparable`
            // so we can only test for eq/neq at this time.
            // In the future bool may become `Comparable` in which case we can
            // rewrite this test using the `fn` defined at the top of this file.
            |> map(fn: (r) => ({r with eq: r.a == r.b, neq: r.a != r.b}))

    testing.diff(want: want, got: got)
}

testcase vec_equality_bool_repeat {
    want =
        array.from(
            rows: [
                {
                    a: false,
                    eq: false,
                    neq: true,
                    tteq: true,
                    ffeq: true,
                    tfeq: false,
                    fteq: false,
                    ttneq: false,
                    ffneq: false,
                    tfneq: true,
                    ftneq: true,
                },
                {
                    a: true,
                    eq: true,
                    neq: false,
                    tteq: true,
                    ffeq: true,
                    tfeq: false,
                    fteq: false,
                    ttneq: false,
                    ffneq: false,
                    tfneq: true,
                    ftneq: true,
                },
                {
                    a: false,
                    eq: false,
                    neq: true,
                    tteq: true,
                    ffeq: true,
                    tfeq: false,
                    fteq: false,
                    ttneq: false,
                    ffneq: false,
                    tfneq: true,
                    ftneq: true,
                },
                {
                    a: true,
                    eq: true,
                    neq: false,
                    tteq: true,
                    ffeq: true,
                    tfeq: false,
                    fteq: false,
                    ttneq: false,
                    ffneq: false,
                    tfneq: true,
                    ftneq: true,
                },
            ],
        )

    got =
        array.from(rows: [{a: false}, {a: true}, {a: false}, {a: true}])
            // N.b. bool currently only supports `Equatable` but not `Comparable`
            // so we can only test for eq/neq at this time.
            // In the future bool may become `Comparable` in which case we can
            // rewrite this test using an `fn` more similar to the one defined
            // at the top of this file.
            |> map(
                fn: (r) =>
                    ({r with
                        eq: r.a == true,
                        neq: r.a != true,
                        tteq: true == true,
                        ffeq: false == false,
                        tfeq: true == false,
                        fteq: false == true,
                        ttneq: true != true,
                        ffneq: false != false,
                        tfneq: true != false,
                        ftneq: false != true,
                    }),
            )

    testing.diff(want: want, got: got)
}

// Ensure implicit casting between numerics (float, int, uint) is supported.
// There's special handling for cases where int and uint are compared.
// In situations where any other number type is compared to a float, the other
// is cast to float.
testcase vec_equality_casts {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want =
        array.from(
            rows: [
                {
                    i: 0,
                    u: uint(v: 0),
                    f: 0.0,
                    iuEq: true,
                    ifEq: true,
                    fiEq: true,
                    fuEq: true,
                    uiEq: true,
                    ufEq: true,
                    iuNeq: false,
                    ifNeq: false,
                    fiNeq: false,
                    fuNeq: false,
                    uiNeq: false,
                    ufNeq: false,
                    iuLt: false,
                    ifLt: false,
                    fiLt: false,
                    fuLt: false,
                    uiLt: false,
                    ufLt: false,
                    iuLte: true,
                    ifLte: true,
                    fiLte: true,
                    fuLte: true,
                    uiLte: true,
                    ufLte: true,
                    iuGt: false,
                    ifGt: false,
                    fiGt: false,
                    fuGt: false,
                    uiGt: false,
                    ufGt: false,
                    iuGte: true,
                    ifGte: true,
                    fiGte: true,
                    fuGte: true,
                    uiGte: true,
                    ufGte: true,
                },
                {
                    i: 123,
                    u: uint(v: 123),
                    f: 123.0,
                    iuEq: true,
                    ifEq: true,
                    fiEq: true,
                    fuEq: true,
                    uiEq: true,
                    ufEq: true,
                    iuNeq: false,
                    ifNeq: false,
                    fiNeq: false,
                    fuNeq: false,
                    uiNeq: false,
                    ufNeq: false,
                    iuLt: false,
                    ifLt: false,
                    fiLt: false,
                    fuLt: false,
                    uiLt: false,
                    ufLt: false,
                    iuLte: true,
                    ifLte: true,
                    fiLte: true,
                    fuLte: true,
                    uiLte: true,
                    ufLte: true,
                    iuGt: false,
                    ifGt: false,
                    fiGt: false,
                    fuGt: false,
                    uiGt: false,
                    ufGt: false,
                    iuGte: true,
                    ifGte: true,
                    fiGte: true,
                    fuGte: true,
                    uiGte: true,
                    ufGte: true,
                },
                {
                    i: -123,
                    u: uint(v: 123),
                    f: 0.1,
                    iuEq: false,
                    ifEq: false,
                    fiEq: false,
                    fuEq: false,
                    uiEq: false,
                    ufEq: false,
                    iuNeq: true,
                    ifNeq: true,
                    fiNeq: true,
                    fuNeq: true,
                    uiNeq: true,
                    ufNeq: true,
                    iuLt: true,
                    ifLt: true,
                    fiLt: false,
                    fuLt: true,
                    uiLt: false,
                    ufLt: false,
                    iuLte: true,
                    ifLte: true,
                    fiLte: false,
                    fuLte: true,
                    uiLte: false,
                    ufLte: false,
                    iuGt: false,
                    ifGt: false,
                    fiGt: true,
                    fuGt: false,
                    uiGt: true,
                    ufGt: true,
                    iuGte: false,
                    ifGte: false,
                    fiGte: true,
                    fuGte: false,
                    uiGte: true,
                    ufGte: true,
                },
            ],
        )

    got =
        array.from(
            rows: [
                {i: 0, u: uint(v: 0), f: 0.0},
                {i: 123, u: uint(v: 123), f: 123.0},
                {i: -123, u: uint(v: 123), f: 0.1},
            ],
        )
            |> map(
                fn: (r) =>
                    ({r with
                        iuEq: r.i == r.u,
                        ifEq: r.i == r.f,
                        fiEq: r.f == r.i,
                        fuEq: r.f == r.u,
                        uiEq: r.u == r.i,
                        ufEq: r.u == r.f,
                        iuNeq: r.i != r.u,
                        ifNeq: r.i != r.f,
                        fiNeq: r.f != r.i,
                        fuNeq: r.f != r.u,
                        uiNeq: r.u != r.i,
                        ufNeq: r.u != r.f,
                        iuLt: r.i < r.u,
                        ifLt: r.i < r.f,
                        fiLt: r.f < r.i,
                        fuLt: r.f < r.u,
                        uiLt: r.u < r.i,
                        ufLt: r.u < r.f,
                        iuLte: r.i <= r.u,
                        ifLte: r.i <= r.f,
                        fiLte: r.f <= r.i,
                        fuLte: r.f <= r.u,
                        uiLte: r.u <= r.i,
                        ufLte: r.u <= r.f,
                        iuGt: r.i > r.u,
                        ifGt: r.i > r.f,
                        fiGt: r.f > r.i,
                        fuGt: r.f > r.u,
                        uiGt: r.u > r.i,
                        ufGt: r.u > r.f,
                        iuGte: r.i >= r.u,
                        ifGte: r.i >= r.f,
                        fiGte: r.f >= r.i,
                        fuGte: r.f >= r.u,
                        uiGte: r.u >= r.i,
                        ufGte: r.u >= r.f,
                    }),
            )

    testing.diff(want: want, got: got)
}
