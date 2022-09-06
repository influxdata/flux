package universe_test


import "array"
import "internal/debug"
import "testing"

testcase group_nulls {
    // FIXME: in C2/OSS since the group key incorrectly empty for the case where `t0` is null.
    // Remove skip tag after https://github.com/influxdata/flux/issues/5178
    option testing.tags = ["skip"]

    input =
        array.from(
            rows: [
                {
                    _time: 2022-09-07T00:00:01Z,
                    _measurement: "m0",
                    _field: "f0",
                    _value: 1,
                    t0: "a",
                },
                {
                    _time: 2022-09-07T00:00:02Z,
                    _measurement: "m0",
                    _field: "f0",
                    _value: 2,
                    t0: "a",
                },
                {
                    _time: 2022-09-07T00:00:03Z,
                    _measurement: "m0",
                    _field: "f0",
                    _value: 3,
                    t0: "a",
                },
                {
                    _time: 2022-09-07T00:00:01Z,
                    _measurement: "m0",
                    _field: "f0",
                    _value: 1,
                    t0: debug.null(type: "string"),
                },
                {
                    _time: 2022-09-07T00:00:02Z,
                    _measurement: "m0",
                    _field: "f0",
                    _value: 2,
                    t0: debug.null(type: "string"),
                },
                {
                    _time: 2022-09-07T00:00:03Z,
                    _measurement: "m0",
                    _field: "f0",
                    _value: 3,
                    t0: debug.null(type: "string"),
                },
            ],
        )

    want = input |> group(columns: ["t0"])

    got =
        input
            |> group(columns: ["_measurement", "_field", "_time", "t0"])
            |> testing.load()
            |> range(start: 2022-09-07T00:00:00Z)
            |> group(columns: ["t0"])
            |> drop(columns: ["_start", "_stop"])

    testing.diff(got, want)
}
