package universe_test


import "array"
import "internal/debug"
import "join"
import "testing"
import "testing/expect"

rows =
    array.from(
        rows: [
            {_time: 2010-11-11T00:00:00Z, i: 10, desc: "cat"},
            {_time: 2010-11-11T00:00:10Z, i: 9, desc: "dog"},
            {_time: 2010-11-11T00:00:20Z, i: 8, desc: "tiger"},
            {_time: 2010-11-11T00:00:30Z, i: 7, desc: "bear"},
            {_time: 2010-11-11T00:00:40Z, i: 6, desc: "lion"},
            {_time: 2010-11-11T00:00:50Z, i: 5, desc: "zebra"},
            {_time: 2010-11-11T00:01:00Z, i: 4, desc: "koala"},
            {_time: 2010-11-11T00:01:10Z, i: 3, desc: "giraffe"},
            {_time: 2010-11-11T00:01:20Z, i: 2, desc: "canary"},
            {_time: 2010-11-11T00:01:30Z, i: 1, desc: "gazelle"},
        ],
    )

testcase remove_sort {
    expect.planner(rules: ["universe/RemoveRedundantSort": 1])

    input =
        rows
            |> sort(columns: ["i"])
    got =
        input
            // workaround for https://github.com/influxdata/flux/issues/4699
            |> debug.pass()
            |> sort(columns: ["i"])
    want = input

    testing.diff(want, got) |> yield()
}

testcase remove_sort_more_columns {
    expect.planner(rules: ["universe/RemoveRedundantSort": 1])

    input =
        rows
            |> sort(columns: ["i", "desc"])
    got =
        input
            // workaround for https://github.com/influxdata/flux/issues/4699
            |> debug.pass()
            |> sort(columns: ["i"])
    want = input

    testing.diff(want, got) |> yield()
}

testcase remove_sort_fewer_columns {
    // sort should not be removed here because input rows
    // are not sorted by "desc"
    expect.planner(rules: ["universe/RemoveRedundantSort": 0])

    input =
        rows
            |> sort(columns: ["i"])
    got =
        input
            // workaround for https://github.com/influxdata/flux/issues/4699
            |> debug.pass()
            |> sort(columns: ["i", "desc"])
    want = input

    testing.diff(want, got) |> yield()
}

testcase remove_sort_aggregate {
    expect.planner(rules: ["universe/RemoveRedundantSort": 1])

    input =
        rows
            |> sort(columns: ["i"])
    got =
        input
            |> sum(column: "i")
            |> sort(columns: ["i"])
    want = input |> sum(column: "i")

    testing.diff(want, got) |> yield()
}

testcase remove_sort_selector {
    expect.planner(rules: ["universe/RemoveRedundantSort": 1])

    input =
        rows
            |> sort(columns: ["i"])
    got =
        input
            |> min(column: "i")
            |> sort(columns: ["i"])
    want = input |> min(column: "i")

    testing.diff(want, got) |> yield()
}

testcase remove_sort_filter_range {
    expect.planner(rules: ["universe/RemoveRedundantSort": 1])

    input =
        rows
            |> sort(columns: ["i"])
    got =
        input
            |> range(start: -100y)
            |> filter(fn: (r) => r.i < 100)
            |> sort(columns: ["i"])
            |> drop(columns: ["_start", "_stop"])
    want =
        input
            |> filter(fn: (r) => r.i < 100)

    testing.diff(want, got) |> yield()
}

testcase remove_sort_aggregate_window {
    expect.planner(rules: ["universe/RemoveRedundantSort": 1])

    input =
        rows
            |> range(start: 2010-11-11T00:00:00Z, stop: 2010-11-11T00:05:00Z)
    got =
        input
            // workaround for https://github.com/influxdata/flux/issues/4699
            |> debug.pass()
            |> aggregateWindow(every: 30s, fn: sum, column: "i")
            |> sort(columns: ["_time"])
    want =
        input
            |> debug.pass()
            // workaround for https://github.com/influxdata/flux/issues/4699
            |> aggregateWindow(every: 30s, fn: sum, column: "i")

    testing.diff(want, got) |> yield()
}

testcase remove_sort_join {
    expect.planner(rules: ["universe/RemoveRedundantSort": 2])

    inputLeft = rows
    inputRight = rows

    sortTime = (tables=<-) => tables |> sort(columns: ["_time"])

    // When join is planned, it will get sort nodes generated for each input
    //   join(left, right)
    // becomes
    //   join(sort(left), sort(right))
    //
    // Since both inputs to the join are already sorted, the planner should remove both
    // of the generated sort nodes, hence the rule will fire twice.
    got =
        join.time(
            left: inputLeft |> sortTime(),
            right: inputRight |> sortTime(),
            as: (l, r) => {
                return {_time: l._time, ldesc: l.desc, rdesc: r.desc, i: l.i}
            },
        )

    // Neither side is sorted, so the generated sort nodes will remain,
    // hence the rule will *not* fire four times.
    want =
        join.time(
            left: inputLeft |> debug.pass(),
            right: inputRight |> debug.pass(),
            as: (l, r) => {
                return {_time: l._time, ldesc: l.desc, rdesc: r.desc, i: l.i}
            },
        )

    testing.diff(want, got) |> yield()
}
