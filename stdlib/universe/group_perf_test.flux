package universe_test


import "testing"
import "internal/gen"
import "array"

option now = () => 2030-01-01T00:00:00Z

numSeries = 1000
pointsPerSeries = 10

input = gen.tables(
    n: pointsPerSeries,
    tags: [
        {cardinality: 1, name: "_measurement"},
        {cardinality: 1, name: "_field"},
        {cardinality: numSeries, name: "orgID"},
    ],
)
    |> range(start: -10y)
    |> map(fn: (r) => ({r with _measurement: "query_log", _field: "totalDuration"}))

want = array.from(rows: [{_measurement: "query_log", _value: numSeries * pointsPerSeries}])
    |> group(columns: ["_measurement"])

testcase group_perf_good {
    result = input |> testing.load()
        |> range(start: -10y)
        |> filter(fn: (r) => r._measurement == "query_log" or r._measurement == "influxql_query_log")
        |> filter(fn: (r) => r._field == "totalDuration")
        |> group(columns: ["_measurement", "orgID"])
        |> count()
        |> group(columns: ["_measurement"])
        |> sum()

    testing.diff(got: result, want: want) |> yield()
}

testcase group_perf_bad {
    result = input |> testing.load()
        |> range(start: -10y)
        |> filter(fn: (r) => r._measurement == "query_log" or r._measurement == "influxql_query_log")
        |> filter(fn: (r) => r._field == "totalDuration")
        |> count()
        |> group(columns: ["_measurement"])
        |> sum()

    testing.diff(got: result, want: want) |> yield()
}
