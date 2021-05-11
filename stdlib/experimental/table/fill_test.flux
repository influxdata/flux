package table_test


import "array"
import "planner"
import "testing"
import "testing/expect"
import "experimental/table"

inData = array.from(
    rows: [
        {_time: 2021-04-13T09:00:00Z, _measurement: "m0", _field: "f0", _value: 2.0, t0: "a"},
        {_time: 2021-04-13T09:05:00Z, _measurement: "m0", _field: "f0", _value: 9.0, t0: "a"},
        {_time: 2021-04-13T09:15:00Z, _measurement: "m0", _field: "f0", _value: 7.0, t0: "a"},
        {_time: 2021-04-13T09:25:00Z, _measurement: "m0", _field: "f0", _value: 3.0, t0: "a"},
        {_time: 2021-04-13T09:45:00Z, _measurement: "m0", _field: "f0", _value: 5.0, t0: "a"},
        {_time: 2021-04-13T09:55:00Z, _measurement: "m0", _field: "f0", _value: 1.0, t0: "a"},
        {_time: 2021-04-13T09:05:00Z, _measurement: "m0", _field: "f0", _value: 4.0, t0: "b"},
        {_time: 2021-04-13T09:10:00Z, _measurement: "m0", _field: "f0", _value: 1.0, t0: "b"},
        {_time: 2021-04-13T09:15:00Z, _measurement: "m0", _field: "f0", _value: 2.0, t0: "b"},
        {_time: 2021-04-13T09:30:00Z, _measurement: "m0", _field: "f0", _value: 5.0, t0: "b"},
        {_time: 2021-04-13T09:35:00Z, _measurement: "m0", _field: "f0", _value: 6.0, t0: "b"},
        {_time: 2021-04-13T09:40:00Z, _measurement: "m0", _field: "f0", _value: 8.0, t0: "b"},
    ],
)
    |> group(columns: ["_measurement", "_field", "t0"])
loadData = () => inData
    |> testing.load()
    |> range(start: 2021-04-13T09:00:00Z, stop: 2021-04-13T10:00:00Z)

testcase window {
    want = testing.loadMem(
        csv: "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double,string
#group,false,false,true,true,false,true,true,false,true
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_measurement,_field,_value,t0
,,0,2021-04-13T09:00:00Z,2021-04-13T09:15:00Z,2021-04-13T09:00:00Z,m0,f0,2.0,a
,,0,2021-04-13T09:00:00Z,2021-04-13T09:15:00Z,2021-04-13T09:05:00Z,m0,f0,9.0,a
,,1,2021-04-13T09:15:00Z,2021-04-13T09:30:00Z,2021-04-13T09:15:00Z,m0,f0,7.0,a
,,1,2021-04-13T09:15:00Z,2021-04-13T09:30:00Z,2021-04-13T09:25:00Z,m0,f0,3.0,a
,,2,2021-04-13T09:30:00Z,2021-04-13T09:45:00Z,,m0,f0,,a
,,3,2021-04-13T09:45:00Z,2021-04-13T10:00:00Z,2021-04-13T09:45:00Z,m0,f0,5.0,a
,,3,2021-04-13T09:45:00Z,2021-04-13T10:00:00Z,2021-04-13T09:55:00Z,m0,f0,1.0,a
,,4,2021-04-13T09:00:00Z,2021-04-13T09:15:00Z,2021-04-13T09:05:00Z,m0,f0,4.0,b
,,4,2021-04-13T09:00:00Z,2021-04-13T09:15:00Z,2021-04-13T09:10:00Z,m0,f0,1.0,b
,,5,2021-04-13T09:15:00Z,2021-04-13T09:30:00Z,2021-04-13T09:15:00Z,m0,f0,2.0,b
,,6,2021-04-13T09:30:00Z,2021-04-13T09:45:00Z,2021-04-13T09:30:00Z,m0,f0,5.0,b
,,6,2021-04-13T09:30:00Z,2021-04-13T09:45:00Z,2021-04-13T09:35:00Z,m0,f0,6.0,b
,,6,2021-04-13T09:30:00Z,2021-04-13T09:45:00Z,2021-04-13T09:40:00Z,m0,f0,8.0,b
,,7,2021-04-13T09:45:00Z,2021-04-13T10:00:00Z,,m0,f0,,b
",
    )
    got = loadData()
        |> window(every: 15m, createEmpty: true)
        |> table.fill()

    testing.diff(got, want) |> yield()
}
testcase selector_fill {
    want = array.from(
        rows: [
            {_start: 2021-04-13T09:00:00Z, _stop: 2021-04-13T10:00:00Z, _time: 2021-04-13T09:15:00Z, _measurement: "m0", _field: "f0", _value: 2.0, t0: "a"},
            {_start: 2021-04-13T09:00:00Z, _stop: 2021-04-13T10:00:00Z, _time: 2021-04-13T09:30:00Z, _measurement: "m0", _field: "f0", _value: 3.0, t0: "a"},
            {_start: 2021-04-13T09:00:00Z, _stop: 2021-04-13T10:00:00Z, _time: 2021-04-13T09:45:00Z, _measurement: "m0", _field: "f0", _value: 0.0, t0: "a"},
            {_start: 2021-04-13T09:00:00Z, _stop: 2021-04-13T10:00:00Z, _time: 2021-04-13T10:00:00Z, _measurement: "m0", _field: "f0", _value: 1.0, t0: "a"},
            {_start: 2021-04-13T09:00:00Z, _stop: 2021-04-13T10:00:00Z, _time: 2021-04-13T09:15:00Z, _measurement: "m0", _field: "f0", _value: 1.0, t0: "b"},
            {_start: 2021-04-13T09:00:00Z, _stop: 2021-04-13T10:00:00Z, _time: 2021-04-13T09:30:00Z, _measurement: "m0", _field: "f0", _value: 2.0, t0: "b"},
            {_start: 2021-04-13T09:00:00Z, _stop: 2021-04-13T10:00:00Z, _time: 2021-04-13T09:45:00Z, _measurement: "m0", _field: "f0", _value: 5.0, t0: "b"},
            {_start: 2021-04-13T09:00:00Z, _stop: 2021-04-13T10:00:00Z, _time: 2021-04-13T10:00:00Z, _measurement: "m0", _field: "f0", _value: 0.0, t0: "b"},
        ],
    )
        |> group(
            columns: [
                "_start",
                "_stop",
                "_measurement",
                "_field",
                "t0",
            ],
        )
    got = loadData()
        |> window(every: 15m, createEmpty: true)
        |> min()
        |> table.fill()
        |> duplicate(column: "_stop", as: "_time")
        |> window(every: inf)
        |> fill(value: 0.0)

    testing.diff(got, want) |> yield()
}

test_idempotent = () => {
    want = testing.loadMem(
        csv: "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double,string
#group,false,false,true,true,false,true,true,false,true
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_measurement,_field,_value,t0
,,0,2021-04-13T09:00:00Z,2021-04-13T09:15:00Z,2021-04-13T09:00:00Z,m0,f0,2.0,a
,,0,2021-04-13T09:00:00Z,2021-04-13T09:15:00Z,2021-04-13T09:05:00Z,m0,f0,9.0,a
,,1,2021-04-13T09:15:00Z,2021-04-13T09:30:00Z,2021-04-13T09:15:00Z,m0,f0,7.0,a
,,1,2021-04-13T09:15:00Z,2021-04-13T09:30:00Z,2021-04-13T09:25:00Z,m0,f0,3.0,a
,,2,2021-04-13T09:30:00Z,2021-04-13T09:45:00Z,,m0,f0,,a
,,3,2021-04-13T09:45:00Z,2021-04-13T10:00:00Z,2021-04-13T09:45:00Z,m0,f0,5.0,a
,,3,2021-04-13T09:45:00Z,2021-04-13T10:00:00Z,2021-04-13T09:55:00Z,m0,f0,1.0,a
,,4,2021-04-13T09:00:00Z,2021-04-13T09:15:00Z,2021-04-13T09:05:00Z,m0,f0,4.0,b
,,4,2021-04-13T09:00:00Z,2021-04-13T09:15:00Z,2021-04-13T09:10:00Z,m0,f0,1.0,b
,,5,2021-04-13T09:15:00Z,2021-04-13T09:30:00Z,2021-04-13T09:15:00Z,m0,f0,2.0,b
,,6,2021-04-13T09:30:00Z,2021-04-13T09:45:00Z,2021-04-13T09:30:00Z,m0,f0,5.0,b
,,6,2021-04-13T09:30:00Z,2021-04-13T09:45:00Z,2021-04-13T09:35:00Z,m0,f0,6.0,b
,,6,2021-04-13T09:30:00Z,2021-04-13T09:45:00Z,2021-04-13T09:40:00Z,m0,f0,8.0,b
,,7,2021-04-13T09:45:00Z,2021-04-13T10:00:00Z,,m0,f0,,b
",
    )
    got = loadData()
        |> window(every: 15m, createEmpty: true)
        |> table.fill()
        |> table.fill()

    return testing.diff(got, want) |> yield()
}

testcase idempotent {
    option planner.disableLogicalRules = ["experimental/table.IdempotentTableFill"]

    expect.planner(rules: ["experimental/table.IdempotentTableFill": 0])
    test_idempotent()
}
testcase idempotent_planner_rule {
    expect.planner(rules: ["experimental/table.IdempotentTableFill": 1])
    test_idempotent()
}
