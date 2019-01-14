package testing

import c "csv"

loadStorage = (csv) => c.from(csv: csv)
loadMem = (csv) => c.from(csv: csv)

test = (name, input, want, testFn) => {
    got = input |> testFn()
    return assertEquals(name: name, want: want, got: got)
}