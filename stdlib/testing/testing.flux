package testing

import c "csv"

builtin assertEquals
builtin assertEmpty
builtin diff

option loadStorage = (csv) => c.from(csv: csv)
    |> range(start: 1800-01-01T00:00:00Z, stop: 2200-12-31T11:59:59Z)
    |> map(fn: (r) => ({r with
    _field: if exists r._field then r._field else die(msg: "test input table does not have _field column"),
    _measurement: if exists r._measurement then r._measurement else die(msg: "test input table does not have _measurement column"),
    _time: if exists r._time then r._time else die(msg: "test input table does not have _time column")
    }))

option loadMem = (csv) => c.from(csv: csv)

inspect = (case) => {
    tc = case()
    got = tc.input |> tc.fn()
    dif = got |> diff(want: tc.want)
    return {
        fn:    tc.fn,
        input: tc.input,
        want:  tc.want |> yield(name: "want"),
        got:   got |> yield(name: "got"),
        diff:  dif |> yield(name: "diff"),
    }
}

run = (case) => {
    return inspect(case: case).diff |> assertEmpty()
}

benchmark = (case) => {
	tc = case()
	return tc.input |> tc.fn()
}
