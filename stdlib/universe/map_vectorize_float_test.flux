package universe_test


import "array"
import "testing"
import "testing/expect"
import "internal/debug"

testcase vec_with_float {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want =
        array.from(
            rows: [
                {
                    // float
                    f1: 123.0,
                    f2: 123.0,
                    // int
                    i1: 456,
                    i2: 456.0,
                    // uint
                    u1: uint(v: 789),
                    u2: 789.0,
                    // string
                    s1: "1011.12",
                    s2: 1011.12,
                    // bool false
                    b1F: false,
                    b2F: 0.0,
                    // bool true
                    b1T: true,
                    b2T: 1.0,
                },
            ],
        )

    got =
        array.from(
            rows: [
                {
                    f1: 123.0,
                    i1: 456,
                    u1: uint(v: 789),
                    s1: "1011.12",
                    b1F: false,
                    b1T: true,
                },
            ],
        )
            |> map(
                fn: (r) =>
                    ({r with
                        f2: float(v: r.f1),
                        i2: float(v: r.i1),
                        u2: float(v: r.u1),
                        s2: float(v: r.s1),
                        b2F: float(v: r.b1F),
                        b2T: float(v: r.b1T),
                    }),
            )

    testing.diff(want, got)
}

testcase vec_with_float_const {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want =
        array.from(
            rows: [
                {
                    x: 1,
                    f: 123.456,
                    i: 123.0,
                    bt: 1.0,
                    bf: 0.0,
                },
            ],
        )

    got =
        array.from(rows: [{x: 1}])
            |> map(
                fn: (r) =>
                    ({r with f: float(v: 123.456),
                        i: float(v: 123),
                        bt: float(v: true),
                        bf: float(v: false),
                    }),
            )

    testing.diff(want, got)
}

testcase vec_with_float_typed_null {
    expect.planner(rules: ["vectorizeMapRule": 1])

    // When the input to float() is a typed null, the output will be a null of
    // type float since this matches the column type generally. The input type
    // is not propagated - essentially we do a type cast of other-null to float-null.
    want =
        array.from(
            rows: [{x: 1, _null: debug.null(type: "int"), output: debug.null(type: "float")}],
        )

    got =
        array.from(rows: [{x: 1, _null: debug.null(type: "int")}])
            |> map(fn: (r) => ({r with output: float(v: r._null)}))

    testing.diff(want, got)
}

testcase vec_with_float_untyped_null {
    expect.planner(rules: ["vectorizeMapRule": 1])

    // When the input to float() is an untyped null, the output is also an
    // untyped null meaning the column is dropped from the output.
    want =
        array.from(rows: [{x: 1}])
            |> debug.opaque()

    got =
        array.from(rows: [{x: 1}])
            |> debug.opaque()
            |> map(fn: (r) => ({r with output: float(v: r._null)}))

    testing.diff(want, got)
}
