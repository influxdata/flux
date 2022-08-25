package universe_test


import "array"
import "testing"
import "testing/expect"
import "internal/debug"

testcase vec_with_unary_add {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want =
        array.from(
            rows: [
                {i: 1, f: 1.0, ic: 2, fc: 2.0},
                {i: -1, f: -1.0, ic: 2, fc: 2.0},
                {i: debug.null(type: "int"), f: debug.null(type: "float"), ic: 2, fc: 2.0},
            ],
        )
            |> debug.opaque()

    got =
        array.from(
            rows: [
                {i: 1, f: 1.0},
                {i: -1, f: -1.0},
                {i: debug.null(type: "int"), f: debug.null(type: "float")},
            ],
        )
            |> debug.opaque()
            |> map(
                fn: (r) =>
                    ({r with
                        i: +r.i,
                        f: +r.f,
                        never: +r.never,
                        ic: 2,
                        fc: 2.0,
                    }),
            )

    testing.diff(want, got)
}

testcase vec_with_unary_sub {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want =
        array.from(
            rows: [
                {i: -1, f: -1.0, ic: -2, fc: -2.0},
                {i: 1, f: 1.0, ic: -2, fc: -2.0},
                {i: debug.null(type: "int"), f: debug.null(type: "float"), ic: -2, fc: -2.0},
            ],
        )
            |> debug.opaque()

    got =
        array.from(
            rows: [
                {i: 1, f: 1.0},
                {i: -1, f: -1.0},
                {i: debug.null(type: "int"), f: debug.null(type: "float")},
            ],
        )
            |> debug.opaque()
            |> map(
                fn: (r) =>
                    ({r with
                        i: -r.i,
                        f: -r.f,
                        never: -r.never,
                        ic: -2,
                        fc: -2.0,
                    }),
            )

    testing.diff(want: want, got: got)
}

testcase vec_with_unary_not {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want =
        array.from(
            rows: [
                {a: false, fc: true, tc: false},
                {a: true, fc: true, tc: false},
                {a: debug.null(type: "bool"), fc: true, tc: false},
            ],
        )
            |> debug.opaque()

    got =
        array.from(rows: [{a: true}, {a: false}, {a: debug.null(type: "bool")}])
            |> debug.opaque()
            |> map(
                fn: (r) => ({r with a: not r.a, never: not r.never, fc: not false, tc: not true}),
            )

    testing.diff(want, got)
}

testcase vec_with_unary_exists {
    expect.planner(rules: ["vectorizeMapRule": 1])

    // n.b. of these unary tests, this test for exists is the only one where we
    // expect `never` to appear in the output.
    // In all the other cases, untyped nulls are omitted from the output since they
    // are not valid column data without an associated type. However, since `exists`
    // converts the nulls to bool we actually see `never` this time.
    want =
        array.from(
            rows: [
                {
                    i: true,
                    u: true,
                    f: true,
                    s: true,
                    t: true,
                    b: true,
                    always: true,
                    never: false,
                },
                {
                    i: false,
                    u: true,
                    f: true,
                    s: true,
                    t: true,
                    b: true,
                    always: true,
                    never: false,
                },
                {
                    i: true,
                    u: false,
                    f: true,
                    s: true,
                    t: true,
                    b: true,
                    always: true,
                    never: false,
                },
                {
                    i: true,
                    u: true,
                    f: false,
                    s: true,
                    t: true,
                    b: true,
                    always: true,
                    never: false,
                },
                {
                    i: true,
                    u: true,
                    f: true,
                    s: false,
                    t: true,
                    b: true,
                    always: true,
                    never: false,
                },
                {
                    i: true,
                    u: true,
                    f: true,
                    s: true,
                    t: false,
                    b: true,
                    always: true,
                    never: false,
                },
                {
                    i: true,
                    u: true,
                    f: true,
                    s: true,
                    t: true,
                    b: false,
                    always: true,
                    never: false,
                },
            ],
        )
            |> debug.opaque()

    got =
        array.from(
            rows: [
                {
                    i: 1,
                    u: uint(v: 1),
                    f: 1.0,
                    s: "",
                    t: 2022-08-24T00:00:00Z,
                    b: true,
                },
                {
                    i: debug.null(type: "int"),
                    u: uint(v: 1),
                    f: 1.0,
                    s: "",
                    t: 2022-08-24T00:00:00Z,
                    b: true,
                },
                {
                    i: 1,
                    u: debug.null(type: "uint"),
                    f: 1.0,
                    s: "",
                    t: 2022-08-24T00:00:00Z,
                    b: true,
                },
                {
                    i: 1,
                    u: uint(v: 1),
                    f: debug.null(type: "float"),
                    s: "",
                    t: 2022-08-24T00:00:00Z,
                    b: true,
                },
                {
                    i: 1,
                    u: uint(v: 1),
                    f: 1.0,
                    s: debug.null(type: "string"),
                    t: 2022-08-24T00:00:00Z,
                    b: true,
                },
                {
                    i: 1,
                    u: uint(v: 1),
                    f: 1.0,
                    s: "",
                    t: debug.null(type: "time"),
                    b: true,
                },
                {
                    i: 1,
                    u: uint(v: 1),
                    f: 1.0,
                    s: "",
                    t: 2022-08-24T00:00:00Z,
                    b: debug.null(type: "bool"),
                },
            ],
        )
            |> debug.opaque()
            |> map(
                fn: (r) =>
                    ({r with
                        i: exists r.i,
                        u: exists r.u,
                        f: exists r.f,
                        s: exists r.s,
                        t: exists r.t,
                        b: exists r.b,
                        always: exists 123,
                        never: exists r.never,
                    }),
            )

    testing.diff(want, got)
}
