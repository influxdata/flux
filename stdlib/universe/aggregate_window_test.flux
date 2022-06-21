package universe_test


import "array"
import "csv"
import "testing"

do_test = (every, fn) =>
    array.from(
        rows: [
            {
                _time: 2019-11-25T00:00:00Z,
                _measurement: "m0",
                _field: "f0",
                t0: "a-0",
                _value: 1.0,
            },
            {
                _time: 2019-11-25T00:00:15Z,
                _measurement: "m0",
                _field: "f0",
                t0: "a-0",
                _value: 2.0,
            },
            {
                _time: 2019-11-25T00:00:30Z,
                _measurement: "m0",
                _field: "f0",
                t0: "a-0",
                _value: 3.0,
            },
            {
                _time: 2019-11-25T00:00:45Z,
                _measurement: "m0",
                _field: "f0",
                t0: "a-0",
                _value: 4.0,
            },
            {
                _time: 2019-11-25T00:00:00Z,
                _measurement: "m0",
                _field: "f0",
                t0: "a-1",
                _value: 1.0,
            },
            {
                _time: 2019-11-25T00:00:15Z,
                _measurement: "m0",
                _field: "f0",
                t0: "a-1",
                _value: 2.0,
            },
            {
                _time: 2019-11-25T00:00:30Z,
                _measurement: "m0",
                _field: "f0",
                t0: "a-1",
                _value: 3.0,
            },
            {
                _time: 2019-11-25T00:00:45Z,
                _measurement: "m0",
                _field: "f0",
                t0: "a-1",
                _value: 4.0,
            },
            {
                _time: 2019-11-25T00:00:00Z,
                _measurement: "m0",
                _field: "f0",
                t0: "a-2",
                _value: 1.0,
            },
            {
                _time: 2019-11-25T00:00:15Z,
                _measurement: "m0",
                _field: "f0",
                t0: "a-2",
                _value: 2.0,
            },
            {
                _time: 2019-11-25T00:00:30Z,
                _measurement: "m0",
                _field: "f0",
                t0: "a-2",
                _value: 3.0,
            },
            {
                _time: 2019-11-25T00:00:45Z,
                _measurement: "m0",
                _field: "f0",
                t0: "a-2",
                _value: 4.0,
            },
        ],
    )
        |> group(columns: ["_measurement", "_field", "t0"])
        |> testing.load()
        |> range(start: 2019-11-25T00:00:00Z, stop: 2019-11-25T00:01:00Z)
        |> aggregateWindow(every, fn)
        |> drop(columns: ["_start", "_stop"])

testcase count_with_nulls {
    want =
        array.from(
            rows: [
                {
                    _time: 2019-11-25T00:00:10Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-0",
                    _value: 1,
                },
                {
                    _time: 2019-11-25T00:00:20Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-0",
                    _value: 1,
                },
                {
                    _time: 2019-11-25T00:00:30Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-0",
                    _value: 0,
                },
                {
                    _time: 2019-11-25T00:00:40Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-0",
                    _value: 1,
                },
                {
                    _time: 2019-11-25T00:00:50Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-0",
                    _value: 1,
                },
                {
                    _time: 2019-11-25T00:01:00Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-0",
                    _value: 0,
                },
                {
                    _time: 2019-11-25T00:00:10Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-1",
                    _value: 1,
                },
                {
                    _time: 2019-11-25T00:00:20Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-1",
                    _value: 1,
                },
                {
                    _time: 2019-11-25T00:00:30Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-1",
                    _value: 0,
                },
                {
                    _time: 2019-11-25T00:00:40Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-1",
                    _value: 1,
                },
                {
                    _time: 2019-11-25T00:00:50Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-1",
                    _value: 1,
                },
                {
                    _time: 2019-11-25T00:01:00Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-1",
                    _value: 0,
                },
                {
                    _time: 2019-11-25T00:00:10Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-2",
                    _value: 1,
                },
                {
                    _time: 2019-11-25T00:00:20Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-2",
                    _value: 1,
                },
                {
                    _time: 2019-11-25T00:00:30Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-2",
                    _value: 0,
                },
                {
                    _time: 2019-11-25T00:00:40Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-2",
                    _value: 1,
                },
                {
                    _time: 2019-11-25T00:00:50Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-2",
                    _value: 1,
                },
                {
                    _time: 2019-11-25T00:01:00Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-2",
                    _value: 0,
                },
            ],
        )
            |> group(columns: ["_measurement", "_field", "t0"])
    got = do_test(every: 10s, fn: count)

    testing.diff(got, want) |> yield()
}
testcase min_with_nulls {
    want =
        csv.from(
            csv:
                "#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,t0,_value
,,0,2019-11-25T00:00:10Z,m0,f0,a-0,1.0
,,0,2019-11-25T00:00:20Z,m0,f0,a-0,2.0
,,0,2019-11-25T00:00:30Z,m0,f0,a-0,
,,0,2019-11-25T00:00:40Z,m0,f0,a-0,3.0
,,0,2019-11-25T00:00:50Z,m0,f0,a-0,4.0
,,0,2019-11-25T00:01:00Z,m0,f0,a-0,
,,1,2019-11-25T00:00:10Z,m0,f0,a-1,1.0
,,1,2019-11-25T00:00:20Z,m0,f0,a-1,2.0
,,1,2019-11-25T00:00:30Z,m0,f0,a-1,
,,1,2019-11-25T00:00:40Z,m0,f0,a-1,3.0
,,1,2019-11-25T00:00:50Z,m0,f0,a-1,4.0
,,1,2019-11-25T00:01:00Z,m0,f0,a-1,
,,2,2019-11-25T00:00:10Z,m0,f0,a-2,1.0
,,2,2019-11-25T00:00:20Z,m0,f0,a-2,2.0
,,2,2019-11-25T00:00:30Z,m0,f0,a-2,
,,2,2019-11-25T00:00:40Z,m0,f0,a-2,3.0
,,2,2019-11-25T00:00:50Z,m0,f0,a-2,4.0
,,2,2019-11-25T00:01:00Z,m0,f0,a-2,
",
        )
    got = do_test(every: 10s, fn: min)

    testing.diff(got, want) |> yield()
}
testcase max_with_nulls {
    want =
        csv.from(
            csv:
                "#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,t0,_value
,,0,2019-11-25T00:00:10Z,m0,f0,a-0,1.0
,,0,2019-11-25T00:00:20Z,m0,f0,a-0,2.0
,,0,2019-11-25T00:00:30Z,m0,f0,a-0,
,,0,2019-11-25T00:00:40Z,m0,f0,a-0,3.0
,,0,2019-11-25T00:00:50Z,m0,f0,a-0,4.0
,,0,2019-11-25T00:01:00Z,m0,f0,a-0,
,,1,2019-11-25T00:00:10Z,m0,f0,a-1,1.0
,,1,2019-11-25T00:00:20Z,m0,f0,a-1,2.0
,,1,2019-11-25T00:00:30Z,m0,f0,a-1,
,,1,2019-11-25T00:00:40Z,m0,f0,a-1,3.0
,,1,2019-11-25T00:00:50Z,m0,f0,a-1,4.0
,,1,2019-11-25T00:01:00Z,m0,f0,a-1,
,,2,2019-11-25T00:00:10Z,m0,f0,a-2,1.0
,,2,2019-11-25T00:00:20Z,m0,f0,a-2,2.0
,,2,2019-11-25T00:00:30Z,m0,f0,a-2,
,,2,2019-11-25T00:00:40Z,m0,f0,a-2,3.0
,,2,2019-11-25T00:00:50Z,m0,f0,a-2,4.0
,,2,2019-11-25T00:01:00Z,m0,f0,a-2,
",
        )
    got = do_test(every: 10s, fn: max)

    testing.diff(got, want) |> yield()
}
