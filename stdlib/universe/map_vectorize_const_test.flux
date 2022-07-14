package universe_test


import "array"
import "testing"

// Vectorized constant values, aka "vec repeat" are vectors that only hold a
// singleton value, unchanged during the processing of all rows.
// Special care must be given to how these interact with array-backed vectors
// making them sensitive to the ordering of vectorized expressions.
//
// The base cases targeted here are record extension adding a new column using:
// - a plain constant value.
// - a binary expression (addition) with 2 constants (aka "constant folding")
// - a binary expression (addition) with a constant and a record member.
// - nested binary expressions (addition) with constants and record members in
//   different orderings.
data = array.from(rows: [{x: 1}, {x: 2}])

testcase vec_const_with_const {
    want = array.from(rows: [{x: 1, y: 5}, {x: 2, y: 5}])
    got = data |> map(fn: (r) => ({r with y: 5}))

    testing.diff(want: want, got: got)
}

testcase vec_const_with_const_add_const {
    want = array.from(rows: [{x: 1, y: 7}, {x: 2, y: 7}])
    got = data |> map(fn: (r) => ({r with y: 5 + 2}))

    testing.diff(want: want, got: got)
}

testcase vec_const_add_member_const {
    want = array.from(rows: [{x: 1, y: 6}, {x: 2, y: 7}])
    got = data |> map(fn: (r) => ({r with y: r.x + 5}))

    testing.diff(want: want, got: got)
}

testcase vec_const_with_const_add_const_add_member {
    want = array.from(rows: [{x: 1, y: 7}, {x: 2, y: 8}])
    got = data |> map(fn: (r) => ({r with y: 5 + 1 + r.x}))

    testing.diff(want: want, got: got)
}

testcase vec_const_with_const_add_member_add_const {
    want = array.from(rows: [{x: 1, y: 7}, {x: 2, y: 8}])
    got = data |> map(fn: (r) => ({r with y: 5 + r.x + 1}))

    testing.diff(want: want, got: got)
}

testcase vec_const_with_member_add_const_add_const {
    want = array.from(rows: [{x: 1, y: 7}, {x: 2, y: 8}])
    got = data |> map(fn: (r) => ({r with y: r.x + 5 + 1}))

    testing.diff(want: want, got: got)
}

// Test to check that a range of literals are supported
testcase vec_const_kitchen_sink_column_types {
    want =
        array.from(
            rows: [
                {
                    x: 1,
                    i: 99,
                    f: 1.23,
                    t: 1985-10-26T00:00:00Z,
                    s: "flux rules",
                },
                {
                    x: 2,
                    i: 99,
                    f: 1.23,
                    t: 1985-10-26T00:00:00Z,
                    s: "flux rules",
                },
            ],
        )
    got =
        data
            |> map(fn: (r) => ({r with i: 99, f: 1.23, t: 1985-10-26T00:00:00Z, s: "flux rules"}))

    testing.diff(want: want, got: got)
}

testcase vec_const_bools {
    option testing.tags = [
        // FIXME: https://github.com/influxdata/flux/issues/4997
        //  bool literals are not vectorized currently
        "skip",
    ]

    input = array.from(rows: [{a: false, b: false}, {a: false, b: true}, {a: true, b: true}])
    want =
        array.from(
            rows: [
                {
                    a: false,
                    a_and_true: false,
                    a_or_true: true,
                    a_and_false: false,
                    a_or_false: false,
                },
                {
                    a: false,
                    a_and_true: false,
                    a_or_true: true,
                    a_and_false: false,
                    a_or_false: false,
                },
                {
                    a: true,
                    a_and_true: true,
                    a_or_true: true,
                    a_and_false: false,
                    a_or_false: true,
                },
            ],
        )
    got =
        input
            |> map(
                fn: (r) =>
                    ({r with a_and_true: r.a and true,
                        a_or_true: r.a or true,
                        a_and_false: r.a and false,
                        a_or_false: r.a or false,
                    }),
            )

    testing.diff(want: want, got: got)
}
