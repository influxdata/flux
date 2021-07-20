package math_test


import "array"
import "math"
import "testing"

xytest = (rows, fn, epsilon=0.000000001, nansEqual=true) => {
    data = array.from(rows: rows)
    got = data
        |> map(fn: (r) => ({_value: fn(x: r.x, y: r.y)}))
    want = data
        |> map(fn: (r) => ({_value: r._value}))

    return testing.diff(got: got, want: want, epsilon: epsilon, nansEqual: nansEqual)
}

testcase atan2 {
    xytest(
        fn: math.atan2,
        rows: [
            // 3 4 5 triangle, 51.13 deg
            {y: 4.0, x: 3.0, _value: 0.9272952180016122},
            // 30 60 90 triangle, 60 deg
            {y: math.sqrt(x: 3.0), x: 1.0, _value: math.pi / 3.0},
            // 30 60 90 triangle, 30 deg
            {y: 1.0, x: math.sqrt(x: 3.0), _value: math.pi / 6.0},
        ],
    )
}
testcase dim {
    xytest(
        fn: math.dim,
        rows: [
            {x: 10.0, y: 5.0, _value: 5.0},
            {x: 10.0, y: 15.0, _value: 0.0},
        ],
    )
}
