package bitwise_test


import "array"
import "math"
import "experimental/bitwise"
import "testing"

testcase uand_exp {
    cases =
        array.from(
            rows: [
                {a: 1, b: 1, want: 1},
                {a: 1, b: 0, want: 0},
                {a: 5, b: 1, want: 1},
                {a: 5, b: 4, want: 4},
            ],
        )
            |> map(fn: (r) => ({a: uint(v: r.a), b: uint(v: r.b), want: uint(v: r.want)}))

    got =
        cases
            |> map(fn: (r) => ({_value: bitwise.uand(a: r.a, b: r.b)}))

    want =
        cases
            |> map(fn: (r) => ({_value: uint(v: r.want)}))

    testing.diff(want: want, got: got)
}
testcase uor_exp {
    cases =
        array.from(
            rows: [
                {a: 1, b: 1, want: 1},
                {a: 1, b: 0, want: 1},
                {a: 5, b: 1, want: 5},
                {a: 5, b: 4, want: 5},
            ],
        )
            |> map(fn: (r) => ({a: uint(v: r.a), b: uint(v: r.b), want: uint(v: r.want)}))

    got =
        cases
            |> map(fn: (r) => ({_value: bitwise.uor(a: r.a, b: r.b)}))

    want =
        cases
            |> map(fn: (r) => ({_value: uint(v: r.want)}))

    testing.diff(want: want, got: got)
}

testcase unot_exp {
    cases =
        array.from(
            rows: [
                {a: uint(v: 1), want: math.maxuint - uint(v: 1)},
                {a: math.maxuint, want: uint(v: 0)},
            ],
        )

    got =
        cases
            |> map(fn: (r) => ({_value: bitwise.unot(a: r.a)}))

    want =
        cases
            |> map(fn: (r) => ({_value: uint(v: r.want)}))

    testing.diff(want: want, got: got)
}

testcase uclear_exp {
    cases =
        array.from(
            rows: [
                {a: 1, b: 1, want: 0},
                {a: 1, b: 0, want: 1},
                {a: 5, b: 1, want: 4},
                {a: 5, b: 4, want: 1},
            ],
        )
            |> map(fn: (r) => ({a: uint(v: r.a), b: uint(v: r.b), want: uint(v: r.want)}))

    got =
        cases
            |> map(fn: (r) => ({_value: bitwise.uclear(a: r.a, b: r.b)}))

    want =
        cases
            |> map(fn: (r) => ({_value: uint(v: r.want)}))

    testing.diff(want: want, got: got)
}
testcase ulshift_exp {
    cases =
        array.from(
            rows: [
                {a: 1, b: 1, want: 2},
                {a: 1, b: 0, want: 1},
                {a: 5, b: 1, want: 10},
                {a: 5, b: 4, want: 80},
            ],
        )
            |> map(fn: (r) => ({a: uint(v: r.a), b: uint(v: r.b), want: uint(v: r.want)}))

    got =
        cases
            |> map(fn: (r) => ({_value: bitwise.ulshift(a: r.a, b: r.b)}))

    want =
        cases
            |> map(fn: (r) => ({_value: uint(v: r.want)}))

    testing.diff(want: want, got: got)
}
testcase urshift_exp {
    cases =
        array.from(
            rows: [
                {a: 2, b: 1, want: 1},
                {a: 1, b: 0, want: 1},
                {a: 10, b: 1, want: 5},
                {a: 80, b: 4, want: 5},
            ],
        )
            |> map(fn: (r) => ({a: uint(v: r.a), b: uint(v: r.b), want: uint(v: r.want)}))

    got =
        cases
            |> map(fn: (r) => ({_value: bitwise.urshift(a: r.a, b: r.b)}))

    want =
        cases
            |> map(fn: (r) => ({_value: uint(v: r.want)}))

    testing.diff(want: want, got: got)
}

testcase sand_exp {
    cases =
        array.from(
            rows: [
                {a: 1, b: 1, want: 1},
                {a: 1, b: 0, want: 0},
                {a: 5, b: 1, want: 1},
                {a: 5, b: 4, want: 4},
                {a: -1, b: 1, want: 1},
                {a: -1, b: 0, want: 0},
                {a: -5, b: 1, want: 1},
                {a: -5, b: -1, want: -5},
                {a: -5, b: 4, want: 0},
                {a: -5, b: -4, want: -8},
            ],
        )

    got =
        cases
            |> map(fn: (r) => ({_value: bitwise.sand(a: r.a, b: r.b)}))

    want =
        cases
            |> map(fn: (r) => ({_value: r.want}))

    testing.diff(want: want, got: got)
}
testcase sor_exp {
    cases =
        array.from(
            rows: [
                {a: 1, b: 1, want: 1},
                {a: 1, b: 0, want: 1},
                {a: 5, b: 1, want: 5},
                {a: 5, b: 4, want: 5},
            ],
        )

    got =
        cases
            |> map(fn: (r) => ({_value: bitwise.sor(a: r.a, b: r.b)}))

    want =
        cases
            |> map(fn: (r) => ({_value: r.want}))

    testing.diff(want: want, got: got)
}

testcase snot_exp {
    cases = array.from(rows: [{a: 1, want: -2}, {a: math.maxint, want: math.minint}])

    got =
        cases
            |> map(fn: (r) => ({_value: bitwise.snot(a: r.a)}))

    want =
        cases
            |> map(fn: (r) => ({_value: r.want}))

    testing.diff(want: want, got: got)
}

testcase sclear_exp {
    cases =
        array.from(
            rows: [
                {a: 1, b: 1, want: 0},
                {a: 1, b: 0, want: 1},
                {a: 5, b: 1, want: 4},
                {a: 5, b: 4, want: 1},
            ],
        )

    got =
        cases
            |> map(fn: (r) => ({_value: bitwise.sclear(a: r.a, b: r.b)}))

    want =
        cases
            |> map(fn: (r) => ({_value: r.want}))

    testing.diff(want: want, got: got)
}
testcase slshift_exp {
    cases =
        array.from(
            rows: [
                {a: 1, b: 1, want: 2},
                {a: 1, b: 0, want: 1},
                {a: 5, b: 1, want: 10},
                {a: 5, b: 4, want: 80},
            ],
        )

    got =
        cases
            |> map(fn: (r) => ({_value: bitwise.slshift(a: r.a, b: r.b)}))

    want =
        cases
            |> map(fn: (r) => ({_value: r.want}))

    testing.diff(want: want, got: got)
}
testcase srshift_exp {
    cases =
        array.from(
            rows: [
                {a: 2, b: 1, want: 1},
                {a: 1, b: 0, want: 1},
                {a: 10, b: 1, want: 5},
                {a: 80, b: 4, want: 5},
            ],
        )

    got =
        cases
            |> map(fn: (r) => ({_value: bitwise.srshift(a: r.a, b: r.b)}))

    want =
        cases
            |> map(fn: (r) => ({_value: r.want}))

    testing.diff(want: want, got: got)
}
