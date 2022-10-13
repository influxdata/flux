package types_test


import "array"
import "internal/debug"
import "testing"
import "types"

testcase isType {
    testing.assertEqualValues(want: true, got: types.isType(v: "a", type: "string"))
}
testcase isType2 {
    testing.assertEqualValues(want: false, got: types.isType(v: "a", type: "strin"))
}
testcase isType3 {
    testing.assertEqualValues(want: false, got: types.isType(v: "a", type: "int"))
}
testcase isType4 {
    testing.assertEqualValues(want: true, got: types.isType(v: 1, type: "int"))
}
testcase isType5 {
    testing.assertEqualValues(want: true, got: types.isType(v: 2030-01-01T00:00:00Z, type: "time"))
}
testcase isType6 {
    testing.assertEqualValues(want: false, got: types.isType(v: 2030-01-01T00:00:00Z, type: "int"))
}

tableDataF0 =
    array.from(
        rows: [
            {_measurement: "m", _field: "f0", _value: "cat", _time: 2021-12-16T00:00:00Z},
            {_measurement: "m", _field: "f0", _value: "dog", _time: 2021-12-16T00:01:00Z},
            {_measurement: "m", _field: "f0", _value: "emu", _time: 2021-12-16T00:02:00Z},
            {_measurement: "m", _field: "f0", _value: "rat", _time: 2021-12-16T00:03:00Z},
            {_measurement: "m", _field: "f0", _value: "bird", _time: 2021-12-16T00:04:00Z},
        ],
    )
        |> group(columns: ["_measurement", "_field"])
        |> debug.opaque()

tableDataF1 =
    array.from(
        rows: [
            {_measurement: "m", _field: "f1", _value: 10, _time: 2021-12-16T00:00:00Z},
            {_measurement: "m", _field: "f1", _value: 11, _time: 2021-12-16T00:01:00Z},
            {_measurement: "m", _field: "f1", _value: 12, _time: 2021-12-16T00:02:00Z},
            {_measurement: "m", _field: "f1", _value: 13, _time: 2021-12-16T00:03:00Z},
            {_measurement: "m", _field: "f1", _value: 14, _time: 2021-12-16T00:04:00Z},
        ],
    )
        |> group(columns: ["_measurement", "_field"])
        |> debug.opaque()

tableDataF2 =
    array.from(
        rows: [
            {_measurement: "m", _field: "f2", _value: uint(v: 10), _time: 2021-12-16T00:00:00Z},
            {_measurement: "m", _field: "f2", _value: uint(v: 11), _time: 2021-12-16T00:01:00Z},
            {_measurement: "m", _field: "f2", _value: uint(v: 12), _time: 2021-12-16T00:02:00Z},
            {_measurement: "m", _field: "f2", _value: uint(v: 13), _time: 2021-12-16T00:03:00Z},
            {_measurement: "m", _field: "f2", _value: uint(v: 14), _time: 2021-12-16T00:04:00Z},
        ],
    )
        |> group(columns: ["_measurement", "_field"])
        |> debug.opaque()

tableDataF3 =
    array.from(
        rows: [
            {_measurement: "m", _field: "f3", _value: 10.0, _time: 2021-12-16T00:00:00Z},
            {_measurement: "m", _field: "f3", _value: 11.0, _time: 2021-12-16T00:01:00Z},
            {_measurement: "m", _field: "f3", _value: 12.0, _time: 2021-12-16T00:02:00Z},
            {_measurement: "m", _field: "f3", _value: 13.0, _time: 2021-12-16T00:03:00Z},
            {_measurement: "m", _field: "f3", _value: 14.0, _time: 2021-12-16T00:04:00Z},
        ],
    )
        |> group(columns: ["_measurement", "_field"])
        |> debug.opaque()

tableDataF4 =
    array.from(
        rows: [
            {_measurement: "m", _field: "f4", _value: true, _time: 2021-12-16T00:00:00Z},
            {_measurement: "m", _field: "f4", _value: false, _time: 2021-12-16T00:01:00Z},
            {_measurement: "m", _field: "f4", _value: true, _time: 2021-12-16T00:02:00Z},
            {_measurement: "m", _field: "f4", _value: false, _time: 2021-12-16T00:03:00Z},
            {_measurement: "m", _field: "f4", _value: true, _time: 2021-12-16T00:04:00Z},
        ],
    )
        |> group(columns: ["_measurement", "_field"])
        |> debug.opaque()

tableData =
    union(
        tables: [
            tableDataF0,
            tableDataF1,
            tableDataF2,
            tableDataF3,
            tableDataF4,
        ],
    )

numericTableData = union(tables: [tableDataF1, tableDataF2, tableDataF3])

testcase isTypeTableString {
    want = tableDataF0
    got =
        testing.load(tables: tableData)
            |> range(start: -100)
            |> filter(fn: (r) => types.isType(v: r._value, type: "string"))
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase isTypeTableInt {
    want = tableDataF1
    got =
        testing.load(tables: tableData)
            |> range(start: -100)
            |> filter(fn: (r) => types.isType(v: r._value, type: "int"))
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase isTypeTableUint {
    want = tableDataF2
    got =
        testing.load(tables: tableData)
            |> range(start: -100)
            |> filter(fn: (r) => types.isType(v: r._value, type: "uint"))
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase isTypeTableFloat {
    want = tableDataF3
    got =
        testing.load(tables: tableData)
            |> range(start: -100)
            |> filter(fn: (r) => types.isType(v: r._value, type: "float"))
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase isTypeTableBool {
    want = tableDataF4
    got =
        testing.load(tables: tableData)
            |> range(start: -100)
            |> filter(fn: (r) => types.isType(v: r._value, type: "bool"))
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}

testcase isTypeTableNumeric {
    want = numericTableData
    got =
        testing.load(tables: tableData)
            |> range(start: -100)
            |> filter(fn: (r) => types.isNumeric(v: r._value))
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want) |> yield()
}
