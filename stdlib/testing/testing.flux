package testing

import c "csv"

builtin assertEquals
builtin assertEmpty
builtin diff

option loadStorage = (csv) => c.from(csv: csv)
option loadMem = (csv) => c.from(csv: csv)

run = (case) => {
    tc = case()
    return tc.input
        |> tc.fn()
        |> diff(want: tc.want)
        |> yield(name: "diff")
        |> assertEmpty()
}

inspect = (case) => {
    tc = case()
    got = tc.input |> tc.fn()
    dif = got |> diff(want: tc.want)
    pass = dif |> assertEmpty()
    return {
        fn:    tc.fn,
        input: tc.input
        want:  tc.want,
        got:   got,
        diff:  dif,
        pass:  pass,
    }
}