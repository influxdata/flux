// Package promql provides an internal API for implementing PromQL via Flux.
//
// **Important**: This package is not meant for external use.
//
// ## Metadata
// introduced: 0.47.0
package promql


import "math"
import "universe"
import "experimental"

// changes implements functionality equivalent to
// [PromQL's `changes()` function](https://prometheus.io/docs/prometheus/latest/querying/functions/#changes).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the number of times that values in a series change
// ```
// import "internal/promql"
// import "sampledata"
//
// < sampledata.float()
// >     |> promql.changes()
// ```
//
builtin changes : (<-tables: stream[{A with _value: float}]) => stream[{B with _value: float}]

// promqlDayOfMonth implements functionality equivalent to
// [PromQL's `day_of_month()` function](https://prometheus.io/docs/prometheus/latest/querying/functions/#day_of_month).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - timestamp: Time as a floating point value.
builtin promqlDayOfMonth : (timestamp: float) => float

// promqlDayOfWeek implements functionality equivalent to
// [PromQL's `day_of_week()` function](https://prometheus.io/docs/prometheus/latest/querying/functions/#day_of_week).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - timestamp: Time as a floating point value.
builtin promqlDayOfWeek : (timestamp: float) => float

// promqlDaysInMonth implements functionality equivalent to
// [PromQL's `days_in_month()` function](https://prometheus.io/docs/prometheus/latest/querying/functions/#days_in_month).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - timestamp: Time as a floating point value.
builtin promqlDaysInMonth : (timestamp: float) => float

// emptyTable returns an empty table, which is used as a helper function to implement
// PromQL's [`time()`](https://prometheus.io/docs/prometheus/latest/querying/functions/#time) and
// [`vector()`](https://prometheus.io/docs/prometheus/latest/querying/functions/#vector) functions.
//
// **Important**: The `internal/promql` package is not meant for external use.
//
builtin emptyTable : () => stream[{_start: time, _stop: time, _time: time, _value: float}]

// extrapolatedRate is a helper function that calculates extrapolated rates over
// counters and is used to implement PromQL's
// [`rate()`](https://prometheus.io/docs/prometheus/latest/querying/functions/#rate),
// [`delta()`](https://prometheus.io/docs/prometheus/latest/querying/functions/#increase),
// and [`increase()`](https://prometheus.io/docs/prometheus/latest/querying/functions/#delta) functions.
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
// - isCounter: Data represents a counter.
// - isRate: Data represents a rate.
builtin extrapolatedRate : (
        <-tables: stream[{A with _start: time, _stop: time, _time: time, _value: float}],
        ?isCounter: bool,
        ?isRate: bool,
    ) => stream[{B with _value: float}]

// holtWinters implements functionality equivalent to
// [PromQL's `holt_winters()` function](https://prometheus.io/docs/prometheus/latest/querying/functions/#holt_winters).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
// - smoothingFactor: Exponential smoothing factor.
// - trendFactor: Trend factor.
builtin holtWinters : (
        <-tables: stream[{A with _time: time, _value: float}],
        ?smoothingFactor: float,
        ?trendFactor: float,
    ) => stream[{B with _value: float}]

// promqlHour implements functionality equivalent to
// [PromQL's `hour()` function](https://prometheus.io/docs/prometheus/latest/querying/functions/#hour).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - timestamp: Time as a floating point value.
builtin promqlHour : (timestamp: float) => float

// instantRate is a helper function that calculates instant rates over
// counters and is used to implement PromQL's
// [`irate()`](https://prometheus.io/docs/prometheus/latest/querying/functions/#irate) and
// [`idelta()`](https://prometheus.io/docs/prometheus/latest/querying/functions/#idelta) functions.
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - tables: Input data. Defaults is piped-forward data (`<-`).
// - isRate: Data represents a rate.
builtin instantRate : (
        <-tables: stream[{A with _time: time, _value: float}],
        ?isRate: bool,
    ) => stream[{B with _value: float}]

// labelReplace implements functionality equivalent to
// [PromQL's `label_replace()` function](https://prometheus.io/docs/prometheus/latest/querying/functions/#label_replace).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
// - source: Input label.
// - destination: Output label.
// - regex: Pattern as a regex string.
// - replacement: Replacement value.
builtin labelReplace : (
        <-tables: stream[{A with _value: float}],
        source: string,
        destination: string,
        regex: string,
        replacement: string,
    ) => stream[{B with _value: float}]

// linearRegression implements linear regression functionality required to implement
// PromQL's [`deriv()`](https://prometheus.io/docs/prometheus/latest/querying/functions/#deriv)
// and [`predict_linear()`](https://prometheus.io/docs/prometheus/latest/querying/functions/#predict_linear) functions.
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
// - predict: Output should contain a prediction.
// - fromNow: Time as a floating point value.
builtin linearRegression : (
        <-tables: stream[{A with _time: time, _stop: time, _value: float}],
        ?predict: bool,
        ?fromNow: float,
    ) => stream[{B with _value: float}]

// promqlMinute implements functionality equivalent to
// [PromQL's `minute()` function]( https://prometheus.io/docs/prometheus/latest/querying/functions/#minute).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - timestamp: Time as a floating point value.
builtin promqlMinute : (timestamp: float) => float

// promqlMonth implements functionality equivalent to
// [PromQL's `month()` function](https://prometheus.io/docs/prometheus/latest/querying/functions/#month).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - timestamp: Time as a floating point value.
builtin promqlMonth : (timestamp: float) => float

// promHistogramQuantile implements functionality equivalent to
// [PromQL's `histogram_quantile()` function](https://prometheus.io/docs/prometheus/latest/querying/functions/#histogram_quantile).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
// - quantile: Quantile to compute (`[0.0 - 1.0]`).
// - countColumn: Count column name.
// - upperBoundColumn: Upper bound column name.
// - valueColumn: Output value column name.
builtin promHistogramQuantile : (
        <-tables: stream[A],
        ?quantile: float,
        ?countColumn: string,
        ?upperBoundColumn: string,
        ?valueColumn: string,
    ) => stream[B]
    where
    A: Record,
    B: Record

// resets implements functionality equivalent to
// [PromQL's `resets()` function](https://prometheus.io/docs/prometheus/latest/querying/functions/#resets).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - tables: Input data. Defaults is piped-forward data (`<-`).
builtin resets : (<-tables: stream[{A with _value: float}]) => stream[{B with _value: float}]

// timestamp implements functionality equivalent to
// [PromQL's `timestamp()` function](https://prometheus.io/docs/prometheus/latest/querying/functions/#timestamp).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - tables: Input data. Defaults is piped-forward data (`<-`).
//
// ## Examples
//
// ### Convert timestamps into seconds since the Unix epoch
//
// ```
// import "internal/promql"
// import "sampledata"
//
// < sampledata.float()
// >     |> promql.timestamp()
// ```
builtin timestamp : (<-tables: stream[{A with _value: float}]) => stream[{A with _value: float}]

// promqlYear implements functionality equivalent to
// [PromQL's `year()` function](https://prometheus.io/docs/prometheus/latest/querying/functions/#year).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - timestamp: Time as a floating point value.
builtin promqlYear : (timestamp: float) => float

// quantile accounts checks for quantile values that are out of range, above 1.0 or
// below 0.0, by either returning positive infinity or negative infinity in the `_value`
// column respectively. `q` must be a float.
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
// - q: Quantile to compute (`[0.0 - 1.0]`).
// - method: Quantile method to use.
quantile = (q, tables=<-, method="exact_mean") =>
    // value is in normal range. We can use the normal quantile function
    if q <= 1.0 and q >= 0.0 then
        tables
            |> universe.quantile(q: q, method: method)
    else if q < 0.0 then
        tables
            |> reduce(identity: {_value: math.mInf(sign: -1)}, fn: (r, accumulator) => accumulator)
    else
        tables
            |> reduce(identity: {_value: math.mInf(sign: 1)}, fn: (r, accumulator) => accumulator)

// join joins two streams of tables on the **group key and `_time` column**.
// See [`experimental.join`](https://docs.influxdata.com/flux/v0.x/stdlib/experimental/join/).
//
// **Important**: The `internal/promql` package is not meant for external use.
//
// ## Parameters
// - left: First of two streams of tables to join.
// - right: Second of two streams of tables to join.
// - fn: Function with left and right arguments that maps a new output record
//   using values from the `left` and `right` input records.
//   The return value must be a record.
join = experimental.join
