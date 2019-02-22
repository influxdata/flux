package testing

import c "csv"

builtin assertEquals
builtin assertEmpty
builtin diff

option loadStorage = (csv) => c.from(csv: csv)
option loadMem = (csv) => c.from(csv: csv)

inspect = (case) => {
    tc = case()
    got = tc.input |> tc.fn()
    dif = got |> diff(want: tc.want)
    return {
        fn:    tc.fn,
        input: tc.input
        want:  tc.want |> yield(name: "want"),
        got:   got |> yield(name: "got"),
        diff:  dif |> yield(name: "diff"),
    }
}

run = (case) => {
    return inspect(case: case).diff |> assertEmpty()
}

