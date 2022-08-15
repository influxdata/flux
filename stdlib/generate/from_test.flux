package generate_test


import "array"
import "testing"
import "generate"

testcase unionTwoStreamsWithEmptyGroupKeys {
    // Adaptation of the example at:
    // <https://docs.influxdata.com/flux/v0.x/stdlib/universe/union/#union-two-streams-of-tables-with-empty-group-keys>
    t1 =
        generate.from(
            count: 4,
            fn: (n) => n + 1,
            start: 2021-01-01T00:00:00Z,
            stop: 2021-01-05T00:00:00Z,
        )
            |> set(key: "tag", value: "foo")
            |> group()

    t2 =
        generate.from(
            count: 4,
            fn: (n) => n * (-1),
            start: 2021-01-01T00:00:00Z,
            stop: 2021-01-05T00:00:00Z,
        )
            |> set(key: "tag", value: "bar")
            |> group()

    got =
        union(tables: [t1, t2])
            |> drop(columns: ["_start", "_stop"])
            |> sort(columns: ["tag", "_time"])

    want =
        array.from(
            rows: [
                {_time: 2021-01-01T00:00:00Z, tag: "bar", _value: 0},
                {_time: 2021-01-02T00:00:00Z, tag: "bar", _value: -1},
                {_time: 2021-01-03T00:00:00Z, tag: "bar", _value: -2},
                {_time: 2021-01-04T00:00:00Z, tag: "bar", _value: -3},
                {_time: 2021-01-01T00:00:00Z, tag: "foo", _value: 1},
                {_time: 2021-01-02T00:00:00Z, tag: "foo", _value: 2},
                {_time: 2021-01-03T00:00:00Z, tag: "foo", _value: 3},
                {_time: 2021-01-04T00:00:00Z, tag: "foo", _value: 4},
            ],
        )

    testing.diff(want: want, got: got) |> yield()
}
