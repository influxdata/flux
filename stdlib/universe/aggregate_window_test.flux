package universe_test


import "array"
import "csv"
import "internal/debug"
import "testing"
import "testing/expect"
import "planner"

sampleData = [
    {_time: 2019-11-25T00:00:00Z, t0: "a-0", _value: 1.0},
    {_time: 2019-11-25T00:00:05Z, t0: "a-0", _value: 5.0},
    {_time: 2019-11-25T00:00:15Z, t0: "a-0", _value: 2.0},
    {_time: 2019-11-25T00:00:30Z, t0: "a-0", _value: 3.0},
    {_time: 2019-11-25T00:00:45Z, t0: "a-0", _value: 4.0},
    {_time: 2019-11-25T00:00:00Z, t0: "a-1", _value: 1.0},
    {_time: 2019-11-25T00:00:05Z, t0: "a-1", _value: 5.0},
    {_time: 2019-11-25T00:00:15Z, t0: "a-1", _value: 2.0},
    {_time: 2019-11-25T00:00:30Z, t0: "a-1", _value: 3.0},
    {_time: 2019-11-25T00:00:45Z, t0: "a-1", _value: 4.0},
    {_time: 2019-11-25T00:00:00Z, t0: "a-2", _value: 1.0},
    {_time: 2019-11-25T00:00:05Z, t0: "a-2", _value: 5.0},
    {_time: 2019-11-25T00:00:15Z, t0: "a-2", _value: 2.0},
    {_time: 2019-11-25T00:00:30Z, t0: "a-2", _value: 3.0},
    {_time: 2019-11-25T00:00:45Z, t0: "a-2", _value: 4.0},
]

do_test = (every, fn) =>
    array.from(rows: sampleData)
        |> map(fn: (r) => ({r with _measurement: "m0", _field: "f0"}))
        |> group(columns: ["_measurement", "_field", "t0"])
        |> testing.load()
        |> range(start: 2019-11-25T00:00:00Z, stop: 2019-11-25T00:01:00Z)
        |> aggregateWindow(every: every, fn: fn, timeSrc: "_start")
        |> drop(columns: ["_start", "_stop"])

testcase count_empty_windows {
    want =
        array.from(
            rows: [
                {
                    _time: 2019-11-25T00:00:00Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-0",
                    _value: 2,
                },
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
                    _value: 0,
                },
                {
                    _time: 2019-11-25T00:00:30Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-0",
                    _value: 1,
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
                    _value: 0,
                },
                {
                    _time: 2019-11-25T00:00:00Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-1",
                    _value: 2,
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
                    _value: 0,
                },
                {
                    _time: 2019-11-25T00:00:30Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-1",
                    _value: 1,
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
                    _value: 0,
                },
                {
                    _time: 2019-11-25T00:00:00Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-2",
                    _value: 2,
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
                    _value: 0,
                },
                {
                    _time: 2019-11-25T00:00:30Z,
                    _measurement: "m0",
                    _field: "f0",
                    t0: "a-2",
                    _value: 1,
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
                    _value: 0,
                },
            ],
        )
            |> group(columns: ["_measurement", "_field", "t0"])
    got = do_test(every: 10s, fn: count)

    testing.diff(got, want)
}

testcase count_with_nulls {
    want =
        array.from(
            rows: [
                {_time: 2019-11-25T00:00:00Z, t0: "a-0", _value: 6},
                {_time: 2019-11-25T00:00:00Z, t0: "a-1", _value: 6},
                {_time: 2019-11-25T00:00:00Z, t0: "a-2", _value: 6},
            ],
        )
            |> map(fn: (r) => ({r with _measurement: "m0", _field: "f0"}))
            |> group(columns: ["_measurement", "_field", "t0"])
    got =
        do_test(every: 10s, fn: sum)
            |> aggregateWindow(every: 1m, fn: count, timeSrc: "_start")
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want)
}

testcase count_null_windows {
    want =
        array.from(
            rows: [
                {_time: 2019-11-25T00:00:00Z, _value: 3},
                {_time: 2019-11-25T00:00:10Z, _value: 3},
                {_time: 2019-11-25T00:00:20Z, _value: 3},
                {_time: 2019-11-25T00:00:30Z, _value: 3},
                {_time: 2019-11-25T00:00:40Z, _value: 3},
                {_time: 2019-11-25T00:00:50Z, _value: 3},
            ],
        )
            |> map(fn: (r) => ({r with _measurement: "m0", _field: "f0"}))
            |> group(columns: ["_measurement", "_field"])
    got =
        do_test(every: 10s, fn: sum)
            |> group(columns: ["_measurement", "_field"])
            |> aggregateWindow(every: 10s, fn: count, timeSrc: "_start")
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want)
}

testcase sum_empty_windows {
    want =
        csv.from(
            csv:
                "#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,t0,_value
,,0,2019-11-25T00:00:00Z,m0,f0,a-0,6.0
,,0,2019-11-25T00:00:10Z,m0,f0,a-0,2.0
,,0,2019-11-25T00:00:20Z,m0,f0,a-0,
,,0,2019-11-25T00:00:30Z,m0,f0,a-0,3.0
,,0,2019-11-25T00:00:40Z,m0,f0,a-0,4.0
,,0,2019-11-25T00:00:50Z,m0,f0,a-0,
,,1,2019-11-25T00:00:00Z,m0,f0,a-1,6.0
,,1,2019-11-25T00:00:10Z,m0,f0,a-1,2.0
,,1,2019-11-25T00:00:20Z,m0,f0,a-1,
,,1,2019-11-25T00:00:30Z,m0,f0,a-1,3.0
,,1,2019-11-25T00:00:40Z,m0,f0,a-1,4.0
,,1,2019-11-25T00:00:50Z,m0,f0,a-1,
,,2,2019-11-25T00:00:00Z,m0,f0,a-2,6.0
,,2,2019-11-25T00:00:10Z,m0,f0,a-2,2.0
,,2,2019-11-25T00:00:20Z,m0,f0,a-2,
,,2,2019-11-25T00:00:30Z,m0,f0,a-2,3.0
,,2,2019-11-25T00:00:40Z,m0,f0,a-2,4.0
,,2,2019-11-25T00:00:50Z,m0,f0,a-2,
",
        )
    got = do_test(every: 10s, fn: sum)

    testing.diff(got, want)
}

testcase sum_with_nulls {
    want =
        array.from(
            rows: [
                {_time: 2019-11-25T00:00:00Z, t0: "a-0", _value: 15.0},
                {_time: 2019-11-25T00:00:00Z, t0: "a-1", _value: 15.0},
                {_time: 2019-11-25T00:00:00Z, t0: "a-2", _value: 15.0},
            ],
        )
            |> map(fn: (r) => ({r with _measurement: "m0", _field: "f0"}))
            |> group(columns: ["_measurement", "_field", "t0"])
    got =
        do_test(every: 10s, fn: sum)
            |> aggregateWindow(every: 1m, fn: sum, timeSrc: "_start")
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want)
}

testcase sum_null_windows {
    want =
        csv.from(
            csv:
                "#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2019-11-25T00:00:00Z,m0,f0,18.0
,,0,2019-11-25T00:00:10Z,m0,f0,6.0
,,0,2019-11-25T00:00:20Z,m0,f0,
,,0,2019-11-25T00:00:30Z,m0,f0,9.0
,,0,2019-11-25T00:00:40Z,m0,f0,12.0
,,0,2019-11-25T00:00:50Z,m0,f0,
",
        )
    got =
        do_test(every: 10s, fn: sum)
            |> group(columns: ["_measurement", "_field"])
            |> aggregateWindow(every: 10s, fn: sum, timeSrc: "_start")
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want)
}

testcase mean_empty_windows {
    want =
        csv.from(
            csv:
                "#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,t0,_value
,,0,2019-11-25T00:00:00Z,m0,f0,a-0,3.0
,,0,2019-11-25T00:00:10Z,m0,f0,a-0,2.0
,,0,2019-11-25T00:00:20Z,m0,f0,a-0,
,,0,2019-11-25T00:00:30Z,m0,f0,a-0,3.0
,,0,2019-11-25T00:00:40Z,m0,f0,a-0,4.0
,,0,2019-11-25T00:00:50Z,m0,f0,a-0,
,,1,2019-11-25T00:00:00Z,m0,f0,a-1,3.0
,,1,2019-11-25T00:00:10Z,m0,f0,a-1,2.0
,,1,2019-11-25T00:00:20Z,m0,f0,a-1,
,,1,2019-11-25T00:00:30Z,m0,f0,a-1,3.0
,,1,2019-11-25T00:00:40Z,m0,f0,a-1,4.0
,,1,2019-11-25T00:00:50Z,m0,f0,a-1,
,,2,2019-11-25T00:00:00Z,m0,f0,a-2,3.0
,,2,2019-11-25T00:00:10Z,m0,f0,a-2,2.0
,,2,2019-11-25T00:00:20Z,m0,f0,a-2,
,,2,2019-11-25T00:00:30Z,m0,f0,a-2,3.0
,,2,2019-11-25T00:00:40Z,m0,f0,a-2,4.0
,,2,2019-11-25T00:00:50Z,m0,f0,a-2,
",
        )
    got = do_test(every: 10s, fn: mean)

    testing.diff(got, want)
}

testcase mean_with_nulls {
    want =
        array.from(
            rows: [
                {_time: 2019-11-25T00:00:00Z, t0: "a-0", _value: 3.75},
                {_time: 2019-11-25T00:00:00Z, t0: "a-1", _value: 3.75},
                {_time: 2019-11-25T00:00:00Z, t0: "a-2", _value: 3.75},
            ],
        )
            |> map(fn: (r) => ({r with _measurement: "m0", _field: "f0"}))
            |> group(columns: ["_measurement", "_field", "t0"])
    got =
        do_test(every: 10s, fn: sum)
            |> aggregateWindow(every: 1m, fn: mean, timeSrc: "_start")
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want)
}

testcase mean_null_windows {
    want =
        csv.from(
            csv:
                "#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2019-11-25T00:00:00Z,m0,f0,6.0
,,0,2019-11-25T00:00:10Z,m0,f0,2.0
,,0,2019-11-25T00:00:20Z,m0,f0,
,,0,2019-11-25T00:00:30Z,m0,f0,3.0
,,0,2019-11-25T00:00:40Z,m0,f0,4.0
,,0,2019-11-25T00:00:50Z,m0,f0,
",
        )
    got =
        do_test(every: 10s, fn: sum)
            |> group(columns: ["_measurement", "_field"])
            |> aggregateWindow(every: 10s, fn: mean, timeSrc: "_start")
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want)
}

testcase min_empty_windows {
    want =
        csv.from(
            csv:
                "#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,t0,_value
,,0,2019-11-25T00:00:00Z,m0,f0,a-0,1.0
,,0,2019-11-25T00:00:10Z,m0,f0,a-0,2.0
,,0,2019-11-25T00:00:20Z,m0,f0,a-0,
,,0,2019-11-25T00:00:30Z,m0,f0,a-0,3.0
,,0,2019-11-25T00:00:40Z,m0,f0,a-0,4.0
,,0,2019-11-25T00:00:50Z,m0,f0,a-0,
,,1,2019-11-25T00:00:00Z,m0,f0,a-1,1.0
,,1,2019-11-25T00:00:10Z,m0,f0,a-1,2.0
,,1,2019-11-25T00:00:20Z,m0,f0,a-1,
,,1,2019-11-25T00:00:30Z,m0,f0,a-1,3.0
,,1,2019-11-25T00:00:40Z,m0,f0,a-1,4.0
,,1,2019-11-25T00:00:50Z,m0,f0,a-1,
,,2,2019-11-25T00:00:00Z,m0,f0,a-2,1.0
,,2,2019-11-25T00:00:10Z,m0,f0,a-2,2.0
,,2,2019-11-25T00:00:20Z,m0,f0,a-2,
,,2,2019-11-25T00:00:30Z,m0,f0,a-2,3.0
,,2,2019-11-25T00:00:40Z,m0,f0,a-2,4.0
,,2,2019-11-25T00:00:50Z,m0,f0,a-2,
",
        )
    got = do_test(every: 10s, fn: min)

    testing.diff(got, want)
}

testcase min_with_nulls {
    want =
        array.from(
            rows: [
                {_time: 2019-11-25T00:00:00Z, t0: "a-0", _value: 2.0},
                {_time: 2019-11-25T00:00:00Z, t0: "a-1", _value: 2.0},
                {_time: 2019-11-25T00:00:00Z, t0: "a-2", _value: 2.0},
            ],
        )
            |> map(fn: (r) => ({r with _measurement: "m0", _field: "f0"}))
            |> group(columns: ["_measurement", "_field", "t0"])
    got =
        do_test(every: 10s, fn: sum)
            |> aggregateWindow(every: 1m, fn: min, timeSrc: "_start")
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want)
}

testcase min_null_windows {
    want =
        csv.from(
            csv:
                "#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,false,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,t0,_value
,,0,2019-11-25T00:00:00Z,m0,f0,a-0,6.0
,,0,2019-11-25T00:00:10Z,m0,f0,a-0,2.0
,,0,2019-11-25T00:00:20Z,m0,f0,,
,,0,2019-11-25T00:00:30Z,m0,f0,a-0,3.0
,,0,2019-11-25T00:00:40Z,m0,f0,a-0,4.0
,,0,2019-11-25T00:00:50Z,m0,f0,,
",
        )
    got =
        do_test(every: 10s, fn: sum)
            |> group(columns: ["_measurement", "_field"])
            |> aggregateWindow(every: 10s, fn: min, timeSrc: "_start")
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want)
}

testcase max_empty_windows {
    want =
        csv.from(
            csv:
                "#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,t0,_value
,,0,2019-11-25T00:00:00Z,m0,f0,a-0,5.0
,,0,2019-11-25T00:00:10Z,m0,f0,a-0,2.0
,,0,2019-11-25T00:00:20Z,m0,f0,a-0,
,,0,2019-11-25T00:00:30Z,m0,f0,a-0,3.0
,,0,2019-11-25T00:00:40Z,m0,f0,a-0,4.0
,,0,2019-11-25T00:00:50Z,m0,f0,a-0,
,,1,2019-11-25T00:00:00Z,m0,f0,a-1,5.0
,,1,2019-11-25T00:00:10Z,m0,f0,a-1,2.0
,,1,2019-11-25T00:00:20Z,m0,f0,a-1,
,,1,2019-11-25T00:00:30Z,m0,f0,a-1,3.0
,,1,2019-11-25T00:00:40Z,m0,f0,a-1,4.0
,,1,2019-11-25T00:00:50Z,m0,f0,a-1,
,,2,2019-11-25T00:00:00Z,m0,f0,a-2,5.0
,,2,2019-11-25T00:00:10Z,m0,f0,a-2,2.0
,,2,2019-11-25T00:00:20Z,m0,f0,a-2,
,,2,2019-11-25T00:00:30Z,m0,f0,a-2,3.0
,,2,2019-11-25T00:00:40Z,m0,f0,a-2,4.0
,,2,2019-11-25T00:00:50Z,m0,f0,a-2,
",
        )
    got = do_test(every: 10s, fn: max)

    testing.diff(got, want)
}

testcase max_with_nulls {
    want =
        array.from(
            rows: [
                {_time: 2019-11-25T00:00:00Z, t0: "a-0", _value: 6.0},
                {_time: 2019-11-25T00:00:00Z, t0: "a-1", _value: 6.0},
                {_time: 2019-11-25T00:00:00Z, t0: "a-2", _value: 6.0},
            ],
        )
            |> map(fn: (r) => ({r with _measurement: "m0", _field: "f0"}))
            |> group(columns: ["_measurement", "_field", "t0"])
    got =
        do_test(every: 10s, fn: sum)
            |> aggregateWindow(every: 1m, fn: max, timeSrc: "_start")
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want)
}

testcase max_null_windows {
    want =
        csv.from(
            csv:
                "#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,false,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,t0,_value
,,0,2019-11-25T00:00:00Z,m0,f0,a-0,6.0
,,0,2019-11-25T00:00:10Z,m0,f0,a-0,2.0
,,0,2019-11-25T00:00:20Z,m0,f0,,
,,0,2019-11-25T00:00:30Z,m0,f0,a-0,3.0
,,0,2019-11-25T00:00:40Z,m0,f0,a-0,4.0
,,0,2019-11-25T00:00:50Z,m0,f0,,
",
        )
    got =
        do_test(every: 10s, fn: sum)
            |> group(columns: ["_measurement", "_field"])
            |> aggregateWindow(every: 10s, fn: max, timeSrc: "_start")
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want)
}

testcase aggregate_window_create_empty_predecessor_multi_successor {
    expect.planner(rules: ["AggregateWindowCreateEmptyRule": 1])

    data =
        array.from(rows: sampleData)
            |> range(start: 2019-11-25T00:00:00Z, stop: 2019-11-25T00:01:00Z)
            |> group(columns: ["t0", "_start", "_stop"])

    // This additional successor to the input to aggregateWindow caused trouble
    // for the planner, so we are just verifying that the rule is applied
    // and the query succeeds.
    data
        |> debug.sink()

    got =
        data
            |> aggregateWindow(fn: count, every: 1m)
            |> drop(columns: ["_start", "_stop"])
    want =
        array.from(
            rows: [
                {_time: 2019-11-25T00:01:00Z, t0: "a-0", _value: 5},
                {_time: 2019-11-25T00:01:00Z, t0: "a-1", _value: 5},
                {_time: 2019-11-25T00:01:00Z, t0: "a-2", _value: 5},
            ],
        )
            |> group(columns: ["t0", "_start", "_stop"])

    testing.diff(want, got)
}

testcase aggregate_window_predecessor_multi_successor {
    expect.planner(rules: ["AggregateWindowRule": 1])

    data =
        array.from(rows: sampleData)
            |> range(start: 2019-11-25T00:00:00Z, stop: 2019-11-25T00:01:00Z)
            |> group(columns: ["t0", "_start", "_stop"])

    // This additional successor to the input to aggregateWindow caused trouble
    // for the planner, so we are just verifying that the rule is applied
    // and the query succeeds.
    data
        |> debug.sink()

    got =
        data
            |> aggregateWindow(fn: count, every: 1m, createEmpty: false)
            |> drop(columns: ["_start", "_stop"])
    want =
        array.from(
            rows: [
                {_time: 2019-11-25T00:01:00Z, t0: "a-0", _value: 5},
                {_time: 2019-11-25T00:01:00Z, t0: "a-1", _value: 5},
                {_time: 2019-11-25T00:01:00Z, t0: "a-2", _value: 5},
            ],
        )
            |> group(columns: ["t0", "_start", "_stop"])

    testing.diff(want, got)
}
