// Package universe provides options and primitive functions that are
// loaded into the Flux runtime by default and do not require an
// import statement.
//
// introduced: 0.14.0
//
package universe


import "system"
import "date"
import "math"
import "strings"
import "regexp"
import "experimental/table"

// now is a function option that, by default, returns the current system time.
//
// #### now() vs system.time()
// `now()` returns the current system time (UTC). `now()` is cached at runtime,
// so all executions of `now()` in a Flux script return the same time value.
// `system.time()` returns the system time (UTC) at which `system.time()` is executed.
// Each instance of `system.time()` in a Flux script returns a unique value.
//
// ## Examples
//
// ### Use the current UTC time as a query boundary
// ```no_run
// data
//     |> range(start: -10h, stop: now)
// ```
//
// ### Define a custom now time
// ```no_run
// option now = () => 2022-01-01T00:00:00Z
// ```
//
// introduced: 0.7.0
// tags: date/time
//
option now = system.time

// chandeMomentumOscillator applies the technical momentum indicator developed
// by Tushar Chande to input data.
//
// The Chande Momentum Oscillator (CMO) indicator calculates the difference
// between the sum of all recent data points with values greater than the median
// value of the data set and the sum of all recent data points with values lower
// than the median value of the data set, then divides the result by the sum of
// all data movement over a given time period.
// It then multiplies the result by 100 and returns a value between -100 and +100.
//
// #### Output tables
// For each input table with x rows, `chandeMomentumOscillator()` outputs a
// table with `x - n` rows.
//
// ## Parameters
// - n: Period or number of points to use in the calculation.
// - columns: List of columns to operate on. Default is `["_value"]`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Apply the Chande Momentum Oscillator to input data
// ```
// import "sampledata"
//
// sampledata.int()
//     |> chandeMomentumOscillator(n: 2)
// ```
//
// introduced: 0.39.0
// tags: transformations
//
builtin chandeMomentumOscillator : (<-tables: [A], n: int, ?columns: [string]) => [B] where A: Record, B: Record

// columns returns the column labels in each input table.
//
// For each input table, `columns` outputs a table with the same group key
// columns and a new column containing the column labels in the input table.
// Each row in an output table contains the group key value and the label of one
//  column of the input table.
// Each output table has the same number of rows as the number of columns of the
// input table.
//
// ## Parameters
// - column: Name of the output column to store column labels in.
//   Default is "_value".
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### List all columns per input table
// ```
// import "sampledata"
//
// sampledata.string()
//     |> columns(column: "labels")
// ```
//
// introduced: 0.14.0
// tags: transformations
//
builtin columns : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record

// count returns the number of records in a column.
//
// The function counts both null and non-null records.
//
// #### Empty tables
// `count()` returns `0` for empty tables.
// To keep empty tables in your data, set the following parameters for the
// following functions:
//
// | Function            | Parameter           |
// | :------------------ | :------------------ |
// | `filter()`          | `onEmpty: "keep"`   |
// | `window()`          | `createEmpty: true` |
// | `aggregateWindow()` | `createEmpty: true` |
//
// ## Parameters
// - column: Column to count values in and store the total count.
// - tables: Input data. Default is piped-wforward data (`<-`).
//
// ## Examples
//
// ### Count the number of rows in each input table
// ```
// import "sampledata"
//
// sampledata.string()
//     |> count()
// ```
//
// introduced: 0.7.0
// tags: transformations,aggregates
//
builtin count : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record

// covariance computes the covariance between two columns.
//
// ## Parameters
// - columns: List of two columns to operate on.
// - pearsonr: Normalize results to the Pearson R coefficient. Default is `false`.
// - valueDst: Column to store the result in. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate the covariance between two columns
// ```
// # import "generate"
// #
// # data =
// #     generate.from(count: 5, fn: (n) => n * n, start: 2021-01-01T00:00:00Z, stop: 2021-01-01T00:01:00Z)
// #         |> toFloat()
// #         |> map(fn: (r) => ({_time: r._time, x: r._value, y: r._value * r._value / 2.0}))
// #
// < data
// >     |> covariance(columns: ["x", "y"])
// ```
//
// introduced: 0.7.0
// tags: transformations,aggregates
//
builtin covariance : (<-tables: [A], ?pearsonr: bool, ?valueDst: string, columns: [string]) => [B]
    where
    A: Record,
    B: Record

// cumulativeSum  computes a running sum for non-null records in a table.
//
// The output table schema will be the same as the input table.
//
// ## Parameters
// - columns: List of columns to operate on. Default is `["_value"]`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the running total of values in each table
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> cumulativeSum()
// ```
//
// introduced: 0.7.0
// tags: transformations
//
builtin cumulativeSum : (<-tables: [A], ?columns: [string]) => [B] where A: Record, B: Record

// derivative computes the rate of change per unit of time between subsequent
// non-null records.
//
// The function assumes rows are ordered by the `_time`.
//
// #### Output tables
// The output table schema will be the same as the input table.
// For each input table with `n` rows, `derivative()` outputs a table with
// `n - 1` rows.
//
// ## Parameters
// - unit: Time duration used to calculate the derivative. Default is `1s`.
// - nonNegative: Disallow negative derivative values. Default is `true`.
//
//   When `true`, if a value is less than the previous value, the function
//   assumes the previous value should have been a zero.
//
// - columns: List of columns to operate on. Default is `["_value"]`.
// - timeColumn: Column containing time values to use in the calculation.
//   Default is `_time`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate the non-negative rate of change per second
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> derivative(nonNegative: true)
// ```
//
// ### Calculate the rate of change per second with null values
// ```
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> derivative()
// ```
//
// introduced: 0.7.0
// tags: transformations
//
builtin derivative : (
        <-tables: [A],
        ?unit: duration,
        ?nonNegative: bool,
        ?columns: [string],
        ?timeColumn: string,
    ) => [B]
    where
    A: Record,
    B: Record

// die stops the Flux script execution and returns an error message.
//
// ## Parameters
// - msg: Error message to return.
//
// ## Examples
//
// ### Force a script to exit with an error message
// ```no_run
// die(msg: "This is an error message")
// ```
//
// introduced: 0.82.0
//
builtin die : (msg: string) => A

// difference returns the difference between subsequent values.
//
// ### Subtraction rules for numeric types
// - The difference between two non-null values is their algebraic difference;
//   or `null`, if the result is negative and `nonNegative: true`;
// - `null` minus some value is always `null`;
// - Some value `v` minus `null` is `v` minus the last non-null value seen
//   before `v`; or `null` if `v` is the first non-null value seen.
// - If `nonNegative` and `initialZero` are set to `true`, `difference()`
//   returns the difference between `0` and the subsequent value.
//   If the subsequent value is less than zero, `difference()` returns `null`.
//
// ### Output tables
// For each input table with `n` rows, `difference()` outputs a table with
// `n - 1` rows.
//
// ## Parameters
// - nonNegative: Disallow negative differences. Default is `false`.
//
//   When `true`, if a value is less than the previous value, the function
//   assumes the previous value should have been a zero.
//
// - columns: List of columns to operate on. Default is `["_value"]`.
// - keepFirst: Keep the first row in each input table. Default is `false`.
//
//   If `true`, the difference of the first row of each output table is null.
//
// - initialZero: Use zero (0) as the initial value in the difference calculation
//   when the subsequent value is less than the previous value and `nonNegative` is
//   `true`. Default is `false`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate the difference between subsequent values
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> difference()
// ```
//
// ### Calculate the non-negative difference between subsequent values
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> difference(nonNegative: true)
// ```
//
// ### Calculate the difference between subsequent values with null values
// ```
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> difference()
// ```
//
// ### Keep the first value when calculating the difference between values
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> difference(keepFirst: true)
// ```
//
// introduced: 0.7.1
// tags: transformations
//
builtin difference : (
        <-tables: [T],
        ?nonNegative: bool,
        ?columns: [string],
        ?keepFirst: bool,
        ?initialZero: bool,
    ) => [R]
    where
    T: Record,
    R: Record

// distinct returns all unique values in a specified column.
//
// The `_value` of each output record is set to a distinct value in the specified column.
// `null` is considered its own distinct value if present.
//
// ## Parameters
// - column: Column to return unique values from. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return distinct values from the _value column
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> distinct()
// ```
//
// ### Return distinct values from a non-default column
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> distinct(column: "tag")
// ```
//
// ### Return distinct values from data with null values
// ```
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> distinct()
// ```
//
// introduced: 0.7.0
// tags: transformations, selectors
//
builtin distinct : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record

// drop removes specified columns from a table.
//
// Columns are specified either through a list or a predicate function.
// When a dropped column is part of the group key, it is removed from the key.
// If a specified column is not present in a table, the function returns an error.
//
// ## Parameters
// - columns: List of columns to remove from input tables. Mutually exclusive with `fn`.
// - fn: Predicate function with a `column` parameter that returns a boolean
//   value indicating whether or not the column should be removed from input tables.
//   Mutually exclusive with `columns`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Drop a list of columns
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> drop(columns: ["_time", "tag"])
// ```
//
// ### Drop columns matching a predicate
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> drop(fn: (column) => column =~ /^t/)
// ```
//
// introduced: 0.7.0
// tags: transformations
//
builtin drop : (<-tables: [A], ?fn: (column: string) => bool, ?columns: [string]) => [B] where A: Record, B: Record

// duplicate duplicates a specified column in a table.
//
// If the specified column is part of the group key, it will be duplicated, but
// the duplicate column will not be part of the output’s group key.
//
// ## Parameters
// - column: Column to duplicate.
// - as: Name to assign to the duplicate column.
//
//   If the `as` column already exists, it will be overwritten by the duplicated column.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Duplicate a column
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> duplicate(column: "tag", as: "tag_dup")
// ```
//
// introduced: 0.7.0
// tags: transformations
//
builtin duplicate : (<-tables: [A], column: string, as: string) => [B] where A: Record, B: Record

// elapsed returns the time between subsequent records.
//
// For each input table, `elapsed()` returns the same table without the first row
// (because there is no previous time to derive the elapsed time from) and an
// additional column containing the elapsed time.
//
// ## Parameters
// - unit: Unit of time used in the calculation. Default is `1s`.
// - timeColumn: Column to use to compute the elapsed time. Default is `_time`.
// - columnName: Column to store elapsed times in. Default is `elapsed`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate the time between points in seconds
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> elapsed(unit: 1s)
// ```
//
// introduced: 0.36.0
// tags: transformations
//
builtin elapsed : (<-tables: [A], ?unit: duration, ?timeColumn: string, ?columnName: string) => [B]
    where
    A: Record,
    B: Record

// exponentialMovingAverage calculates the exponential moving average of `n`
// number of values in the `_value` column giving more weight to more recent data.
//
// ### Exponential moving average rules
//
// - The first value of an exponential moving average over `n` values is the algebraic mean of `n` values.
// - Subsequent values are calculated as `y(t) = x(t) * k + y(t-1) * (1 - k)`, where:
//     - `y(t)` is the exponential moving average at time `t`.
//     - `x(t)` is the value at time `t`.
//     - `k = 2 / (1 + n)`.
// - The average over a period populated by only `null` values is `null`.
// - Exponential moving averages skip `null` values.
//
// ## Parameters
// - n: Number of values to average.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate a three point exponential moving average
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> exponentialMovingAverage(n: 3)
// ```
//
// ### Calculate a three point exponential moving average with null values
// ```
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> exponentialMovingAverage(n: 3)
// ```
//
// introduced: 0.37.0
// tags: transformations
//
builtin exponentialMovingAverage : (<-tables: [{B with _value: A}], n: int) => [{B with _value: A}] where A: Numeric

// fill replaces all null values in input tables with a non-null value.
//
// Output tables are the same as the input tables with all null values replaced
// in the specified column.
//
// ## Parameters
// - column: Column to replace null values in. Default is `_value`.
// - value: Constant value to replace null values with.
//
//   Value type must match the type of the specified column.
//
// - usePrevious: Replace null values with the previous non-null value.
//   Default is `false`.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Fill null values with a specified non-null value
// ```
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> fill(value: 0)
// ```
//
// ### Fill null values with the previous non-null value
// ```
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> fill(usePrevious: true)
// ```
//
// introduced: 0.14.0
// tags: transformations
//
builtin fill : (<-tables: [A], ?column: string, ?value: B, ?usePrevious: bool) => [C] where A: Record, C: Record

// filter filters data based on conditions defined in a predicate function (`fn`).
//
// Output tables have the same schema as the corresponding input tables.
//
// ## Parameters
// - fn: Single argument predicate function that evaluates `true` or `false`.
//
//   Records representing each row are passed to the function as `r`.
//   Records that evaluate to `true` are included in output tables.
//   Records that evaluate to _null_ or `false` are excluded from output tables.
//
// - onEmpty: Action to take with empty tables. Default is `drop`.
//
//   **Supported values**:
//   - **keep**: Keep empty tables.
//   - **drop**: Drop empty tables.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Filter based on InfluxDB measurement, field, and tag
// ```no_run
// from(bucket: "example-bucket")
//     |> range(start: -1h)
//     |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system" and r.cpu == "cpu-total")
// ```
//
// ### Keep empty tables when filtering
// ```
// import "sampledata"
// import "experimental/table"
//
// < sampledata.int()
// >     |> filter(fn: (r) => r._value > 18, onEmpty: "keep")
// ```
//
// ### Filter values based on thresholds
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> filter(fn: (r) => r._value > 0 and r._value < 10 )
// ```
//
// introduced: 0.7.0
// tags: transformations,filters
//
builtin filter : (<-tables: [A], fn: (r: A) => bool, ?onEmpty: string) => [A] where A: Record

// first returns the first non-null record from each input table.
//
// **Note**: `first()` drops empty tables.
//
// ## Parameters
// - column: Column to operate on. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the first row in each input table
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> first()
// ```
//
// introduced: 0.7.0
// tags: transformations,selectors
//
builtin first : (<-tables: [A], ?column: string) => [A] where A: Record

// group regroups input data by modifying group key of input tables.
//
// **Note**: Group does not gaurantee sort order.
// To ensure data is sorted correctly, use `sort()` after `group()`.
//
// ## Parameters
// - columns: List of columns to use in the grouping operation. Default is `[]`.
//
//   **Note**: When `columns` is set to an empty array, `group()` ungroups
//   all data merges it into a single output table.
//
// - mode: Grouping mode. Default is `by`.
//
//   **Avaliable modes**:
//   - **by**: Group by columns defined in the `columns` parameter.
//   - **except**: Group by all columns _except_ those in defined in the
//     `columns` parameter.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Group by specific columns
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> group(columns: ["_time", "tag"])
// ```
//
// ### Group by everything except time
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> group(columns: ["_time"], mode: "except")
// ```
//
// ### Ungroup data
// ```
// import "sampledata"
//
// // Merge all tables into a single table
// < sampledata.int()
// >     |> group()
// ```
//
// introduced: 0.7.0
// tags: transformations
//
builtin group : (<-tables: [A], ?mode: string, ?columns: [string]) => [A] where A: Record

// histogram approximates the cumulative distribution of a dataset by counting
// data frequencies for a list of bins.
//
// A bin is defined by an upper bound where all data points that are less than
// or equal to the bound are counted in the bin. Bin counts are cumulative.
//
// Each input table is converted into a single output table representing a single histogram.
// Each output table has the same group key as the corresponding input table.
// Columns not part of the group key are dropped.
// Output tables include additional columns for the upper bound and count of bins.
//
// ## Parameters
// - column: Column containing input values. Column must be of type float.
//   Default is `_value`.
// - upperBoundColumn: Column to store bin upper bounds in. Default is `le`.
// - countColumn: Column to store bin counts in. Default is `_value`.
// - bins: List of upper bounds to use when computing the histogram frequencies.
//
//   Bins should contain a bin whose bound is the maximum value of the data set.
//   This value can be set to positive infinity if no maximum is known.
//
//   #### Bin helper functions
//   The following helper functions can be used to generated bins.
//
//   - linearBins()
//   - logarithmicBins()
//
// - normalize: Convert counts into frequency values between 0 and 1.
//   Default is `false`.
//
//   **Note**: Normalized histograms cannot be aggregated by summing their counts.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Create a cumulative histogram
// ```
// import "sampledata"
//
// < sampledata.float()
// >     |> histogram(bins: [0.0, 5.0, 10.0, 20.0])
// ```
//
// ### Create a cumulative histogram with dynamically generated bins
// ```
// import "sampledata"
//
// < sampledata.float()
// >     |> histogram(bins: linearBins(start: 0.0, width: 4.0, count: 3))
// ```
//
// introduced: 0.7.0
// tags: transformations
//
builtin histogram : (
        <-tables: [A],
        ?column: string,
        ?upperBoundColumn: string,
        ?countColumn: string,
        bins: [float],
        ?normalize: bool,
    ) => [B]
    where
    A: Record,
    B: Record

// histogramQuantile approximates a quantile given a histogram that approximates
// the cumulative distribution of the dataset.
//
// Each input table represents a single histogram.
// The histogram tables must have two columns – a count column and an upper bound column.
//
// The count is the number of values that are less than or equal to the upper bound value.
// The table can have any number of records, each representing a bin in the histogram.
// The counts must be monotonically increasing when sorted by upper bound.
// If any values in the count column or upper bound column are _null_, it returns an error.
// The count and upper bound columns must **not** be part of the group key.
//
// The quantile is computed using linear interpolation between the two closest bounds.
// If either of the bounds used in interpolation are infinite, the other finite
// bound is used and no interpolation is performed.
//
// ### Output tables
// Output tables have the same group key as corresponding input tables.
// Columns not part of the group key are dropped.
// A single value column of type float is added.
// The value column represents the value of the desired quantile from the histogram.
//
// ## Parameters
// - quantile: Quantile to compute. Value must be between 0 and 1.
// - countColumn: Column containing histogram bin counts. Default is `_value`.
// - upperBoundColumn: Column containing histogram bin upper bounds.
//   Default is `le`.
// - valueColumn: Column to store the computed quantile in. Default is `_value.
// - minValue: Assumed minimum value of the dataset. Default is `0.0`.
//
//   If the quantile falls below the lowest upper bound, interpolation is
//   performed between `minValue` and the lowest upper bound.
//   When `minValue` is equal to negative infinity, the lowest upper bound is used.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Compute the 90th quantile of a histogram
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.float()
// #         |> histogram(bins: [0.0, 5.0, 10.0, 20.0])
// #
// < data
// >     |> histogramQuantile(quantile: 0.9)
// ```
builtin histogramQuantile : (
        <-tables: [A],
        ?quantile: float,
        ?countColumn: string,
        ?upperBoundColumn: string,
        ?valueColumn: string,
        ?minValue: float,
    ) => [B]
    where
    A: Record,
    B: Record

// holtWinters applies the Holt-Winters forecasting method to input tables.
//
// The Holt-Winters method predicts `n` seasonally-adjusted values for the
// specified column at the specified interval. For example, if interval is six
// minutes (`6m`) and `n` is `3`, results include three predicted values six
// minutes apart.
//
// #### Seasonality
// `seasonality` delimits the length of a seasonal pattern according to interval.
// If the interval is two minutes (`2m`) and `seasonality` is `4`, then the
// seasonal pattern occurs every eight minutes or every four data points.
// If your interval is two months (`2mo`) and `seasonality` is `4`, then the
// seasonal pattern occurs every eight months or every four data points.
// If data doesn’t have a seasonal pattern, set `seasonality` to `0`.
//
// #### Space values at even time intervals
// `holtWinters()` expects values to be spaced at even time intervales.
// To ensure values are spaced evenly in time, `holtWinters()` applies the
// following rules:
//
// - Data is grouped into time-based "buckets" determined by the interval.
// - If a bucket includes many values, the first value is used.
// - If a bucket includes no values, a missing value (`null`) is added for that bucket.
//
// By default, `holtWinters()` uses the first value in each time bucket to run
// the Holt-Winters calculation. To specify other values to use in the
// calculation, use `aggregateWindow` to normalize irregular times and apply
// an aggregate or selector transformation.
//
// #### Fitted model
// `holtWinters()` applies the [Nelder-Mead optimization](https://en.wikipedia.org/wiki/Nelder%E2%80%93Mead_method)
// to include "fitted" data points in results when `withFit` is set to `true`.
//
// #### Null timestamps
// `holtWinters()` discards rows with null timestamps before running the
// Holt-Winters calculation.
//
// #### Null values
// `holtWinters()` treats `null` values as missing data points and includes them
// in the Holt-Winters calculation.
//
// ## Parameters
// - n: Number of values to predict.
// - interval: Interval between two data points.
// - withFit: Return fitted data in results. Default is `false`.
// - column: Column to operate on. Default is `_value`.
// - timeColumn: Column containing time values to use in the calculating.
//   Default is `_time`.
// - seasonality: Number of points in a season. Default is `0`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Use holtWinters to predict future values
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> holtWinters(n: 6, interval: 10s)
// ```
//
// ### Use holtWinters with seasonality to predict future values
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> holtWinters(n: 4, interval: 10s, seasonality: 4)
// ```
//
// ### Use the holtWinters fitted model to predict future values
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> holtWinters(n: 3, interval: 10s, withFit: true)
// ```
//
// introduced: 0.38.0
// tags: transformations
//
builtin holtWinters : (
        <-tables: [A],
        n: int,
        interval: duration,
        ?withFit: bool,
        ?column: string,
        ?timeColumn: string,
        ?seasonality: int,
    ) => [B]
    where
    A: Record,
    B: Record

// hourSelection retains all rows with time values in a specified hour range.
//
// ## Parameters
// - start: First hour of the hour range (inclusive). Hours range from `[0-23]`.
// - stop: Last hour of the hour range (inclusive). Hours range from `[0-23]`.
// - timeColumn: Column that contains the time value. Default is `_time`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Filter by business hours
// ```
// # import "generate"
// #
// # data = generate.from(count: 8, fn: (n) => n * n, start: 2021-01-01T00:00:00Z, stop: 2021-01-02T00:00:00Z)
// #
// < data
// >     |> hourSelection(start: 9, stop: 17)
// ```
//
// introduced: 0.39.0
// tags: transformations, date/time, filters
//
builtin hourSelection : (<-tables: [A], start: int, stop: int, ?timeColumn: string) => [A] where A: Record

// integral computes the area under the curve per unit of time of subsequent non-null records.
//
// `integral()` requires `_start` and `_stop` columns that are part of the group key.
// The curve is defined using `_time` as the domain and record values as the range.
//
// ## Parameters
// - unit: Unit of time to use to compute the integral.
// - column: Column to operate on. Default is `_value`.
// - timeColumn: Column that contains time values to use in the operation.
//   Default is `_time`.
// - interpolate: Type of interpolation to use. Default is `""`.
//
//   **Available interplation types**:
//   - linear
//   - _empty string for no interpolation_
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate the integral
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.int()
// #         |> range(start: sampledata.start, stop: sampledata.stop)
// #
// < data
// >     |> integral(unit: 10s)
// ```
//
// ### Calculate the integral with linear interpolation
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.int(includeNull: true)
// #         |> range(start: sampledata.start, stop: sampledata.stop)
// #
// < data
// >     |> integral(unit: 10s, interpolate: "linear")
// ```
//
// introduced: 0.7.0
// tags: transformations, aggregates
//
builtin integral : (
        <-tables: [A],
        ?unit: duration,
        ?timeColumn: string,
        ?column: string,
        ?interpolate: string,
    ) => [B]
    where
    A: Record,
    B: Record

// join merges two streams of tables into a single output stream based on columns with equal values.
// Null values are not considered equal when comparing column values.
// The resulting schema is the union of the input schemas.
// The resulting group key is the union of the input group keys.
//
// #### Output data
// The schema and group keys of the joined output output data is the union of
// the input schemas and group keys.
// Columns that exist in both input streams that are not part specified as
// columns to join on are renamed using the pattern `<column>_<table>` to
// prevent ambiguity in joined tables.
//
// ### Join vs union
// `join()` creates new rows based on common values in one or more specified columns.
// Output rows also contain the differing values from each of the joined streams.
// `union()` does not modify data in rows, but unions separate streams of tables
// into a single stream of tables and groups rows of data based on existing group keys.
//
// ## Parameters
// - tables: Record containing two input streams to join.
// - on: List of columns to join on.
// - method: Join method. Default is `inner`.
//
//   **Supported methods**:
//   - inner
//
// ## Examples
//
// ### Join two streams of tables
// ```
// import "generate"
//
// t1 =
//     generate.from(count: 4, fn: (n) => n + 1, start: 2021-01-01T00:00:00Z, stop: 2021-01-05T00:00:00Z)
//         |> set(key: "tag", value: "foo")
//
// t2 =
//     generate.from(count: 4, fn: (n) => n * (-1), start: 2021-01-01T00:00:00Z, stop: 2021-01-05T00:00:00Z)
//         |> set(key: "tag", value: "foo")
//
// > join(tables: {t1: t1, t2: t2}, on: ["_time", "tag"])
// ```
//
// ### Join data from separate data sources
// ```no_run
// import "sql"
//
// sqlData =
//     sql.from(
//         driverName: "postgres",
//         dataSourceName: "postgresql://username:password@localhost:5432",
//         query: "SELECT * FROM example_table",
//     )
//
// tsData =
//     from(bucket: "example-bucket")
//         |> range(start: -1h)
//         |> filter(fn: (r) => r._measurement == "example-measurement")
//         |> filter(fn: (r) => exists r.sensorID)
//
// join(tables: {sql: sqlData, ts: tsData}, on: ["_time", "sensorID"])
// ```
//
// introduced: 0.7.0
// tags: transformations
//
builtin join : (<-tables: A, ?method: string, ?on: [string]) => [B] where A: Record, B: Record

// kaufmansAMA calculates the Kaufman’s Adaptive Moving Average (KAMA) using
// values in input tables.
//
// Kaufman’s Adaptive Moving Average is a trend-following indicator designed to
// account for market noise or volatility.
//
// ## Parameters
// - n: Period or number of points to use in the calculation.
// - column: Column to operate on. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Caclulate Kaufman's Adaptive Moving Average for input data
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> kaufmansAMA(n: 3)
// ```
//
// introduced: 0.40.0
// tags: transformations
//
builtin kaufmansAMA : (<-tables: [A], n: int, ?column: string) => [B] where A: Record, B: Record

// keep returns a stream of tables containing only the specified columns.
//
// Columns in the group key that are not specifed in the `columns` parameter or
// identified by the `fn` parameter are removed from the group key and dropped
// from output tables. `keep()` is the inverse of `drop()`.
//
// ## Parameters
// - columns: Columns to keep in output tables. Cannot be used with `fn`.
// - fn: Predicate function that takes a column name as a parameter (column) and
//   returns a boolean indicating whether or not the column should be kept in
//   output tables. Cannot be used with `columns`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Keep a list of columns
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> keep(columns: ["_time", "_value"])
// ```
//
// ### Keep columns matching a predicate
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> keep(fn: (column) => column =~ /^_?t/)
// ```
//
// introduced: 0.7.0
// tags: transformations
//
builtin keep : (<-tables: [A], ?columns: [string], ?fn: (column: string) => bool) => [B] where A: Record, B: Record

// keyValues returns a stream of tables with each input tables' group key and
// two columns, _key and _value, that correspond to unique column label and value
// pairs for each input table.
//
// ## Parameters
// - keyColumns: List of columns from which values are extracted.
//
//   All columns must be of the same type.
//   Each input table must have all of the columns in the `keyColumns` parameter.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Get key values from explicitly defined columns
// ```
// # import "array"
// #
// # data =
// #     array.from(
// #         rows: [
// #             {
// #                 _field: "co",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:21:57Z,
// #                 _value: 0.31,
// #             },
// #             {
// #                 _field: "co",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:07Z,
// #                 _value: 0.295,
// #             },
// #             {
// #                 _field: "co",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:17Z,
// #                 _value: 0.314,
// #             },
// #             {
// #                 _field: "co",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:27Z,
// #                 _value: 0.313,
// #             },
// #             {
// #                 _field: "humidity",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:21:57Z,
// #                 _value: 36.03,
// #             },
// #             {
// #                 _field: "humidity",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:07Z,
// #                 _value: 36.07,
// #             },
// #             {
// #                 _field: "humidity",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:17Z,
// #                 _value: 36.1,
// #             },
// #             {
// #                 _field: "humidity",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:27Z,
// #                 _value: 36.12,
// #             },
// #             {
// #                 _field: "temperature",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:21:57Z,
// #                 _value: 70.84,
// #             },
// #             {
// #                 _field: "temperature",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:07Z,
// #                 _value: 70.86,
// #             },
// #             {
// #                 _field: "temperature",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:17Z,
// #                 _value: 70.89,
// #             },
// #             {
// #                 _field: "temperature",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:27Z,
// #                 _value: 70.85,
// #             },
// #         ],
// #     )
// #         |> group(columns: ["_time", "_value"], mode: "except")
// #
// < data
// >     |> keyValues(keyColumns: ["sensorID", "_field"])
// ```
//
// introduced: 0.13.0
// tags: transformations
//
builtin keyValues : (<-tables: [A], ?keyColumns: [string]) => [{C with _key: string, _value: B}]
    where
    A: Record,
    C: Record

// keys returns the columns that are in the group key of each input table.
//
// Each output table contains a row for each group key column label.
// A single group key column label is stored in the specified `column` for each row.
// All columns not in the group key are dropped.
//
// ## Parameters
// - column: Column to store group key labels in. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return group key columns for each input table
// ```
// # import "array"
// #
// # data =
// #     array.from(
// #         rows: [
// #             {
// #                 _field: "co",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:21:57Z,
// #                 _value: 0.31,
// #             },
// #             {
// #                 _field: "co",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:07Z,
// #                 _value: 0.295,
// #             },
// #             {
// #                 _field: "co",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:17Z,
// #                 _value: 0.314,
// #             },
// #             {
// #                 _field: "co",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:27Z,
// #                 _value: 0.313,
// #             },
// #             {
// #                 _field: "humidity",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:21:57Z,
// #                 _value: 36.03,
// #             },
// #             {
// #                 _field: "humidity",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:07Z,
// #                 _value: 36.07,
// #             },
// #             {
// #                 _field: "humidity",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:17Z,
// #                 _value: 36.1,
// #             },
// #             {
// #                 _field: "humidity",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:27Z,
// #                 _value: 36.12,
// #             },
// #             {
// #                 _field: "temperature",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:21:57Z,
// #                 _value: 70.84,
// #             },
// #             {
// #                 _field: "temperature",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:07Z,
// #                 _value: 70.86,
// #             },
// #             {
// #                 _field: "temperature",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:17Z,
// #                 _value: 70.89,
// #             },
// #             {
// #                 _field: "temperature",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:27Z,
// #                 _value: 70.85,
// #             },
// #         ],
// #     )
// #         |> group(columns: ["_time", "_value"], mode: "except")
// #
// < data
// >     |> keys()
// ```
//
// ### Return all distinct group key columns in a single table
// ```
// # import "array"
// #
// # data =
// #     array.from(
// #         rows: [
// #             {
// #                 _field: "co",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:21:57Z,
// #                 _value: 0.31,
// #             },
// #             {
// #                 _field: "co",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:07Z,
// #                 _value: 0.295,
// #             },
// #             {
// #                 _field: "co",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:17Z,
// #                 _value: 0.314,
// #             },
// #             {
// #                 _field: "co",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:27Z,
// #                 _value: 0.313,
// #             },
// #             {
// #                 _field: "humidity",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:21:57Z,
// #                 _value: 36.03,
// #             },
// #             {
// #                 _field: "humidity",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:07Z,
// #                 _value: 36.07,
// #             },
// #             {
// #                 _field: "humidity",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:17Z,
// #                 _value: 36.1,
// #             },
// #             {
// #                 _field: "humidity",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:27Z,
// #                 _value: 36.12,
// #             },
// #             {
// #                 _field: "temperature",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:21:57Z,
// #                 _value: 70.84,
// #             },
// #             {
// #                 _field: "temperature",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:07Z,
// #                 _value: 70.86,
// #             },
// #             {
// #                 _field: "temperature",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:17Z,
// #                 _value: 70.89,
// #             },
// #             {
// #                 _field: "temperature",
// #                 _measurement: "airSensors",
// #                 sensorID: "TLM0100",
// #                 _time: 2021-09-08T14:22:27Z,
// #                 _value: 70.85,
// #             },
// #         ],
// #     )
// #         |> group(columns: ["_time", "_value"], mode: "except")
// #
// < data
//     |> keys()
//     |> keep(columns: ["_value"])
// >     |> distinct()
// ```
//
// ### Return group key columns as an array
// 1. Use `keys()` to replace the `_value` column with the group key labels.
// 2. Use `findColumn()` to return the `_value` column as an array.
//
// ```no_run
// import "sampledata"
//
// sampledata.int()
//     |> keys()
//     |> findColumn(fn: (key) => true, column: "_value")
//
// // Returns [tag]
// ```
//
// introduced: 0.13.0
// tags: transformations
//
builtin keys : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record

// last returns the last row with a non-null value from each input table.
//
// **Note**: `last()` drops empty tables.
//
// ## Parameters
// - column: Column to use to verify the existence of a value.
//   Default is `_value`.
//
//   If this column is `null` in the last record, `last()` returns the previous
//   record with a non-null value.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the last row from each input table
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> last()
// ```
//
// introduced: 0.7.0
// tags: transformations,selectors
//
builtin last : (<-tables: [A], ?column: string) => [A] where A: Record

// limit returns the first `n` rows after the specified `offset` from each input table.
//
// If an input table has less than `offset + n` rows, `limit()` returns all rows
// after the offset.
//
// ## Parameters
// - n: Maximum number of rows to return.
// - offset: Number of rows to skip per table before limiting to `n`.
//   Default is `0`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Limit results to the first three rows in each table
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> limit(n: 3)
// ```
//
// ### Limit results to the first three rows in each input table after the first two
// ```
// import "sampledata"
//
// sampledata.int()
//     |> limit(n: 3, offset: 2)
// ```
//
// introduced: 0.7.0
// tags: transformations, selectors
//
builtin limit : (<-tables: [A], n: int, ?offset: int) => [A]

// map iterates over and applies a function to input rows.
//
// Each input row is passed to the `fn` as a record, `r`.
// Each `r` property represents a column key-value pair.
// Output values must be of the following supported column types:
//
// - float
// - integer
// - unsigned integer
// - string
// - boolean
// - time
//
// ### Output data
// Output tables are the result of applying the map function (`fn`) to each
// record of the input tables. Output records are assigned to new tables based
// on the group key of the input stream.
// If output record contains a different value for a group key column, the
// record is regrouped into the appropriate table.
// If the output record drops a group key column, that column is removed from
// the group key.
//
// #### Preserve columns
// By default, map() drops any columns that:
//
// - Are not part of the input table’s group key.
// - Are not explicitly mapped in the `fn` function.
//
// This often results in the `_time` column being dropped.
// To preserve the `_time` column and other columns that do not meet the
// criteria above, use the `with` operator to extend the `r` record.
// The `with` operator updates a record property if it already exists, creates
// a new record property if it doesn’t exist, and includes all existing
// properties in the output record.
//
// ```no_run
// data
//     |> map(fn: (r) => ({ r with newColumn: r._value * 2 }))
// ```
//
// ## Parameters
// - fn: Single argument function to apply to each record.
//   The return value must be a record.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Square the value in each row
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> map(fn: (r) => ({ r with _value: r._value * r._value }))
// ```
//
// ### Create a new table with new columns
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> map(fn: (r) => ({time: r._time, source: r.tag, alert: if r._value > 10 then true else false}))
// ```
//
// ### Add new columns and preserve existing columns
// Use the `with` operator on the `r` record to preserve columns not directly
// operated on by the map operation.
//
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> map(fn: (r) => ({r with server: "server-${r.tag}", valueFloat: float(v: r._value)}))
// ```
//
// introduced: 0.7.0
// tags: transformations
//
builtin map : (<-tables: [A], fn: (r: A) => B, ?mergeKey: bool) => [B]

// max returns the row with the maximum value in a specified column from each
// input table.
//
// **Note:** `max()` drops empty tables.
//
// ## Parameters
// - column: Column to return maximum values from. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the row with the maximum value
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> max()
// ```
//
// introduced: 0.7.0
// tags: transformations, selectors
//
builtin max : (<-tables: [A], ?column: string) => [A] where A: Record

// mean returns the average of non-null values in a specified column from each
// input table.
//
// ## Parameters
// - column: Column to use to compute means. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the average of values in each input table
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> mean()
// ```
//
// introduce: 0.7.0
// tags: transformations, aggregates
//
builtin mean : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record

// min returns the row with the minimum value in a specified column from each
// input table.
//
// **Note:** `min()` drops empty tables.
//
// ## Parameters
// - column: Column to return minimum values from. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the row with the minimum value
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> min()
// ```
//
// introduced: 0.7.0
// tags: transformations, selectors
//
builtin min : (<-tables: [A], ?column: string) => [A] where A: Record

// mode returns the non-null value or values that occur most often in a
// specified column in each input table.
//
// If there are multiple modes, `mode()` returns all mode values in a sorted table.
// If there is no mode, `mode()` returns `null`.
//
// **Note**: `mode()` drops empty tables.
//
// ## Parameters
// - column: Column to return the mode from. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the mode of each input table
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> mode()
// ```
//
// introduced: 0.36.0
// tags: transformtions, aggregates
//
builtin mode : (<-tables: [A], ?column: string) => [{C with _value: B}] where A: Record, C: Record

// movingAverage calculates the mean of non-null values using the current value
// and `n - 1` previous values in the `_values` column.
//
// ### Moving average rules
// - The average over a period populated by `n` values is equal to their algebraic mean.
// - The average over a period populated by only `null` values is `null`.
// - Moving averages skip `null` values.
// - If `n` is less than the number of records in a table, `movingAverage()`
//   returns the average of the available values.
//
// ## Parameters
// - n: Number of values to average.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate a three point moving average
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> movingAverage(n: 3)
// ```
//
// ### Calculate a three point moving average with null values
// ```
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> movingAverage(n: 3)
// ```
//
// introduced: 0.35.0
// tags: transformations
//
builtin movingAverage : (<-tables: [{B with _value: A}], n: int) => [{B with _value: float}] where A: Numeric

// quantile returns rows from each input table with values that fall within a
// specified quantile or returns the row with the value that represents the
// specified quantile.
//
// `quantile()` supports columns with float values.
//
// ### Function behavior
// `quantile()` acts as an aggregate or selector transformation depending on the
// specified `method`.
//
// - **Aggregate**: When using the `estimate_tdigest` or `exact_mean` methods,
// `quantile()` acts as an aggregate transformation and outputs non-null records
// with values that fall within the specified quantile.
// - **Selector**: When using the `exact_selector` method, `quantile()` acts as
// a selector selector transformation and outputs the non-null record with the
// value that represents the specified quantile.
//
// ## Parameters
// - column: Column to use to compute the quantile. Default is `_value`.
// - q: Quantile to compute. Must be between `0.0` and `1.0`.
// - method: Computation method. Default is `estimate_tdigest`.
//
//     **Avaialable methods**:
//
//     - **estimate_tdigest**: Aggregate method that uses a
//       [t-digest data structure](https://github.com/tdunning/t-digest) to
//       compute an accurate quantile estimate on large data sources.
//     - **exact_mean**: Aggregate method that takes the average of the tw
//       points closest to the quantile value.
//     - **exact_selector**: Selector method that returns the row with the value
//       for which at least `q` points are less than.
//
// - compression: Number of centroids to use when compressing the dataset.
//   Default is `1000.0`.
//
//   A larger number produces a more accurate result at the cost of increased
//   memory requirements.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Quantile as an aggregate
// ```
// import "sampledata"
//
// < sampledata.float()
// >     |> quantile(q: 0.99, method: "estimate_tdigest")
// ```
//
// ### Quantile as a selector
// ```
// import "sampledata"
//
// < sampledata.float()
// >     |> quantile(q: 0.5, method: "exact_selector")
// ```
//
// introduced: 0.24.0
// tags: transformations, aggregates, selectors
//
builtin quantile : (
        <-tables: [A],
        ?column: string,
        q: float,
        ?compression: float,
        ?method: string,
    ) => [A]
    where
    A: Record

// pivot collects unique values stored vertically (column-wise) and aligns them
// horizontally (row-wise) into logical sets.
//
// ### Output data
// The group key of the resulting table is the same as the input tables,
// excluding columns found in the `columnKey` and `valueColumn `parameters.
// These columns are not part of the resulting output table and are dropped from
// the group key.
//
// Every input row should have a 1:1 mapping to a particular row and column
// combination in the output table. Row and column combinations are determined
// by the `rowKey` and `columnKey` parameters. In cases where more than one
// value is identified for the same row and column pair, the last value
// encountered in the set of table rows is used as the result.
//
// The output is constructed as follows:
//
// - The set of columns for the new table is the `rowKey` unioned with the group key,
//   but excluding the columns indicated by the `columnKey` and the `valueColumn`.
// - A new column is added to the set of columns for each unique value
//   identified by the `columnKey` parameter.
// - The label of a new column is the concatenation of the values of `columnKey`
//   using `_` as a separator. If the value is null, "null" is used.
// - A new row is created for each unique value identified by the
//   `rowKey` parameter.
// - For each new row, values for group key columns stay the same, while values
//   for new columns are determined from the input tables by the value in
//   `valueColumn` at the row identified by the `rowKey` values and the new
//   column’s label. If no value is found, the value is set to `null`.
// - Any column that is not part of the group key or not specified in the
//   `rowKey`, `columnKey` and `valueColumn` parameters is dropped.
//
// ## Parameters
// - rowKey: Columns to use to uniquely identify an output row.
// - columnKey: Columns to use to identify new output columns.
// - valueColumn: Column to use to populate the value of pivoted `columnKey` columns.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Align fields into rows based on time
// ```
// # import "csv"
// #
// # csvData = "#datatype,string,long,dateTime:RFC3339,string,string,double
// # #group,false,false,false,true,true,false
// # #default,,,,,,
// # ,result,table,_time,_measurement,_field,_value
// # ,,0,1970-01-01T00:00:01Z,m1,f1,1.0
// # ,,0,1970-01-01T00:00:01Z,m1,f2,2.0
// # ,,0,1970-01-01T00:00:01Z,m1,f3,
// # ,,0,1970-01-01T00:00:02Z,m1,f1,4.0
// # ,,0,1970-01-01T00:00:02Z,m1,f2,5.0
// # ,,0,1970-01-01T00:00:02Z,m1,f3,6.0
// # ,,0,1970-01-01T00:00:03Z,m1,f1,
// # ,,0,1970-01-01T00:00:03Z,m1,f2,7.0
// # ,,0,1970-01-01T00:00:04Z,m1,f3,8.0
// # "
// #
// # data = csv.from(csv: csvData)
// #
// < data
// >     |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
// ```
//
// ### Associate values to tags by time
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> pivot(rowKey: ["_time"], columnKey: ["tag"], valueColumn: "_value")
// ```
//
// introduced: 0.7.0
// tags: transformations
//
builtin pivot : (<-tables: [A], rowKey: [string], columnKey: [string], valueColumn: string) => [B]
    where
    A: Record,
    B: Record

// range filters rows based on time bounds.
//
// Input data must have a `_time` column of type time.
// Rows with a null value in the `_time` are filtered.
// Each input table’s group key value is modified to fit within the time bounds.
// Tables with all rows outside the time bounds are filtered entirely.
//
// ## Parameters
// - start: Earliest time to include in results.
//
//   Results _include_ rows with `_time` values that match the specified start time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//   For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//   Durations are relative to `now()`.
//
// - stop: Latest time to include in results. Default is `now()`.
//
//   Results _exclude_ rows with `_time` values that match the specified start time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//   For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//   Durations are relative to `now()`.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Query a time range relative to now
// ```no_run
// from(bucket:"example-bucket")
//     |> range(start: -12h)
// ```
//
// ### Query an absolute time range
// ```no_run
// from(bucket:"example-bucket")
//     |> range(start: 2021-05-22T23:30:00Z, stop: 2021-05-23T00:00:00Z)
// ```
//
// ### Query an absolute time range using Unix timestamps
// ```no_run
// from(bucket:"example-bucket")
//     |> range(start: 1621726200000000000, stop: 1621728000000000000)
// ```
//
// introduced: 0.7.0
// tags: transformations, filters
//
builtin range : (
        <-tables: [{A with _time: time}],
        start: B,
        ?stop: C,
    ) => [{A with _time: time, _start: time, _stop: time}]

// reduce aggregates rows in each input table using a reducer function (`fn`).
//
// The output for each table is the group key of the table with columns
// corresponding to each field in the reducer record.
// If the reducer record contains a column with the same name as a group key column,
// the group key column’s value is overwritten, and the outgoing group key is changed.
// However, if two reduced tables write to the same destination group key, the
// function returns an error.
//
// ### Dropped columns
// `reduce()` drops any columns that:
//
// - Are not part of the input table’s group key.
// - Are not explicitly mapped in the `identity` record or the reducer function (`fn`).
//
// ## Parameters
// - fn: Reducer function to apply to each row record (`r`).
//
//   The reducer function accepts two parameters:
//
//   - **r**: Record representing the current row.
//   - **accumulator**: Record returned from the reducer function's operation on
//     the previous row.
//
// - identity: Record that defines the reducer record and provides initial values
//   for the reducer operation on the first row.
//
//   May be used more than once in asynchronous processing use cases.
//   The data type of values in the identity record determine the data type of
//   output values.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Compute the sum of the value column
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> reduce(fn: (r, accumulator) => ({sum: r._value + accumulator.sum}), identity: {sum: 0})
// ```
//
// ### Compute the sum and count in a single reducer
// ```
// import "sampledata"
//
// < sampledata.int()
//     |> reduce(
//         fn: (r, accumulator) => ({sum: r._value + accumulator.sum, count: accumulator.count + 1}),
//         identity: {sum: 0, count: 0},
// >     )
// ```
//
// ### Compute the product of all values
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> reduce(fn: (r, accumulator) => ({prod: r._value * accumulator.prod}), identity: {prod: 1})
// ```
//
// ### Calculate the average of all values
// ```
// import "sampledata"
//
// < sampledata.int()
//     |> reduce(
//         fn: (r, accumulator) =>
//             ({
//                 count: accumulator.count + 1,
//                 total: accumulator.total + r._value,
//                 avg: float(v: accumulator.total + r._value) / float(v: accumulator.count + 1),
//             }),
//         identity: {count: 0, total: 0, avg: 0.0},
// >     )
// ```
//
// introduced: 0.23.0
// tags: transformations, aggregates
//
builtin reduce : (<-tables: [A], fn: (r: A, accumulator: B) => B, identity: B) => [C]
    where
    A: Record,
    B: Record,
    C: Record

// relativeStrengthIndex measures the relative speed and change of values in input tables.
//
// ### Relative strength index (RSI) rules
// - The general equation for calculating a relative strength index (RSI) is
//   `RSI = 100 - (100 / (1 + (AVG GAIN / AVG LOSS)))`.
// - For the first value of the RSI, `AVG GAIN` and `AVG LOSS` are averages of the `n` period.
// - For subsequent calculations:
//   - `AVG GAIN` = `((PREVIOUS AVG GAIN) * (n - 1)) / n`
//   - `AVG LOSS` = `((PREVIOUS AVG LOSS) * (n - 1)) / n`
// - `relativeStrengthIndex()` ignores `null` values.
//
// ### Output tables
// For each input table with `x` rows, `relativeStrengthIndex()` outputs a table
// with `x - n` rows.
//
// ## Parameters
// - n: Number of values to use to calculate the RSI.
// - columns: Columns to operate on. Default is `["_value"]`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate a three point relative strength index
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> relativeStrengthIndex(n: 3)
// ```
//
// introduced: 0.38.0
// tags: transformations
//
builtin relativeStrengthIndex : (<-tables: [A], n: int, ?columns: [string]) => [B] where A: Record, B: Record

// rename renames columns in a table.
//
// If a column in group key is renamed, the column name in the group key is updated.
//
// ## Parameters
// - columns: Record that maps old column names to new column names.
// - fn: Function that takes the current column name (`column`) and returns a
//   new column name.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Map column names to new column names
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> rename(columns: {tag: "uid", _value: "val"})
// ```
//
// ### Rename columns using a function
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> rename(fn: (column) => "${column}_new")
// ```
//
// introduced: 0.7.0
// tags: transformations
//
builtin rename : (<-tables: [A], ?fn: (column: string) => string, ?columns: B) => [C]
    where
    A: Record,
    B: Record,
    C: Record
builtin sample : (<-tables: [A], n: int, ?pos: int, ?column: string) => [A] where A: Record
builtin set : (<-tables: [A], key: string, value: string) => [A] where A: Record
builtin tail : (<-tables: [A], n: int, ?offset: int) => [A]
builtin timeShift : (<-tables: [A], duration: duration, ?columns: [string]) => [A]
builtin skew : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin spread : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin sort : (<-tables: [A], ?columns: [string], ?desc: bool) => [A] where A: Record
builtin stateTracking : (
        <-tables: [A],
        fn: (r: A) => bool,
        ?countColumn: string,
        ?durationColumn: string,
        ?durationUnit: duration,
        ?timeColumn: string,
    ) => [B]
    where
    A: Record,
    B: Record

builtin stddev : (<-tables: [A], ?column: string, ?mode: string) => [B] where A: Record, B: Record
builtin sum : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin tripleExponentialDerivative : (<-tables: [{B with _value: A}], n: int) => [{B with _value: float}]
    where
    A: Numeric,
    B: Record
builtin union : (tables: [[A]]) => [A] where A: Record
builtin unique : (<-tables: [A], ?column: string) => [A] where A: Record

builtin _window : (
        <-tables: [A],
        every: duration,
        period: duration,
        offset: duration,
        location: {zone: string, offset: duration},
        timeColumn: string,
        startColumn: string,
        stopColumn: string,
        createEmpty: bool,
    ) => [B]
    where
    A: Record,
    B: Record

window = (
    tables=<-,
    every=0s,
    period=0s,
    offset=0s,
    location=location,
    timeColumn="_time",
    startColumn="_start",
    stopColumn="_stop",
    createEmpty=false,
) =>
    tables
        |> _window(
            every,
            period,
            offset,
            location,
            timeColumn,
            startColumn,
            stopColumn,
            createEmpty,
        )

builtin yield : (<-tables: [A], ?name: string) => [A] where A: Record

// stream/table index functions
builtin tableFind : (<-tables: [A], fn: (key: B) => bool) => [A] where A: Record, B: Record
builtin getColumn : (<-table: [A], column: string) => [B] where A: Record
builtin getRecord : (<-table: [A], idx: int) => A where A: Record
builtin findColumn : (<-tables: [A], fn: (key: B) => bool, column: string) => [C] where A: Record, B: Record
builtin findRecord : (<-tables: [A], fn: (key: B) => bool, idx: int) => A where A: Record, B: Record

// type conversion functions
builtin bool : (v: A) => bool
builtin bytes : (v: A) => bytes
builtin duration : (v: A) => duration
builtin float : (v: A) => float
builtin int : (v: A) => int
builtin string : (v: A) => string
builtin time : (v: A) => time
builtin uint : (v: A) => uint

// contains function
builtin contains : (value: A, set: [A]) => bool where A: Nullable

// other builtins
builtin inf : duration
builtin length : (arr: [A]) => int
builtin linearBins : (start: float, width: float, count: int, ?infinity: bool) => [float]
builtin logarithmicBins : (start: float, factor: float, count: int, ?infinity: bool) => [float]

// Time weighted average where values at the beginning and end of the range are linearly interpolated.
timeWeightedAvg = (tables=<-, unit) =>
    tables
        |> integral(unit: unit, interpolate: "linear")
        |> map(
            fn: (r) =>
                ({r with _value: r._value * float(v: uint(v: unit)) / float(v: int(v: r._stop) - int(v: r._start))}),
        )

// covariance function with automatic join
cov = (x, y, on, pearsonr=false) =>
    join(tables: {x: x, y: y}, on: on)
        |> covariance(pearsonr: pearsonr, columns: ["_value_x", "_value_y"])
pearsonr = (x, y, on) => cov(x: x, y: y, on: on, pearsonr: true)

_fillEmpty = (tables=<-, createEmpty) =>
    if createEmpty then
        tables
            |> table.fill()
    else
        tables

// aggregateWindow applies an aggregate function to fixed windows of time.
// The procedure is to window the data, perform an aggregate operation,
// and then undo the windowing to produce an output table for every input table.
aggregateWindow = (
    every,
    period=0s,
    fn,
    offset=0s,
    location=location,
    column="_value",
    timeSrc="_stop",
    timeDst="_time",
    createEmpty=true,
    tables=<-,
) =>
    tables
        |> window(
            every: every,
            period: period,
            offset: offset,
            location: location,
            createEmpty: createEmpty,
        )
        |> fn(column: column)
        |> _fillEmpty(createEmpty: createEmpty)
        |> duplicate(column: timeSrc, as: timeDst)
        |> window(every: inf, timeColumn: timeDst)

// Increase returns the total non-negative difference between values in a table.
// A main usage case is tracking changes in counter values which may wrap over time when they hit
// a threshold or are reset. In the case of a wrap/reset,
// we can assume that the absolute delta between two points will be at least their non-negative difference.
increase = (tables=<-, columns=["_value"]) =>
    tables
        |> difference(nonNegative: true, columns: columns, keepFirst: true, initialZero: true)
        |> cumulativeSum(columns: columns)

// median returns the 50th percentile.
median = (method="estimate_tdigest", compression=0.0, column="_value", tables=<-) =>
    tables
        |> quantile(q: 0.5, method: method, compression: compression, column: column)

// stateCount computes the number of consecutive records in a given state.
// The state is defined via the function fn. For each consecutive point for
// which the expression evaluates as true, the state count will be incremented
// When a point evaluates as false, the state count is reset.
//
// The state count will be added as an additional column to each record. If the
// expression evaluates as false, the value will be -1. If the expression
// generates an error during evaluation, the point is discarded, and does not
// affect the state count.
stateCount = (fn, column="stateCount", tables=<-) =>
    tables
        |> stateTracking(countColumn: column, fn: fn)

// stateDuration computes the duration of a given state.
// The state is defined via the function fn. For each consecutive point for
// which the expression evaluates as true, the state duration will be
// incremented by the duration between points. When a point evaluates as false,
// the state duration is reset.
//
// The state duration will be added as an additional column to each record. If the
// expression evaluates as false, the value will be -1. If the expression
// generates an error during evaluation, the point is discarded, and does not
// affect the state duration.
//
// Note that as the first point in the given state has no previous point, its
// state duration will be 0.
//
// The duration is represented as an integer in the units specified.
stateDuration = (
    fn,
    column="stateDuration",
    timeColumn="_time",
    unit=1s,
    tables=<-,
) =>
    tables
        |> stateTracking(durationColumn: column, timeColumn: timeColumn, fn: fn, durationUnit: unit)

// _sortLimit is a helper function, which sorts and limits a table.
_sortLimit = (n, desc, columns=["_value"], tables=<-) =>
    tables
        |> sort(columns: columns, desc: desc)
        |> limit(n: n)

// top sorts a table by columns and keeps only the top n records.
top = (n, columns=["_value"], tables=<-) =>
    tables
        |> _sortLimit(n: n, columns: columns, desc: true)

// top sorts a table by columns and keeps only the bottom n records.
bottom = (n, columns=["_value"], tables=<-) =>
    tables
        |> _sortLimit(n: n, columns: columns, desc: false)

// _highestOrLowest is a helper function, which reduces all groups into a single group by specific tags and a reducer function,
// then it selects the highest or lowest records based on the column and the _sortLimit function.
// The default reducer assumes no reducing needs to be performed.
_highestOrLowest = (
    n,
    _sortLimit,
    reducer,
    column="_value",
    groupColumns=[],
    tables=<-,
) =>
    tables
        |> group(columns: groupColumns)
        |> reducer()
        |> group(columns: [])
        |> _sortLimit(n: n, columns: [column])

// highestMax returns the top N records from all groups using the maximum of each group.
highestMax =
    (n, column="_value", groupColumns=[], tables=<-) =>
        tables
            |> _highestOrLowest(
                n: n,
                column: column,
                groupColumns: groupColumns,
                // TODO(nathanielc): Once max/min support selecting based on multiple columns change this to pass all columns.
                reducer: (tables=<-) => tables |> max(column: column),
                _sortLimit: top,
            )

// highestAverage returns the top N records from all groups using the average of each group.
highestAverage = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
            n: n,
            column: column,
            groupColumns: groupColumns,
            reducer: (tables=<-) => tables |> mean(column: column),
            _sortLimit: top,
        )

// highestCurrent returns the top N records from all groups using the last value of each group.
highestCurrent = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
            n: n,
            column: column,
            groupColumns: groupColumns,
            reducer: (tables=<-) => tables |> last(column: column),
            _sortLimit: top,
        )

// lowestMin returns the bottom N records from all groups using the minimum of each group.
lowestMin =
    (n, column="_value", groupColumns=[], tables=<-) =>
        tables
            |> _highestOrLowest(
                n: n,
                column: column,
                groupColumns: groupColumns,
                // TODO(nathanielc): Once max/min support selecting based on multiple columns change this to pass all columns.
                reducer: (tables=<-) => tables |> min(column: column),
                _sortLimit: bottom,
            )

// lowestAverage returns the bottom N records from all groups using the average of each group.
lowestAverage = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
            n: n,
            column: column,
            groupColumns: groupColumns,
            reducer: (tables=<-) => tables |> mean(column: column),
            _sortLimit: bottom,
        )

// lowestCurrent returns the bottom N records from all groups using the last value of each group.
lowestCurrent = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
            n: n,
            column: column,
            groupColumns: groupColumns,
            reducer: (tables=<-) => tables |> last(column: column),
            _sortLimit: bottom,
        )

// timedMovingAverage constructs a simple moving average over windows of 'period' duration
// eg: A 5 year moving average would be called as such:
//    movingAverage(1y, 5y)
timedMovingAverage = (every, period, column="_value", tables=<-) =>
    tables
        |> window(every: every, period: period)
        |> mean(column: column)
        |> duplicate(column: "_stop", as: "_time")
        |> window(every: inf)

// Double Exponential Moving Average computes the double exponential moving averages of the `_value` column.
// eg: A 5 point double exponential moving average would be called as such:
// from(bucket: "telegraf/autogen"):
//    |> range(start: -7d)
//    |> doubleEMA(n: 5)
doubleEMA = (n, tables=<-) =>
    tables
        |> exponentialMovingAverage(n: n)
        |> duplicate(column: "_value", as: "__ema")
        |> exponentialMovingAverage(n: n)
        |> map(fn: (r) => ({r with _value: 2.0 * r.__ema - r._value}))
        |> drop(columns: ["__ema"])

// kaufmansER computes the Kaufman's Efficiency Ratio (KER) of values in the
// `_value` column for each input table.
//
// Kaufman’s Efficiency Ratio indicator divides the absolute value of the Chande
// Momentum Oscillator by 100 to return a value between 0 and 1.
// Higher values represent a more efficient or trending market.
//
// ## Parameters
// - n: Period or number of points to use in the calculation.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Compute the Kaufman's Efficiency Ratio
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> kaufmansER(n: 3)
// ```
//
// introduced: 0.40.0
// tags: transformations
//
kaufmansER = (n, tables=<-) =>
    tables
        |> chandeMomentumOscillator(n: n)
        |> map(fn: (r) => ({r with _value: math.abs(x: r._value) / 100.0}))

// Triple Exponential Moving Average computes the triple exponential moving averages of the `_value` column.
// eg: A 5 point triple exponential moving average would be called as such:
// from(bucket: "telegraf/autogen"):
//    |> range(start: -7d)
//    |> tripleEMA(n: 5)
tripleEMA = (n, tables=<-) =>
    tables
        |> exponentialMovingAverage(n: n)
        |> duplicate(column: "_value", as: "__ema1")
        |> exponentialMovingAverage(n: n)
        |> duplicate(column: "_value", as: "__ema2")
        |> exponentialMovingAverage(n: n)
        |> map(fn: (r) => ({r with _value: 3.0 * r.__ema1 - 3.0 * r.__ema2 + r._value}))
        |> drop(columns: ["__ema1", "__ema2"])

// truncateTimeColumn takes in a time column t and a Duration unit and truncates each value of t to the given unit via map
// Change from _time to timeColumn once Flux Issue 1122 is resolved
truncateTimeColumn = (timeColumn="_time", unit, tables=<-) =>
    tables
        |> map(fn: (r) => ({r with _time: date.truncate(t: r._time, unit: unit)}))
toString = (tables=<-) => tables |> map(fn: (r) => ({r with _value: string(v: r._value)}))
toInt = (tables=<-) => tables |> map(fn: (r) => ({r with _value: int(v: r._value)}))
toUInt = (tables=<-) => tables |> map(fn: (r) => ({r with _value: uint(v: r._value)}))
toFloat = (tables=<-) => tables |> map(fn: (r) => ({r with _value: float(v: r._value)}))
toBool = (tables=<-) => tables |> map(fn: (r) => ({r with _value: bool(v: r._value)}))
toTime = (tables=<-) => tables |> map(fn: (r) => ({r with _value: time(v: r._value)}))

// today returns the now() timestamp truncated to the day unit
today = () => date.truncate(t: now(), unit: 1d)
