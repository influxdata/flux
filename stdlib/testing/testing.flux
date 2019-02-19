package testing

import c "csv"

builtin assertEquals
builtin assertEmpty
builtin diff

option loadStorage = (csv) => c.from(csv: csv)
option loadMem = (csv) => c.from(csv: csv)

inspect = (case) => {
    tc = case()
    got = tc.input |> tc.fn() |> yield(name: "_test_result")
    dif = got |> diff(want: tc.want) |> yield(name: "diff")
    return {
        fn:    tc.fn,
        input: tc.input
        want:  tc.want,
        got:   got,
        diff:  dif,
    }
}

run = (case) => {
    return inspect(case: case).diff |> assertEmpty()
}

