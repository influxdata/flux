// Package aggregate provides functions to simplify common aggregate operations.
//
// ## Metadata
// introduced: 0.61.0
//
package aggregate


import "experimental"

// rate calculates the average rate of increase per window of time for each
// input table.
//
// `aggregate.rate()` requires that input data have `_start` and `_stop` columns
// to calculate windows of time to operate on.
// Use `range()` to assign `_start` and `_stop` values.
//
// This function is designed to replicate the
// [Prometheus `rate()` function](https://prometheus.io/docs/prometheus/latest/querying/functions/#rate)
// and should only be used with [counters](https://docs.influxdata.com/flux/v0.x/prometheus/metric-types/counter/).
//
// ## Parameters
//
// - every: Duration of time windows.
// - groupColumns: List of columns to group by. Default is `[]`.
// - unit: Time duration to use when calculating the rate. Default is `1s`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate the average rate of change in data
// ```
// import "experimental/aggregate"
// import "sampledata"
//
// data = sampledata.int()
//     |> range(start: sampledata.start, stop: sampledata.stop)
//
// < data
// >     |> aggregate.rate(every: 30s, unit: 1s, groupColumns: ["tag"])
// ```
//
// ## Metadata
// tags: transformations,aggregates
//
rate = (tables=<-, every, groupColumns=[], unit=1s) =>
    tables
        |> derivative(nonNegative: true, unit: unit)
        |> aggregateWindow(
            every: every,
            fn: (tables=<-, column) =>
                tables
                    |> mean(column: column)
                    |> group(columns: groupColumns)
                    |> experimental.group(columns: ["_start", "_stop"], mode: "extend")
                    |> sum(),
        )
