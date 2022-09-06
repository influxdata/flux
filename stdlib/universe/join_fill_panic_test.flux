package universe_test


import "testing"
import "internal/debug"
import "array"
import "join"

testcase fill_time_previous {
    left =
        array.from(
            rows: [
                {_time: 2022-01-01T00:00:00Z, _value: 2},
                {_time: 2022-01-01T00:00:15Z, _value: 1},
            ],
        )
            |> map(fn: (r) => ({r with myStart: r._time}))

    right = array.from(rows: [{_time: 2022-01-01T00:00:30Z, _value: 0}])
    got =
        join.full(
            left: left,
            right: right,
            on: (l, r) => l._time == r._time,
            as: (l, r) => {
                time = if exists l._time then l._time else r._time
                value = if exists l._value then l._value else r._value

                return {_time: time, value: value, myStart: l.myStart}
            },
        )
            // Adding debug.slurp avoids the panic and the test passes
            // |> debug.slurp()
            |> fill(column: "myStart", usePrevious: true)
    want =
        array.from(
            rows: [
                {_time: 2022-01-01T00:00:00Z, myStart: 2022-01-01T00:00:00Z, value: 2},
                {_time: 2022-01-01T00:00:15Z, myStart: 2022-01-01T00:00:15Z, value: 1},
                {_time: 2022-01-01T00:00:30Z, myStart: 2022-01-01T00:00:15Z, value: 0},
            ],
        )

    testing.diff(got, want)
}
