package testing

import c "csv"

builtin assertEquals
builtin assertEmpty
builtin diff

loadStorage = (csv) => c.from(csv: csv)
loadMem = (csv) => c.from(csv: csv)

run = (name, input, want, testFn) => {
    got = input |> testFn()
    return got |> diff(want: want) |> yield(name: name) |> assertEmpty()
}
