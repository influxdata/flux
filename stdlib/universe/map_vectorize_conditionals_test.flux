package universe_test


import "array"
import "internal/debug"
import "testing"
import "testing/expect"

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
    expect.planner(rules: ["vectorizeMapRule": 1])

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
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{v: 1955-11-12T00:00:00Z}, {v: 1985-10-26T00:00:00Z}])

    got =
        array.from(rows: [{cond: true}, {cond: false}])
            |> map(fn: (r) => ({v: if r.cond then 1955-11-12T00:00:00Z else 1985-10-26T00:00:00Z}))

    testing.diff(want: want, got: got)
}

testcase vec_conditional_int {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{v: -1}, {v: 2}])

    got =
        array.from(rows: [{cond: true, a: -1, b: 0}, {cond: false, a: 1, b: 2}])
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_conditional_int_repeat {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{v: 10}, {v: 20}])

    got =
        array.from(rows: [{cond: true}, {cond: false}])
            |> map(fn: (r) => ({v: if r.cond then 10 else 20}))

    testing.diff(want: want, got: got)
}

testcase vec_conditional_float {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{v: -1.0}, {v: 2.0}])

    got =
        array.from(rows: [{cond: true, a: -1.0, b: 0.0}, {cond: false, a: 1.0, b: 2.0}])
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_conditional_float_repeat {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{v: 10.0}, {v: 20.0}])

    got =
        array.from(rows: [{cond: true}, {cond: false}])
            |> map(fn: (r) => ({v: if r.cond then 10.0 else 20.0}))

    testing.diff(want: want, got: got)
}

// FIXME: can't do a vec repeat version for uint until we vectorize uint
testcase vec_conditional_uint {
    expect.planner(rules: ["vectorizeMapRule": 1])

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
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{v: "a"}, {v: "d"}])

    got =
        array.from(rows: [{cond: true, a: "a", b: "b"}, {cond: false, a: "c", b: "d"}])
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_conditional_string_repeat {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{v: "yes"}, {v: "no"}])

    got =
        array.from(rows: [{cond: true}, {cond: false}])
            |> map(fn: (r) => ({v: if r.cond then "yes" else "no"}))

    testing.diff(want: want, got: got)
}

testcase vec_conditional_bool {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{v: true}, {v: false}])

    got =
        array.from(rows: [{cond: true, a: true, b: false}, {cond: false, a: true, b: false}])
            |> map(fn: fn)

    testing.diff(want: want, got: got)
}

testcase vec_conditional_bool_repeat {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{v: true}, {v: false}])

    got =
        array.from(rows: [{cond: true}, {cond: false}])
            |> map(fn: (r) => ({v: if r.cond then true else false}))

    testing.diff(want: want, got: got)
}

testcase vec_conditional_null_test {
    expect.planner(rules: ["vectorizeMapRule": 1])

    // When the condition is null, it's considered false so we should see the
    // value of `r.a` for each record in the output.
    want = array.from(rows: [{x: 0}, {x: 1}, {x: 2}, {x: 3}])

    got =
        array.from(rows: [{a: 0}, {a: 1}, {a: 2}, {a: 3}])
            |> debug.opaque()
            |> map(fn: (r) => ({x: if r.cond then 0 else r.a}))

    testing.diff(want: want, got: got)
}

testcase vec_conditional_null_consequent {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want =
        array.from(
            rows: [
                {x: debug.null(type: "int")},
                {x: debug.null(type: "int")},
                {x: 2},
                {x: debug.null(type: "int")},
            ],
        )

    got =
        array.from(
            rows: [{cond: true, a: 0}, {cond: true, a: 1}, {cond: false, a: 2}, {cond: true, a: 3}],
        )
            |> debug.opaque()
            |> map(fn: (r) => ({x: if r.cond then r.b else r.a}))

    testing.diff(want: want, got: got)
}

testcase vec_conditional_null_alternate {
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{x: 0}, {x: 1}, {x: debug.null(type: "int")}, {x: 3}])

    got =
        array.from(
            rows: [{cond: true, a: 0}, {cond: true, a: 1}, {cond: false, a: 2}, {cond: true, a: 3}],
        )
            |> debug.opaque()
            |> map(fn: (r) => ({x: if r.cond then r.a else r.b}))

    testing.diff(want: want, got: got)
}

testcase vec_conditional_null_consequent_alternate {
    expect.planner(rules: ["vectorizeMapRule": 1])

    // when both branches are invalid, the outcome should be empty
    want = array.from(rows: [{x: 99}]) |> debug.sink()

    got =
        array.from(
            rows: [{cond: true, a: 0}, {cond: true, a: 1}, {cond: false, a: 2}, {cond: true, a: 3}],
        )
            |> debug.opaque()
            |> map(fn: (r) => ({x: if r.cond then r.b else r.c}))

    testing.diff(want: want, got: got)
}

testcase vec_conditional_null_test_consequent_alternate {
    expect.planner(rules: ["vectorizeMapRule": 1])

    // when both branches AND the `test` are invalid, the outcome should be empty
    want = array.from(rows: [{x: 99}]) |> debug.sink()

    got =
        array.from(rows: [{a: 0}, {a: 1}, {a: 2}, {a: 3}])
            |> debug.opaque()
            |> map(fn: (r) => ({x: if r.cond then r.b else r.c}))

    testing.diff(want: want, got: got)
}

testcase vec_nested_logical_conditional_repro {
    // Using a logical expr in the test of the conditional expr produced a
    // runtime error:
    // ```
    // cannot use test of type vector in conditional expression; expected boolean
    // ```
    expect.planner(rules: ["vectorizeMapRule": 1])

    want = array.from(rows: [{a: 0, x: 1}])

    got =
        array.from(rows: [{a: 0}])
            |> map(fn: (r) => ({r with x: if true and true then 1 else 0}))

    testing.diff(want: want, got: got)
}

testcase vec_nested_logical_conditional_repro2 {
    // Using a logical expr in the test of the conditional expr produced a
    // runtime error:
    // ```
    // cannot use test of type vector in conditional expression; expected boolean
    // ```
    expect.planner(rules: ["vectorizeMapRule": 1])

    want =
        array.from(
            rows: [
                {
                    Brake_Code_3: 1,
                    Brake_Code_2: 0,
                    Brake_Code_1: 1,
                    BrakeEmergency: 1,
                    brakeStep: 4.0,
                },
                {
                    Brake_Code_3: 1,
                    Brake_Code_2: 1,
                    Brake_Code_1: 1,
                    BrakeEmergency: 0,
                    brakeStep: 0.0,
                },
                {
                    Brake_Code_3: 1,
                    Brake_Code_2: 0,
                    Brake_Code_1: 1,
                    BrakeEmergency: 0,
                    brakeStep: 2.0,
                },
                {
                    Brake_Code_3: 0,
                    Brake_Code_2: 0,
                    Brake_Code_1: 0,
                    BrakeEmergency: 0,
                    brakeStep: 3.0,
                },
                {
                    Brake_Code_3: 1,
                    Brake_Code_2: 1,
                    Brake_Code_1: 0,
                    BrakeEmergency: 0,
                    brakeStep: 1.0,
                },
            ],
        )

    got =
        array.from(
            rows: [
                {Brake_Code_3: 1, Brake_Code_2: 0, Brake_Code_1: 1, BrakeEmergency: 1},
                {Brake_Code_3: 1, Brake_Code_2: 1, Brake_Code_1: 1, BrakeEmergency: 0},
                {Brake_Code_3: 1, Brake_Code_2: 0, Brake_Code_1: 1, BrakeEmergency: 0},
                {Brake_Code_3: 0, Brake_Code_2: 0, Brake_Code_1: 0, BrakeEmergency: 0},
                {Brake_Code_3: 1, Brake_Code_2: 1, Brake_Code_1: 0, BrakeEmergency: 0},
            ],
        )
            |> map(
                fn: (r) =>
                    ({r with brakeStep:
                            if r.BrakeEmergency == 1 then
                                4.0
                            else if r.Brake_Code_3 == 1 and r.Brake_Code_2 == 1 and r.Brake_Code_1
                                    ==
                                    1 then
                                0.0
                            else if r.Brake_Code_3 == 1 and r.Brake_Code_2 == 1 and r.Brake_Code_1
                                    ==
                                    0 then
                                1.0
                            else if r.Brake_Code_3 == 1 and r.Brake_Code_2 == 0 and r.Brake_Code_1
                                    ==
                                    1 then
                                2.0
                            else if r.Brake_Code_3 == 0 and r.Brake_Code_2 == 0 and r.Brake_Code_1
                                    ==
                                    0 then
                                3.0
                            else
                                0.0,
                    }),
            )

    testing.diff(want: want, got: got)
}
