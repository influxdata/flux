// Package universe provides options and primitive functions that are
// loaded into the Flux runtime by default and do not require an
// import statement.
//
// ## Metadata
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
//     |> range(start: -10h, stop: now())
// ```
//
// ### Define a custom now time
// ```no_run
// option now = () => 2022-01-01T00:00:00Z
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: date/time
//
option now = system.time

// chandeMomentumOscillator applies the technical momentum indicator developed
// by Tushar Chande to input data.
//
// The Chande Momentum Oscillator (CMO) indicator does the following:
//
// 1. Determines the median value of the each input table and calculates the
//    difference between the sum of rows with values greater than the median
//    and the sum of rows with values lower than the median.
// 2. Divides the result of step 1 by the sum of all data movement over a given
//    time period.
// 3. Multiplies the result of step 2 by 100 and returns a value between -100 and +100.
//
// #### Output tables
// For each input table with `x` rows, `chandeMomentumOscillator()` outputs a
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
// ## Metadata
// introduced: 0.39.0
// tags: transformations
//
builtin chandeMomentumOscillator : (<-tables: stream[A], n: int, ?columns: [string]) => stream[B]
    where
    A: Record,
    B: Record

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
// ## Metadata
// introduced: 0.14.0
// tags: transformations
//
builtin columns : (<-tables: stream[A], ?column: string) => stream[B] where A: Record, B: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations,aggregates
//
builtin count : (<-tables: stream[A], ?column: string) => stream[B] where A: Record, B: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations,aggregates
//
builtin covariance : (<-tables: stream[A], ?pearsonr: bool, ?valueDst: string, columns: [string]) => stream[B]
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
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin cumulativeSum : (<-tables: stream[A], ?columns: [string]) => stream[B] where A: Record, B: Record

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
// - nonNegative: Disallow negative derivative values. Default is `false`.
//
//   When `true`, if a value is less than the previous value, the function
//   assumes the previous value should have been a zero.
//
// - columns: List of columns to operate on. Default is `["_value"]`.
// - timeColumn: Column containing time values to use in the calculation.
//   Default is `_time`.
// - initialZero: Use zero (0) as the initial value in the derivative calculation
//   when the subsequent value is less than the previous value and `nonNegative` is
//   `true`. Default is `false`.
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
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin derivative : (
        <-tables: stream[A],
        ?unit: duration,
        ?nonNegative: bool,
        ?columns: [string],
        ?timeColumn: string,
        ?initialZero: bool,
    ) => stream[B]
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
// ## Metadata
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
// ## Metadata
// introduced: 0.7.1
// tags: transformations
//
builtin difference : (
        <-tables: stream[T],
        ?nonNegative: bool,
        ?columns: [string],
        ?keepFirst: bool,
        ?initialZero: bool,
    ) => stream[R]
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
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
builtin distinct : (<-tables: stream[A], ?column: string) => stream[B] where A: Record, B: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin drop : (<-tables: stream[A], ?fn: (column: string) => bool, ?columns: [string]) => stream[B]
    where
    A: Record,
    B: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin duplicate : (<-tables: stream[A], column: string, as: string) => stream[B] where A: Record, B: Record

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
// ## Metadata
// introduced: 0.36.0
// tags: transformations
//
builtin elapsed : (<-tables: stream[A], ?unit: duration, ?timeColumn: string, ?columnName: string) => stream[B]
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
// ## Metadata
// introduced: 0.37.0
// tags: transformations
//
builtin exponentialMovingAverage : (<-tables: stream[{B with _value: A}], n: int) => stream[{B with _value: A}]
    where
    A: Numeric

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
// ## Metadata
// introduced: 0.14.0
// tags: transformations
//
builtin fill : (<-tables: stream[A], ?column: string, ?value: B, ?usePrevious: bool) => stream[C]
    where
    A: Record,
    C: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations,filters
//
builtin filter : (<-tables: stream[A], fn: (r: A) => bool, ?onEmpty: string) => stream[A] where A: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations,selectors
//
builtin first : (<-tables: stream[A], ?column: string) => stream[A] where A: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin group : (<-tables: stream[A], ?mode: string, ?columns: [string]) => stream[A] where A: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin histogram : (
        <-tables: stream[A],
        ?column: string,
        ?upperBoundColumn: string,
        ?countColumn: string,
        bins: [float],
        ?normalize: bool,
    ) => stream[B]
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
//
// ## Metadata
// tags: transformations
builtin histogramQuantile : (
        <-tables: stream[A],
        ?quantile: float,
        ?countColumn: string,
        ?upperBoundColumn: string,
        ?valueColumn: string,
        ?minValue: float,
    ) => stream[B]
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
// ## Metadata
// introduced: 0.38.0
// tags: transformations
//
builtin holtWinters : (
        <-tables: stream[A],
        n: int,
        interval: duration,
        ?withFit: bool,
        ?column: string,
        ?timeColumn: string,
        ?seasonality: int,
    ) => stream[B]
    where
    A: Record,
    B: Record

// builtin _hourSelection used by hourSelection
builtin _hourSelection : (
        <-tables: stream[A],
        start: int,
        stop: int,
        location: {zone: string, offset: duration},
        ?timeColumn: string,
    ) => stream[A]
    where
    A: Record

// hourSelection filters rows by time values in a specified hour range.
//
// ## Parameters
// - start: First hour of the hour range (inclusive). Hours range from `[0-23]`.
// - stop: Last hour of the hour range (inclusive). Hours range from `[0-23]`.
// - location: Location used to determine timezone. Default is the `location` option.
// - timeColumn: Column that contains the time value. Default is `_time`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Filter by business hours
// ```
// # import "array"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2022-01-01T05:00:00Z, tag: "t1", _value: -2},
// #         {_time: 2022-01-01T09:00:10Z, tag: "t1", _value: 10},
// #         {_time: 2022-01-01T11:00:20Z, tag: "t1", _value: 7},
// #         {_time: 2022-01-01T16:00:30Z, tag: "t1", _value: 17},
// #         {_time: 2022-01-01T19:00:40Z, tag: "t1", _value: 15},
// #         {_time: 2022-01-01T20:00:50Z, tag: "t1", _value: 4},
// #     ],
// # )
// #
// < data
// >     |> hourSelection(start: 9, stop: 17)
// ```
//
// ## Metadata
// introduced: 0.39.0
// tags: transformations, date/time, filters
//
hourSelection = (
    tables=<-,
    start,
    stop,
    location=location,
    timeColumn="_time",
) =>
    tables
        |> _hourSelection(start, stop, location, timeColumn)

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations, aggregates
//
builtin integral : (
        <-tables: stream[A],
        ?unit: duration,
        ?timeColumn: string,
        ?column: string,
        ?interpolate: string,
    ) => stream[B]
    where
    A: Record,
    B: Record

// join merges two streams of tables into a single output stream based on columns with equal values.
// Null values are not considered equal when comparing column values.
// The resulting schema is the union of the input schemas.
// The resulting group key is the union of the input group keys.
//
// **Deprecated**: `join()` is deprecated in favor of [`join.inner()`](https://docs.influxdata.com/flux/v0.x/stdlib/join/inner/).
// The [`join` package](https://docs.influxdata.com/flux/v0.x/stdlib/join/) provides support
// for multiple join methods.
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
// ## Metadata
// introduced: 0.7.0
// deprecated: 0.172.0
// tags: transformations
//
builtin join : (<-tables: A, ?method: string, ?on: [string]) => stream[B] where A: Record, B: Record

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
// ## Metadata
// introduced: 0.40.0
// tags: transformations
//
builtin kaufmansAMA : (<-tables: stream[A], n: int, ?column: string) => stream[B] where A: Record, B: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin keep : (<-tables: stream[A], ?columns: [string], ?fn: (column: string) => bool) => stream[B]
    where
    A: Record,
    B: Record

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
// ## Metadata
// introduced: 0.13.0
// tags: transformations
//
builtin keyValues : (<-tables: stream[A], ?keyColumns: [string]) => stream[{C with _key: string, _value: B}]
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
// ## Metadata
// introduced: 0.13.0
// tags: transformations
//
builtin keys : (<-tables: stream[A], ?column: string) => stream[B] where A: Record, B: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations,selectors
//
builtin last : (<-tables: stream[A], ?column: string) => stream[A] where A: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
builtin limit : (<-tables: stream[A], n: int, ?offset: int) => stream[A]

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
// If the output record contains a different value for a group key column, the
// record is regrouped into the appropriate table.
// If the output record drops a group key column, that column is removed from
// the group key.
//
// #### Preserve columns
// `map()` drops any columns that are not mapped explictly by column label or
// implicitly using the `with` operator in the `fn` function.
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
// - mergeKey: _(Deprecated)_ Merge group keys of mapped records. Default is `false`.
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
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin map : (<-tables: stream[A], fn: (r: A) => B, ?mergeKey: bool) => stream[B]

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
builtin max : (<-tables: stream[A], ?column: string) => stream[A] where A: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations, aggregates
//
builtin mean : (<-tables: stream[A], ?column: string) => stream[B] where A: Record, B: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
builtin min : (<-tables: stream[A], ?column: string) => stream[A] where A: Record

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
// ## Metadata
// introduced: 0.36.0
// tags: transformtions, aggregates
//
builtin mode : (<-tables: stream[A], ?column: string) => stream[{C with _value: B}] where A: Record, C: Record

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
// ## Metadata
// introduced: 0.35.0
// tags: transformations
//
builtin movingAverage : (<-tables: stream[{B with _value: A}], n: int) => stream[{B with _value: float}]
    where
    A: Numeric

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
//   `quantile()` acts as an aggregate transformation and outputs the average of
//   non-null records with values that fall within the specified quantile.
// - **Selector**: When using the `exact_selector` method, `quantile()` acts as
//   a selector selector transformation and outputs the non-null record with the
//   value that represents the specified quantile.
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
//     - **exact_mean**: Aggregate method that takes the average of the two
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
// ## Metadata
// introduced: 0.24.0
// tags: transformations, aggregates, selectors
//
builtin quantile : (
        <-tables: stream[A],
        ?column: string,
        q: float,
        ?compression: float,
        ?method: string,
    ) => stream[A]
    where
    A: Record

// pivot collects unique values stored vertically (column-wise) and aligns them
// horizontally (row-wise) into logical sets.
//
// ### Output data
// The group key of the resulting table is the same as the input tables,
// excluding columns found in the `columnKey` and `valueColumn` parameters.
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
//   `rowKey`, `columnKey`, and `valueColumn` parameters is dropped.
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
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin pivot : (<-tables: stream[A], rowKey: [string], columnKey: [string], valueColumn: string) => stream[B]
    where
    A: Record,
    B: Record

// range filters rows based on time bounds.
//
// Input data must have a `_time` column of type time.
// Rows with a null value in the `_time` are filtered.
// `range()` adds a `_start` column with the value of `start` and a `_stop`
// column with the value of `stop`.
// `_start` and `_stop` columns are added to the group key.
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
// ## Metadata
// introduced: 0.7.0
// tags: transformations, filters
//
builtin range : (
        <-tables: stream[{A with _time: time}],
        start: B,
        ?stop: C,
    ) => stream[{A with _time: time, _start: time, _stop: time}]

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
// ## Metadata
// introduced: 0.23.0
// tags: transformations, aggregates
//
builtin reduce : (<-tables: stream[A], fn: (r: A, accumulator: B) => B, identity: B) => stream[C]
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
// ## Metadata
// introduced: 0.38.0
// tags: transformations
//
builtin relativeStrengthIndex : (<-tables: stream[A], n: int, ?columns: [string]) => stream[B]
    where
    A: Record,
    B: Record

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
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin rename : (<-tables: stream[A], ?fn: (column: string) => string, ?columns: B) => stream[C]
    where
    A: Record,
    B: Record,
    C: Record

// sample selects a subset of the rows from each input table.
//
// **Note:** `sample()` drops empty tables.
//
// ## Parameters
// - n: Sample every Nth element.
// - pos: Position offset from the start of results where sampling begins.
//   Default is -1 (random offset).
//
//   `pos` must be less than `n`. If pos is less than 0, a random offset is used.
//
// - column: Column to operate on.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Sample every other result
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> sample(n: 2, pos: 1)
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
builtin sample : (<-tables: stream[A], n: int, ?pos: int, ?column: string) => stream[A] where A: Record

// set assigns a static column value to each row in the input tables.
//
// `set()` may modify an existing column or add a new column.
// If the modified column is part of the group key, output tables are regrouped as needed.
// `set()` can only set string values.
//
// ## Parameters
// - key: Label of the column to modify or set.
// - value: String value to set.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Set a column to a specific string value
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> set(key: "host", value: "prod1")
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin set : (<-tables: stream[A], key: string, value: string) => stream[A] where A: Record

// tail limits each output table to the last `n` rows.
//
// `tail()` produces one output table for each input table.
// Each output table contains the last `n` records before the `offset`.
// If the input table has less than `offset + n` records, `tail()` outputs all
// records before the `offset`.
//
// ## Parameters
// - n: Maximum number of rows to output.
// - offset: Number of records to skip at the end of a table table before
//   limiting to `n`. Default is 0.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Output the last three rows in each input table
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> tail(n: 3)
// ```
//
// ### Output the last three rows before the last row in each input table
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> tail(n: 3, offset: 1)
// ```
//
// ## Metadata
// introduced: 0.39.0
// tags: transformations
//
builtin tail : (<-tables: stream[A], n: int, ?offset: int) => stream[A]

// timeShift adds a fixed duration to time columns.
//
// The output table schema is the same as the input table schema.
// `null` time values remain `null`.
//
// ## Parameters
// - duration: Amount of time to add to each time value. May be a negative duration.
// - columns: List of time columns to operate on. Default is `["_start", "_stop", "_time"]`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Shift timestamps forward in time
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> timeShift(duration: 12h)
// ```
//
// ### Shift timestamps backward in time
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> timeShift(duration: -12h)
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, date/time
//
builtin timeShift : (<-tables: stream[A], duration: duration, ?columns: [string]) => stream[A]

// skew returns the skew of non-null records in each input table as a float.
//
// ## Parameters
// - column: Column to operate on. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the skew of values
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> skew()
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, aggregates
//
builtin skew : (<-tables: stream[A], ?column: string) => stream[B] where A: Record, B: Record

// spread returns the difference between the minimum and maximum values in a
// specified column.
//
// ## Parameters
// - column: Column to operate on. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the spread of values
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> spread()
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, aggregates
//
builtin spread : (<-tables: stream[A], ?column: string) => stream[B] where A: Record, B: Record

// sort orders rows in each intput table based on values in specified columns.
//
// #### Output data
// One output table is produced for each input table.
// Output tables have the same schema as their corresponding input tables.
//
// #### Sorting with null values
// When `desc: false`, null values are last in the sort order.
// When `desc: true`, null values are first in the sort order.
//
// ## Parameters
// - columns: List of columns to sort by. Default is ["_value"].
//
//   Sort precedence is determined by list order (left to right).
//
// - desc: Sort results in descending order. Default is `false`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Sort values in ascending order
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> sort()
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin sort : (<-tables: stream[A], ?columns: [string], ?desc: bool) => stream[A] where A: Record

// stateTracking returns the cumulative count and duration of consecutive
// rows that match a predicate function that defines a state.
//
// To return the cumulative count of consecutive rows that match the predicate,
// include the `countColumn` parameter.
// To return the cumulative duration of consecutive rows that match the predicate,
// include the `durationColumn` parameter.
// Rows that do not match the predicate function `fn` return `-1` in the count
// and duration columns.
//
// ## Parameters
// - fn: Predicate function to determine state.
// - countColumn: Column to store state count in.
//
//   If not defined, `stateTracking()` does not return the state count.
//
// - durationColumn: Column to store state duration in.
//
//   If not defined, `stateTracking()` does not return the state duration.
//
// - durationUnit: Unit of time to report state duration in. Default is `1s`.
// - timeColumn: Column with time values used to calculate state duration.
//   Default is `_time`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return a cumulative state count
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.int()
// #         |> map(fn: (r) => ({r with state: if r._value > 5 then "crit" else "ok"}))
// #
// < data
// >     |> stateTracking(fn: (r) => r.state == "crit", countColumn: "count")
// ```
//
// ### Return a cumulative state duration in milliseconds
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.int()
// #         |> map(fn: (r) => ({r with state: if r._value > 5 then "crit" else "ok"}))
// #
// < data
// >     |> stateTracking(fn: (r) => r.state == "crit", durationColumn: "duration", durationUnit: 1ms)
// ```
//
// ### Return a cumulative state count and duration
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.int()
// #         |> map(fn: (r) => ({r with state: if r._value > 5 then "crit" else "ok"}))
// #
// < data
// >     |> stateTracking(fn: (r) => r.state == "crit", countColumn: "count", durationColumn: "duration")
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin stateTracking : (
        <-tables: stream[A],
        fn: (r: A) => bool,
        ?countColumn: string,
        ?durationColumn: string,
        ?durationUnit: duration,
        ?timeColumn: string,
    ) => stream[B]
    where
    A: Record,
    B: Record

// stddev returns the standard deviation of non-null values in a specified column.
//
// ## Parameters
// - column: Column to operate on. Default is `_value`.
// - mode: Standard deviation mode or type of standard deviation to calculate.
//   Default is `sample`.
//
//   **Availble modes:**
//
//   - **sample**: Calculate the sample standard deviation where the data is
//     considered part of a larger population.
//   - **population**: Calculate the population standard deviation where the
//     data is considered a population of its own.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the standard deviation of values in each table
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> stddev()
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, aggregates
//
builtin stddev : (<-tables: stream[A], ?column: string, ?mode: string) => stream[B] where A: Record, B: Record

// sum returns the sum of non-null values in a specified column.
//
// ## Parameters
// - column: Column to operate on. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the sum of values in each table
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> stddev()
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, aggregates
//
builtin sum : (<-tables: stream[A], ?column: string) => stream[B] where A: Record, B: Record

// tripleExponentialDerivative returns the triple exponential derivative (TRIX)
// values using `n` points.
//
// Triple exponential derivative, commonly referred to as “[TRIX](https://en.wikipedia.org/wiki/Trix_(technical_analysis)),”
// is a momentum indicator and oscillator. A triple exponential derivative uses
// the natural logarithm (log) of input data to calculate a triple exponential
// moving average over the period of time. The calculation prevents cycles
// shorter than the defined period from being considered by the indicator.
// `tripleExponentialDerivative()` uses the time between `n` points to define
// the period.
//
// Triple exponential derivative oscillates around a zero line.
// A positive momentum **oscillator** value indicates an overbought market;
// a negative value indicates an oversold market.
// A positive momentum **indicator** value indicates increasing momentum;
// a negative value indicates decreasing momentum.
//
// #### Triple exponential moving average rules
// - A triple exponential derivative is defined as:
//     - `TRIX[i] = ((EMA3[i] / EMA3[i - 1]) - 1) * 100`
//     - `EMA3 = EMA(EMA(EMA(data)))`
// - If there are not enough values to calculate a triple exponential derivative,
//   the output `_value` is `NaN`; all other columns are the same as the last
//   record of the input table.
// - The function behaves the same way as the `exponentialMovingAverage()` function:
//     - The function ignores `null` values.
//     - The function operates only on the `_value` column.
//
// ## Parameters
// - n: Number of points to use in the calculation.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate a two-point triple exponential derivative
// ```
// import "sampledata"
//
// sampledata.float()
//     |> tripleExponentialDerivative(n: 2)
// ```
//
// ## Metadata
// introduced: 0.40.0
// tags: transformations
//
builtin tripleExponentialDerivative : (<-tables: stream[{B with _value: A}], n: int) => stream[{B with _value: float}]
    where
    A: Numeric,
    B: Record

// union merges two or more input streams into a single output stream.
//
// The output schemas of `union()` is the union of all input schemas.
// `union()` does not preserve the sort order of the rows within tables.
// Use `sort()` if you need a specific sort order.
//
// ### Union vs join
// `union()` does not modify data in rows, but unions separate streams of tables
// into a single stream of tables and groups rows of data based on existing group keys.
// `join()` creates new rows based on common values in one or more specified columns.
// Output rows also contain the differing values from each of the joined streams.
//
// ## Parameters
// - tables: List of two or more streams of tables to union together.
//
// ## Examples
//
// ### Union two streams of tables with unique group keys
// ```
// import "generate"
//
// t1 =
//     generate.from(count: 4, fn: (n) => n + 1, start: 2022-01-01T00:00:00Z, stop: 2022-01-05T00:00:00Z)
//         |> set(key: "tag", value: "foo")
//         |> group(columns: ["tag"])
//
// t2 =
//     generate.from(count: 4, fn: (n) => n * (-1), start: 2022-01-01T00:00:00Z, stop: 2022-01-05T00:00:00Z)
//         |> set(key: "tag", value: "bar")
//         |> group(columns: ["tag"])
//
// > union(tables: [t1, t2])
// ```
//
// ### Union two streams of tables with empty group keys
// ```
// import "generate"
//
// t1 =
//     generate.from(count: 4, fn: (n) => n + 1, start: 2021-01-01T00:00:00Z, stop: 2021-01-05T00:00:00Z)
//         |> set(key: "tag", value: "foo")
//         |> group()
//
// t2 =
//     generate.from(count: 4, fn: (n) => n * (-1), start: 2021-01-01T00:00:00Z, stop: 2021-01-05T00:00:00Z)
//         |> set(key: "tag", value: "bar")
//         |> group()
//
// > union(tables: [t1, t2])
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
builtin union : (tables: [stream[A]]) => stream[A] where A: Record

// unique returns all records containing unique values in a specified column.
//
// Group keys, columns, and values are not modified.
// `unique()` drops empty tables.
//
// ## Parameters
// - column: Column to search for unique values. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return unique values from input tables
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> unique()
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
builtin unique : (<-tables: stream[A], ?column: string) => stream[A] where A: Record

// _window is a helper function for windowing data by time.
builtin _window : (
        <-tables: stream[A],
        every: duration,
        period: duration,
        offset: duration,
        location: {zone: string, offset: duration},
        timeColumn: string,
        startColumn: string,
        stopColumn: string,
        createEmpty: bool,
    ) => stream[B]
    where
    A: Record,
    B: Record

// window groups records using regular time intervals.
//
// The function calculates time windows and stores window bounds in the
// `_start` and `_stop` columns. `_start` and `_stop` values are assigned to
// rows based on the row's `_time` value.
//
// A single input row may be placed into zero or more output tables depending on
// the parameters passed into `window()`.
//
// #### Window by calendar months and years
// `every`, `period`, and `offset` parameters support all valid duration units,
// including calendar months (`1mo`) and years (`1y`).
//
// #### Window by week
// When windowing by week (`1w`), weeks are determined using the Unix epoch
// (1970-01-01T00:00:00Z UTC). The Unix epoch was on a Thursday, so all
// calculated weeks begin on Thursday.
//
// ## Parameters
// - every: Duration of time between windows.
// - period: Duration of windows. Default is the `every` value.
//
//   `period` can be negative, indicating the start and stop boundaries are reversed.
//
// - offset: Duration to shift the window boundaries by. Defualt is `0s`.
//
//   `offset` can be negative, indicating that the offset goes backwards in time.
//
// - location: Location used to determine timezone. Default is the `location` option.
// - timeColumn: Column that contains time values. Default is `_time`.
// - startColumn: Column to store the window start time in. Default is `_start`.
// - stopColumn: Column to store the window stop time in. Default is `_stop`.
// - createEmpty: Create empty tables for empty window. Default is `false`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Window data into 30 second intervals
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.int()
// #         |> range(start: sampledata.start, stop: sampledata.stop)
// #
// < data
// >     |> window(every: 30s)
// ```
//
// ### Window every 20 seconds covering 40 second periods
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.int()
// #         |> range(start: sampledata.start, stop: sampledata.stop)
// #
// < data
// >     |> window(every: 20s, period: 40s)
// ```
//
// ### Window by calendar month
// ```
// # import "generate"
// #
// # timeRange = {start: 2021-01-01T00:00:00Z, stop: 2021-04-01T00:00:00Z}
// #
// # data =
// #     generate.from(count: 6, fn: (n) => n + n, start: timeRange.start, stop: timeRange.stop)
// #         |> range(start: timeRange.start, stop: timeRange.stop)
// #
// < data
// >     |> window(every: 1mo)
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
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

// yield delivers input data as a result of the query.
//
// A query may have multiple yields, each identified by unique name specified in
// the `name` parameter.
//
// **Note:** `yield()` is implicit for queries that output a single stream of
// tables and is only necessary when yielding multiple results from a query.
//
// ## Parameters
// - name: Unique name for the yielded results. Default is `_results`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Yield multiple results from a query
// ```
// import "sampledata"
//
// sampledata.int()
//     |> yield(name: "unmodified")
//     |> map(fn: (r) => ({r with _value: r._value * r._value}))
//     |> yield(name: "squared")
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: outputs
//
builtin yield : (<-tables: stream[A], ?name: string) => stream[A] where A: Record

// tableFind extracts the first table in a stream with group key values that
// match a specified predicate.
//
// ## Parameters
// - fn: Predicate function to evaluate input table group keys.
//
//   `tableFind()` returns the first table that resolves as `true`.
//   The predicate function requires a `key` argument that represents each input
//   table's group key as a record.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Extract a table from a stream of tables
// ```no_run
// import "sampledata"
//
// t = sampledata.int() |> tableFind(fn: (key) => key.tag == "t2")
//
// // t represents the first table in a stream whose group key
// // contains "tag" with a value of "t2".
// ```
//
// ## Metadata
// introduced: 0.29.0
// tags: dynamic queries
//
builtin tableFind : (<-tables: stream[A], fn: (key: B) => bool) => stream[A] where A: Record, B: Record

// getColumn extracts a specified column from a table as an array.
//
// If the specified column is not present in the table, the function returns an error.
//
// ## Parameters
// - column: Column to extract.
// - table: Input table. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Extract a column from a table
// ```no_run
// import "sampledata"
//
// sampledata.int()
//     |> tableFind(fn: (key) => key.tag == "t1")
//     |> getColumn(column: "_value")
//
// // Returns [-2, 10, 7, 17, 15, 4]
// ```
//
// ## Metadata
// introduced: 0.29.0
// tags: dynamic queries
//
builtin getColumn : (<-table: stream[A], column: string) => [B] where A: Record

// getRecord extracts a row at a specified index from a table as a record.
//
// If the specified index is out of bounds, the function returns an error.
//
// ## Parameters
// - idx: Index of the record to extract.
// - table: Input table. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Extract the first row from a table as a record
// ```no_run
// import "sampledata"
//
// sampledata.int()
//   |> tableFind(fn: (key) => key.tag == "t1")
//   |> getRecord(idx: 0)
//
// // Returns {_time: 2021-01-01T00:00:00.000000000Z, _value: -2, tag: t1}
// ```
//
// ## Metadata
// introduced: 0.29.0
// tags: dynamic queries
//
builtin getRecord : (<-table: stream[A], idx: int) => A where A: Record

// findColumn returns an array of values in a specified column from the first
// table in a stream of tables that matches the specified predicate function.
//
// The function returns an empty array if no table is found or if the column
// label is not present in the set of columns.
//
// ## Parameters
// - column: Column to extract.
// - fn: Predicate function to evaluate input table group keys.
//
//   `findColumn()` uses the first table that resolves as `true`.
//   The predicate function requires a `key` argument that represents each input
//   table's group key as a record.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Extract a column as an array
// ```no_run
// import "sampledata"
//
// sampledata.int()
//     |> findColumn(fn: (key) => key.tag == "t1", column: "_value")
//
// // Returns [-2, 10, 7, 17, 15, 4]
// ```
//
// ## Metadata
// introduced: 0.68.0
// tags: dynamic queries
//
builtin findColumn : (<-tables: stream[A], fn: (key: B) => bool, column: string) => [C] where A: Record, B: Record

// findRecord returns a row at a specified index as a record from the first
// table in a stream of tables that matches the specified predicate function.
//
// The function returns an empty record if no table is found or if the index is
// out of bounds.
//
// ## Parameters
// - idx: Index of the record to extract.
// - fn: Predicate function to evaluate input table group keys.
//
//   `findColumn()` uses the first table that resolves as `true`.
//   The predicate function requires a `key` argument that represents each input
//   table's group key as a record.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Extract a row as a record
// ```no_run
// import "sampledata"
//
// sampledata.int()
//     |> findRecord(fn: (key) => key.tag == "t1", idx: 0)
//
// // Returns {_time: 2021-01-01T00:00:00.000000000Z, _value: -2, tag: t1}
// ```
//
// ## Metadata
// introduced: 0.68.0
// tags: dynamic queries
//
builtin findRecord : (<-tables: stream[A], fn: (key: B) => bool, idx: int) => A where A: Record, B: Record

// bool converts a value to a boolean type.
//
// ## Parameters
// - v: Value to convert.
//
// ## Examples
//
// ### Convert strings to booleans
// ```no_run
// bool(v: "true") // Returns true
// bool(v: "false") // Returns false
// ```
//
// ### Convert numeric values to booleans
// ```no_run
// bool(v: 1.0) // Returns true
// bool(v: 0.0) // Returns false
// bool(v: 1) // Returns true
// bool(v: 0) // Returns false
// bool(v: uint(v: 1)) // Returns true
// bool(v: uint(v: 0)) // Returns false
// ```
//
// ### Convert all values in a column to booleans
// If converting the `_value` column to boolean types, use `toBool()`.
// If converting columns other than `_value`, use `map()` to iterate over each
// row and `bool()` to covert a column value to a boolean type.
//
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.numericBool()
// #         |> rename(columns: {_value: "powerOn"})
// #
// < data
// >     |> map(fn: (r) => ({r with powerOn: bool(v: r.powerOn)}))
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: type-conversions
//
builtin bool : (v: A) => bool

// bytes converts a string value to a bytes type.
//
// ## Parameters
// - v: Value to convert.
//
// ## Examples
//
// ### Convert a string to bytes
// ```no_run
// bytes(v: "Example string") // Returns 0x4578616d706c6520737472696e67
// ```
//
// ## Metadata
// introduced: 0.40.0
// tags: type-conversions
//
builtin bytes : (v: A) => bytes

// duration converts a value to a duration type.
//
// `duration()` treats integers and unsigned integers as nanoseconds.
// For a string to be converted to a duration type, the string must use
// duration literal representation.
//
// ## Parameters
// - v: Value to convert.
//
// ## Examples
//
// ### Convert a string to a duration
// ```no_run
// duration(v: "1h20m") // Returns 1h20m
// ```
//
// ### Convert numeric types to durations
// ```no_run
// duration(v: 4800000000000) // Returns 1h20m
// duration(v: uint(v: 9600000000000)) // Returns 2h40m
// ```
//
// ### Convert values in a column to durations
// Flux does not support duration column types.
// To store durations in a column, convert duration types to strings.
//
// ```
// # import "array"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2022-01-01T05:00:00Z, tag: "t1", _value: -27000000},
// #         {_time: 2022-01-01T09:00:10Z, tag: "t1", _value: 12000000},
// #         {_time: 2022-01-01T11:00:20Z, tag: "t1", _value: 78000000},
// #         {_time: 2022-01-01T16:00:30Z, tag: "t1", _value: 17000000},
// #         {_time: 2022-01-01T19:00:40Z, tag: "t1", _value: 15000000},
// #         {_time: 2022-01-01T20:00:50Z, tag: "t1", _value: -42000000},
// #     ],
// # )
// #
// < data
// >     |> map(fn: (r) => ({r with _value: string(v: duration(v: r._value))}))
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: type-conversions
//
builtin duration : (v: A) => duration

// float converts a value to a float type.
//
// ## Parameters
// - v: Value to convert.
//
// ## Examples
//
// ### Convert a string to a float
// ```no_run
// float(v: "3.14") // Returns 3.14
// ```
//
// ### Convert a scientific notation string to a float
// ```no_run
// float(v: "1.23e+20") // Returns 1.23e+20 (float)
// ```
//
// ### Convert an integer to a float
// ```no_run
// float(v: "10") // Returns 10.0
// ```
//
// ### Convert all values in a column to floats
// If converting the `_value` column to float types, use `toFloat()`.
// If converting columns other than `_value`, use `map()` to iterate over each
// row and `float()` to covert a column value to a float type.
//
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.int()
// #         |> rename(columns: {_value: "exampleCol"})
// #
// < data
// >     |> map(fn: (r) => ({r with exampleCol: float(v: r.exampleCol)}))
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: type-conversions
//
builtin float : (v: A) => float

// int converts a value to an integer type.
//
// `int()` behavior depends on the input data type:
//
// | Input type | Returned value                                  |
// | :--------- | :---------------------------------------------- |
// | string     | Integer equivalent of the numeric string        |
// | bool       | 1 (true) or 0 (false)                           |
// | duration   | Number of nanoseconds in the specified duration |
// | time       | Equivalent nanosecond epoch timestamp           |
// | float      | Value truncated at the decimal                  |
// | uint       | Integer equivalent of the unsigned integer      |
//
// ## Parameters
// - v: Value to convert.
//
// ## Examples
//
// ### Convert basic types to integers
// ```no_run
// int(v: 10.12) // Returns 10
// int(v: "3") // Returns 3
// int(v: true) // Returns 1
// int(v: 1m) // Returns 160000000000
// int(v: 2022-01-01T00:00:00Z) // Returns 1640995200000000000
// ```
//
// ### Convert all values in a column to integers
// If converting the `_value` column to integer types, use `toInt()`.
// If converting columns other than `_value`, use `map()` to iterate over each
// row and `int()` to covert a column value to a integer type.
//
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.float()
// #         |> rename(columns: {_value: "exampleCol"})
// #
// < data
// >     |> map(fn: (r) => ({r with exampleCol: int(v: r.exampleCol)}))
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: type-conversions
//
builtin int : (v: A) => int

// string converts a value to a string type.
//
// ## Parameters
// - v: Value to convert.
//
// ## Examples
//
// ### Convert basic types to strings
// ```no_run
// string(v: true) // Returns "true"
// string(v: 1m) // Returns "1m"
// string(v: 2021-01-01T00:00:00Z) // Returns "2021-01-01T00:00:00Z"
// string(v: 10.12) // Returns "10.12"
// ```
//
// ### Convert all values in a column to strings
// If converting the `_value` column to string types, use `toString()`.
// If converting columns other than `_value`, use `map()` to iterate over each
// row and `string()` to covert a column value to a string type.
//
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.float()
// #         |> rename(columns: {_value: "exampleCol"})
// #
// < data
// >     |> map(fn: (r) => ({r with exampleCol: string(v: r.exampleCol)}))
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: type-conversions
//
builtin string : (v: A) => string

// time converts a value to a time type.
//
// ## Parameters
// - v: Value to convert.
//
//   Strings must be valid [RFC3339 timestamps](https://docs.influxdata.com/influxdb/cloud/reference/glossary/#rfc3339-timestamp).
//   Integer and unsigned integer values are parsed as nanosecond epoch timestamps.
//
// ## Examples
//
// ### Convert a string to a time value
// ```no_run
// time(v: "2021-01-01T00:00:00Z") // Returns 2021-01-01T00:00:00Z (time)
// ```
//
// ### Convert an integer to a time value
// ```no_run
// time(v: 1640995200000000000) // Returns 2022-01-01T00:00:00Z
// ```
//
// ### Convert all values in a column to time
// If converting the `_value` column to time types, use `toTime()`.
// If converting columns other than `_value`, use `map()` to iterate over each
// row and `time()` to covert a column value to a time type.
//
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.int()
// #         |> map(fn: (r) => ({ r with exampleCol: r._value * 1000000000 }))
// #
// < data
// >     |> map(fn: (r) => ({r with exampleCol: time(v: r.exampleCol)}))
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: type-conversions
//
builtin time : (v: A) => time

// uint converts a value to an unsigned integer type.
//
// `uint()` behavior depends on the input data type:
//
// | Input type | Returned value                                                  |
// | :--------- | :-------------------------------------------------------------- |
// | bool       | 1 (true) or 0 (false)                                           |
// | duration   | Number of nanoseconds in the specified duration                 |
// | float      | UInteger equivalent of the float value truncated at the decimal |
// | int        | UInteger equivalent of the integer                              |
// | string     | UInteger equivalent of the numeric string                       |
// | time       | Equivalent nanosecond epoch timestamp                           |
//
// ## Parameters
// - v: Value to convert.
//
// ## Examples
//
// ### Convert basic types to unsigned integers
// ```no_run
// uint(v: "3") // Returns 3
// uint(v: 1m) // Returns 160000000000
// uint(v: 2022-01-01T00:00:00Z) // Returns 1640995200000000000
// uint(v: 10.12) // Returns 10
// uint(v: -100) // Returns 18446744073709551516
// ```
//
// ### Convert all values in a column to unsigned integers
// If converting the `_value` column to uint types, use `toUInt()`.
// If converting columns other than `_value`, use `map()` to iterate over each
// row and `uint()` to covert a column value to a uint type.
//
// ```
// # import "sampledata"
// #
// # data =
// #     sampledata.int()
// #         |> rename(columns: {_value: "exampleCol"})
// #
// < data
// >     |> map(fn: (r) => ({r with exampleCol: uint(v: r.exampleCol)}))
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: type-conversions
//
builtin uint : (v: A) => uint

// display returns the Flux literal representation of any value as a string.
//
// Basic types are converted directly to a string.
// Bytes types are represented as a string of lowercase hexadecimal characters prefixed with `0x`.
// Composite types (arrays, dictionaries, and records) are represented in a syntax similar
// to their equivalent Flux literal representation.
//
// Note the following about the resulting string representation:
// - It cannot always be parsed back into the original value.
// - It may span multiple lines.
// - It may change between Flux versions.
//
//`display()` differs from `string()` in that `display()` recursively converts values inside
// composite types to strings. `string()` does not operate on composite types.
//
// ## Parameters
// - v: Value to convert for display.
//
// ## Examples
//
// ### Display a value as part of a table
//
// Use `array.from()` and `display()` to quickly observe any value.
//
// ```no_run
// import "array"
//
// array.from(
//     rows: [
//         {dict: display(v: ["a": 1, "b": 2]),record: display(v: {x: 1, y: 2}),array: display(v: [5, 6, 7])},
//     ]
// > )
// ```
//
// ### Display a record
//
// ```no_run
// x = {a: 1, b: 2, c: 3}
// display(v: x)
//
// // Returns {a: 1, b: 2, c: 3}
// ```
//
// ### Display an array
//
// ```no_run
// x = [1, 2, 3]
// display(v: x)
//
// // Returns [1, 2, 3]
// ```
//
// ### Display a dictionary
//
// ```no_run
// x = ["a": 1, "b": 2, "c": 3]
// display(v: x)
//
// // Returns [a: 1, b: 2, c: 3]
// ```
//
// ### Display bytes
//
// ```no_run
// x = bytes(v:"abc")
// display(v: x)
//
// // Returns 0x616263
// ```
//
// ### Display a composite value
//
// ```no_run
// x = {
//     bytes: bytes(v: "abc"),
//     string: "str",
//     array: [1,2,3],
//     dict: ["a": 1, "b": 2, "c": 3],
// }
//
// display(v: x)
//
// // Returns
// // {
// //    array: [1, 2, 3],
// //    bytes: 0x616263,
// //    dict: [a: 1, b: 2, c: 3],
// //    string: str
// // }
// ```
//
// ## Metadata
// introduced: 0.154.0
//
builtin display : (v: A) => string

// contains tests if an array contains a specified value and returns `true` or `false`.
//
// ## Parameters
// - value: Value to search for.
// - set: Array to search.
//
// ## Examples
//
// ### Filter on a set of specific fields
// ```
// # import "sampledata"
// #
// # data = sampledata.int()
// #     |> map(fn: (r) => ({r with _measurement: "m", _field: if r._value < 6 then "f1" else if r._value < 11 then "f2" else "f3"}))
// #     |> group(columns: ["tag", "_field"])
// #
// fields = ["f1", "f2"]
//
// < data
// >     |> filter(fn: (r) => contains(value: r._field, set: fields))
// ```
//
// ## Metadata
// introduced: 0.19.0
//
builtin contains : (value: A, set: [A]) => bool where A: Nullable

// inf represents an infinte float value.
builtin inf : duration

// length returns the number of elements in an array.
//
// ## Parameters
// - arr: Array to evaluate. Default is the piped-forward array (`<-`).
//
// ## Examples
//
// ### Return the length of an array
// ```no_run
// people = ["John", "Jane", "Abed"]
//
// people |> length()
// // Returns 3
// ```
//
// ## Metadata
// introduced: 0.7.0
//
builtin length : (<-arr: [A]) => int

// linearBins generates a list of linearly separated float values.
//
// Use `linearBins()` to generate bin bounds for `histogram()`.
//
// ## Parameters
// - start: First value to return in the list.
// - width: Distance between subsequent values.
// - count: Number of values to return.
// - infinity: Include an infinite value at the end of the list. Default is `true`.
//
// ## Examples
//
// ### Generate a list of linearly increasing values
// ```no_run
// linearBins(start: 0.0, width: 10.0, count: 10)
// // Returns [0, 10, 20, 30, 40, 50, 60, 70, 80, 90, +Inf]
// ```
//
// ## Metadata
// introduced: 0.19.0
//
builtin linearBins : (start: float, width: float, count: int, ?infinity: bool) => [float]

// logarithmicBins generates a list of exponentially separated float values.
//
// Use `linearBins()` to generate bin bounds for `histogram()`.
//
// ## Parameters
// - start: First value to return in the list.
// - factor: Multiplier to apply to subsequent values.
// - count: Number of values to return.
// - infinity: Include an infinite value at the end of the list. Default is `true`.
//
// ## Examples
//
// ### Generate a list of exponentially increasing values
// ```no_run
// logarithmicBins(start: 1.0, factor: 2.0, count: 10, infinity: true)
// // Returns [1, 2, 4, 8, 16, 32, 64, 128, 256, 512, +Inf]
// ```
//
// ## Metadata
// introduced: 0.19.0
//
builtin logarithmicBins : (start: float, factor: float, count: int, ?infinity: bool) => [float]

// timeWeightedAvg returns the time-weighted average of non-null values in
// `_value` column as a float for each input table.
//
// Time is weighted using the linearly interpolated integral of values in the table.
//
// ## Parameters
// - unit: Unit of time to use to compute the time-weighted average.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate the time-weighted average of values
// ```
// # import "sampledata"
// #
// # data = sampledata.int(includeNull: true)
// #     |> range(start: sampledata.start, stop: sampledata.stop)
// #     |> fill(usePrevious: true)
// #     |> unique()
// #
// < data
// >     |> timeWeightedAvg(unit: 1s)
// ```
//
// ## Metadata
// introduced: 0.83.0
// tags: transformations, aggregates
//
timeWeightedAvg = (tables=<-, unit) =>
    tables
        |> integral(unit: unit, interpolate: "linear")
        |> map(
            fn: (r) =>
                ({r with _value: r._value * float(v: uint(v: unit)) / float(v: int(v: r._stop) - int(v: r._start))}),
        )

// cov computes the covariance between two streams of tables.
//
// ## Parameters
// - x: First input stream.
// - y: Second input stream.
// - on: List of columns to join on.
// - pearsonr: Normalize results to the Pearson R coefficient. Default is `false`.
//
// ## Examples
//
// ### Return the covariance between two streams of tables
// ```
// import "generate"
//
// stream1 = generate.from(count: 5, fn: (n) => n * n, start: 2021-01-01T00:00:00Z, stop: 2021-01-01T00:01:00Z)
//     |> toFloat()
//
// stream2 = generate.from(count: 5, fn: (n) => n * n * n / 2, start: 2021-01-01T00:00:00Z, stop: 2021-01-01T00:01:00Z)
//     |> toFloat()
//
// > cov(x: stream1, y: stream2, on: ["_time"])
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, aggregates
//
cov = (x, y, on, pearsonr=false) =>
    join(tables: {x: x, y: y}, on: on)
        |> covariance(pearsonr: pearsonr, columns: ["_value_x", "_value_y"])

// pearsonr returns the covariance of two streams of tables normalized to the
// Pearson R coefficient.
//
// ## Parameters
// - x: First input stream.
// - y: Second input stream.
// - on: List of columns to join on.
//
// ## Examples
//
// ### Return the covariance between two streams of tables
// ```
// import "generate"
//
// stream1 = generate.from(count: 5, fn: (n) => n * n, start: 2021-01-01T00:00:00Z, stop: 2021-01-01T00:01:00Z)
//     |> toFloat()
//
// stream2 = generate.from(count: 5, fn: (n) => n * n * n / 2, start: 2021-01-01T00:00:00Z, stop: 2021-01-01T00:01:00Z)
//     |> toFloat()
//
// > pearsonr(x: stream1, y: stream2, on: ["_time"])
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, aggregates
//
pearsonr = (x, y, on) => cov(x: x, y: y, on: on, pearsonr: true)

// _fillEmpty is a helper function that creates and fills empty tables.
_fillEmpty = (tables=<-, createEmpty) =>
    if createEmpty then
        tables
            |> table.fill()
    else
        tables

// aggregateWindow downsamples data by grouping data into fixed windows of time
// and applying an aggregate or selector function to each window.
//
// All columns not in the group key other than the specified `column` are dropped
// from output tables. This includes `_time`. `aggregateWindow()` uses the
// `timeSrc` and `timeDst` parameters to assign a time to the aggregate value.
//
// `aggregateWindow()` requires `_start` and `_stop` columns in input data.
// Use `range()` to assign `_start` and `_stop` values.
//
// #### Downsample by calendar months and years
// `every`, `period`, and `offset` parameters support all valid duration units,
// including calendar months (`1mo`) and years (`1y`).
//
// #### Downsample by week
// When windowing by week (`1w`), weeks are determined using the Unix epoch
// (1970-01-01T00:00:00Z UTC). The Unix epoch was on a Thursday, so all
// calculated weeks begin on Thursday.
//
// ## Parameters
// - every: Duration of time between windows.
// - period: Duration of windows. Default is the `every` value.
//
//   `period` can be negative, indicating the start and stop boundaries are reversed.
//
// - offset: Duration to shift the window boundaries by. Defualt is `0s`.
//
//   `offset` can be negative, indicating that the offset goes backwards in time.
//
// - fn: Aggreate or selector function to apply to each time window.
// - location: Location used to determine timezone. Default is the `location` option.
// - column: Column to operate on.
// - timeSrc: Column to use as the source of the new time value for aggregate values.
//   Default is `_stop`.
// - timeDst: Column to store time values for aggregate values in.
//   Default is `_time`.
// - createEmpty: Create empty tables for empty window. Default is `false`.
//
//   **Note:** When using `createEmpty: true`, aggregate functions return empty
//   tables, but selector functions do not. By design, selectors drop empty tables.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Use an aggregate function with default parameters
// ```
// # import "sampledata"
// #
// # data = sampledata.float()
// #     |> range(start: sampledata.start, stop: sampledata.stop)
// #
// < data
// >     |> aggregateWindow(every: 20s, fn: mean)
// ```
//
// ### Specify parameters of the aggregate function
// To use functions that don’t provide defaults for required parameters with
// `aggregateWindow()`, define an anonymous function with `column` and `tables`
// parameters that pipes-forward tables into the aggregate or selector function
// with all required parameters defined:
//
// ```
// # import "sampledata"
// #
// # data = sampledata.float()
// #     |> range(start: sampledata.start, stop: sampledata.stop)
// #
// < data
//     |> aggregateWindow(
//         column: "_value",
//         every: 20s,
//         fn: (column, tables=<-) => tables |> quantile(q: 0.99, column: column),
// >     )
// ```
//
// ### Downsample by calendar month
// ```
// # import "sampledata"
// #
// # data = sampledata.float()
// #     |> range(start: sampledata.start, stop: sampledata.stop)
// #
// < data
// >     |> aggregateWindow(every: 1mo, fn: mean)
// ```
//
// ### Downsample by calendar week starting on Monday
//
// Flux increments weeks from the Unix epoch, which was a Thursday.
// Because of this, by default, all `1w` windows begin on Thursday.
// Use the `offset` parameter to shift the start of weekly windows to the
// desired day of the week.
//
// | Week start | Offset |
// | :--------- | :----: |
// | Monday     |  -3d   |
// | Tuesday    |  -2d   |
// | Wednesday  |  -1d   |
// | Thursday   |   0d   |
// | Friday     |   1d   |
// | Saturday   |   2d   |
// | Sunday     |   3d   |
//
// ```js
// # import "array"
// #
// # data =
// #     array.from(
// #         rows: [
// #             {_time: 2022-01-01T00:00:00Z, tag: "t1", _value: 2.0},
// #             {_time: 2022-01-03T00:00:00Z, tag: "t1", _value: 2.2},
// #             {_time: 2022-01-06T00:00:00Z, tag: "t1", _value: 4.1},
// #             {_time: 2022-01-09T00:00:00Z, tag: "t1", _value: 3.8},
// #             {_time: 2022-01-11T00:00:00Z, tag: "t1", _value: 1.7},
// #             {_time: 2022-01-12T00:00:00Z, tag: "t1", _value: 2.1},
// #             {_time: 2022-01-15T00:00:00Z, tag: "t1", _value: 3.8},
// #             {_time: 2022-01-16T00:00:00Z, tag: "t1", _value: 4.2},
// #             {_time: 2022-01-20T00:00:00Z, tag: "t1", _value: 5.0},
// #             {_time: 2022-01-24T00:00:00Z, tag: "t1", _value: 5.8},
// #             {_time: 2022-01-28T00:00:00Z, tag: "t1", _value: 3.9},
// #         ],
// #     )
// #         |> range(start: 2022-01-01T00:00:00Z, stop: 2022-01-31T23:59:59Z)
// #         |> group(columns: ["tag"])
// #
// < data
// >     |> aggregateWindow(every: 1w, offset: -3d, fn: mean)
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, aggregates, selectors
//
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

// increase returns the cumulative sum of non-negative differences between subsequent values.
//
// The primary use case for `increase()` is tracking changes in counter values
// which may wrap overtime when they hit a threshold or are reset. In the case
// of a wrap/reset, `increase()` assumes that the absolute delta between two
// points is at least their non-negative difference.
//
// ## Parameters
// - columns: List of columns to operate on. Default is `["_value"]`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Normalize resets in counter metrics
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> increase()
// ```
//
// ## Metadata
// introduced: 0.71.0
// tags: transformations
//
increase = (tables=<-, columns=["_value"]) =>
    tables
        |> difference(nonNegative: true, columns: columns, keepFirst: true, initialZero: true)
        |> cumulativeSum(columns: columns)

// median returns the median `_value` of an input table or all non-null records
// in the input table with values that fall within the 0.5 quantile (50th percentile).
//
// ### Function behavior
// `median()` acts as an aggregate or selector transformation depending on the
// specified `method`.
//
// - **Aggregate**: When using the `estimate_tdigest` or `exact_mean` methods,
//   `median()` acts as an aggregate transformation and outputs the average of
//   non-null records with values that fall within the 0.5 quantile (50th percentile).
// - **Selector**: When using the `exact_selector` method, `meidan()` acts as
//   a selector selector transformation and outputs the non-null record with the
//   value that represents the 0.5 quantile (50th percentile).
//
// ## Parameters
// - column: Column to use to compute the median. Default is `_value`.
// - method: Computation method. Default is `estimate_tdigest`.
//
//     **Avaialable methods**:
//
//     - **estimate_tdigest**: Aggregate method that uses a
//       [t-digest data structure](https://github.com/tdunning/t-digest) to
//       compute an accurate median estimate on large data sources.
//     - **exact_mean**: Aggregate method that takes the average of the two
//       points closest to the median value.
//     - **exact_selector**: Selector method that returns the row with the value
//       for which at least 50% of points are less than.
//
// - compression: Number of centroids to use when compressing the dataset.
//   Default is `0.0`.
//
//   A larger number produces a more accurate result at the cost of increased
//   memory requirements.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Use median as an aggregate transformation
// ```
// import "sampledata"
//
// < sampledata.float()
// >     |> median()
// ```
//
// ### Use median as a selector transformation
// ```
// import "sampledata"
//
// < sampledata.float()
// >     |> median(method: "exact_selector")
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, aggregates, selectors
//
median = (method="estimate_tdigest", compression=0.0, column="_value", tables=<-) =>
    tables
        |> quantile(q: 0.5, method: method, compression: compression, column: column)

// stateCount returns the number of consecutive rows in a given state.
//
// The state is defined by the `fn` predicate function. For each consecutive
// record that evaluates to `true`, the state count is incremented. When a record
// evaluates to `false`, the value is set to `-1` and the state count is reset.
// If the record generates an error during evaluation, the point is discarded,
// and does not affect the state count.
// The state count is added as an additional column to each record.
//
// ## Parameters
// - fn: Predicate function that identifies the state of a record.
// - column: Column to store the state count in. Default is `stateCount`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Count the number rows in a specific state
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> stateCount(fn: (r) => r._value < 10)
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
stateCount = (fn, column="stateCount", tables=<-) =>
    tables
        |> stateTracking(countColumn: column, fn: fn)

// stateDuration returns the cumulative duration of a given state.
//
// The state is defined by the `fn` predicate function. For each consecutive
// record that evaluates to `true`, the state duration is incremented by the
// duration of time between records using the specified `unit`. When a record
// evaluates to `false`, the value is set to `-1` and the state duration is reset.
// If the record generates an error during evaluation, the point is discarded,
// and does not affect the state duration.
//
// The state duration is added as an additional column to each record.
// The duration is represented as an integer in the units specified.
//
// **Note:** As the first point in the given state has no previous point, its
// state duration will be 0.
//
// ## Parameters
// - fn: Predicate function that identifies the state of a record.
// - column: Column to store the state duration in. Default is `stateDuration`.
// - timeColumn: Time column to use to calculate elapsed time between rows.
//   Default is `_time`.
// - unit: Unit of time to use to increment state duration. Default is `1s` (seconds).
//
//   **Example units:**
//   - 1ns (nanoseconds)
//   - 1us (microseconds)
//   - 1ms (milliseconds)
//   - 1s (seconds)
//   - 1m (minutes)
//   - 1h (hours)
//   - 1d (days)
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the time spent in a specified state
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> stateDuration(fn: (r) => r._value < 15)
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//
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

// top sorts each input table by specified columns and keeps the top `n` records
// in each table.
//
// **Note:** `top()` drops empty tables.
//
// ## Parameters
// - n: Number of rows to return from each input table.
// - columns: List of columns to sort by. Default is `["_value"]`.
//
//   Sort precedence is determined by list order (left to right).
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return rows with the three highest values in each input table
// ```
// import "sampledata"
//
// < sampledata.int()
// >    |> top(n: 3)
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
top = (n, columns=["_value"], tables=<-) =>
    tables
        |> _sortLimit(n: n, columns: columns, desc: true)

// bottom sorts each input table by specified columns and keeps the bottom `n`
// records in each table.
//
// **Note:** `bottom()` drops empty tables.
//
// ## Parameters
// - n: Number of rows to return from each input table.
// - columns: List of columns to sort by. Default is `["_value"]`.
//
//   Sort precedence is determined by list order (left to right).
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return rows with the two lowest values in each input table
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> bottom(n:2)
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
bottom = (n, columns=["_value"], tables=<-) =>
    tables
        |> _sortLimit(n: n, columns: columns, desc: false)

// _highestOrLowest is a helper function that reduces all groups into a single
// group by specific tags and a `reducer` function.
// It then selects the highest or lowest records based on the `column` and the
// `_sortLimit` function. The default `reducer` assumes no reducing needs to be
// performed.
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

// highestMax selects the record with the highest value in the specified `column`
// from each input table and returns the highest `n` records.
//
// **Note:** `highestMax()` drops empty tables.
//
// ## Parameters
// - n: Number of records to return.
// - column: Column to evaluate. Default is `_value`.
// - groupColumns: List of columns to group by. Default is `[]`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the highest two values from a stream of tables
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> highestMax(n: 2, groupColumns: ["tag"])
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
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

// highestAverage calculates the average of each input table and returns the
// highest `n` averages.
//
// **Note:** `highestAverage()` drops empty tables.
//
// ## Parameters
// - n: Number of records to return.
// - column: Column to evaluate. Default is `_value`.
// - groupColumns: List of columns to group by. Default is `[]`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the highest table average from a stream of tables
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> highestAverage(n: 1, groupColumns: ["tag"])
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
highestAverage = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
            n: n,
            column: column,
            groupColumns: groupColumns,
            reducer: (tables=<-) => tables |> mean(column: column),
            _sortLimit: top,
        )

// highestCurrent selects the last record from each input table and returns the
// highest `n` records.
//
// **Note:** `highestCurrent()` drops empty tables.
//
// ## Parameters
// - n: Number of records to return.
// - column: Column to evaluate. Default is `_value`.
// - groupColumns: List of columns to group by. Default is `[]`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the highest current value from a stream of tables
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> highestCurrent(n: 1, groupColumns: ["tag"])
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
highestCurrent = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
            n: n,
            column: column,
            groupColumns: groupColumns,
            reducer: (tables=<-) => tables |> last(column: column),
            _sortLimit: top,
        )

// lowestMin selects the record with the lowest value in the specified `column`
// from each input table and returns the bottom `n` records.
//
// **Note:** `lowestMin()` drops empty tables.
//
// ## Parameters
// - n: Number of records to return.
// - column: Column to evaluate. Default is `_value`.
// - groupColumns: List of columns to group by. Default is `[]`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the lowest two values from a stream of tables
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> lowestMin(n: 2, groupColumns: ["tag"])
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
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

// lowestAverage calculates the average of each input table and returns the lowest
// `n` averages.
//
// **Note:** `lowestAverage()` drops empty tables.
//
// ## Parameters
// - n: Number of records to return.
// - column: Column to evaluate. Default is `_value`.
// - groupColumns: List of columns to group by. Default is `[]`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the lowest table average from a stream of tables
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> lowestAverage(n: 1, groupColumns: ["tag"])
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
lowestAverage = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
            n: n,
            column: column,
            groupColumns: groupColumns,
            reducer: (tables=<-) => tables |> mean(column: column),
            _sortLimit: bottom,
        )

// lowestCurrent selects the last record from each input table and returns the
// lowest `n` records.
//
// **Note:** `lowestCurrent()` drops empty tables.
//
// ## Parameters
// - n: Number of records to return.
// - column: Column to evaluate. Default is `_value`.
// - groupColumns: List of columns to group by. Default is `[]`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the lowest current value from a stream of tables
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> lowestCurrent(n: 1, groupColumns: ["tag"])
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, selectors
//
lowestCurrent = (n, column="_value", groupColumns=[], tables=<-) =>
    tables
        |> _highestOrLowest(
            n: n,
            column: column,
            groupColumns: groupColumns,
            reducer: (tables=<-) => tables |> last(column: column),
            _sortLimit: bottom,
        )

// timedMovingAverage returns the mean of values in a defined time range at a
// specified frequency.
//
// For each row in a table, `timedMovingAverage()` returns the average of the
// current value and all row values in the previous `period` (duration).
// It returns moving averages at a frequency defined by the `every` parameter.
//
// #### Aggregate by calendar months and years
// `every` and `period` parameters support all valid duration units, including
// calendar months (`1mo`) and years (`1y`).
//
// #### Aggregate by week
// When aggregating by week (`1w`), weeks are determined using the Unix epoch
// (1970-01-01T00:00:00Z UTC). The Unix epoch was on a Thursday, so all
// calculated weeks begin on Thursday.
//
// ## Parameters
// - every: Frequency of time window.
// - period: Length of each averaged time window.
//
//   A negative duration indicates start and stop boundaries are reversed.
//
// - column: Column to operate on. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate a five year moving average every year
// ```
// # import "generate"
// #
// # timeRange = {start: 2015-01-01T00:00:00Z, stop: 2021-01-01T00:00:00Z}
// # data = generate.from(count: 6, fn: (n) => n * n, start: timeRange.start, stop: timeRange.stop)
// #     |> range(start: timeRange.start, stop: timeRange.stop)
// #
// < data
// >     |> timedMovingAverage(every: 1y, period: 5y)
// ```
//
// ## Metadata
// introduced: 0.36.0
// tags: transformations
//
timedMovingAverage = (every, period, column="_value", tables=<-) =>
    tables
        |> window(every: every, period: period)
        |> mean(column: column)
        |> duplicate(column: "_stop", as: "_time")
        |> window(every: inf)

// doubleEMA returns the double exponential moving average (DEMA) of values in
// the `_value` column grouped into `n` number of points, giving more weight to
// recent data.
//
// #### Double exponential moving average rules
// - A double exponential moving average is defined as `doubleEMA = 2 * EMA_N - EMA of EMA_N`.
//     - `EMA` is an exponential moving average.
//     - `N = n` is the period used to calculate the `EMA`.
// - A true double exponential moving average requires at least `2 * n - 1` values.
//   If not enough values exist to calculate the double `EMA`, it returns a `NaN` value.
// - `doubleEMA()` inherits all `exponentialMovingAverage()` rules.
//
// ## Parameters
// - n: Number of points to average.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate a three point double exponential moving average
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> doubleEMA(n: 3)
// ```
//
// ## Metadata
// introduced: 0.38.0
// tags: transformations
//
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
// ## Metadata
// introduced: 0.40.0
// tags: transformations
//
kaufmansER = (n, tables=<-) =>
    tables
        |> chandeMomentumOscillator(n: n)
        |> map(fn: (r) => ({r with _value: math.abs(x: r._value) / 100.0}))

// tripleEMA returns the triple exponential moving average (TEMA) of values in
// the `_value` column.
//
// `tripleEMA` uses `n` number of points to calculate the TEMA, giving more
// weight to recent data with less lag than `exponentialMovingAverage()` and
// `doubleEMA()`.
//
// #### Triple exponential moving average rules
// - A triple exponential moving average is defined as `tripleEMA = (3 * EMA_1) - (3 * EMA_2) + EMA_3`.
//     - `EMA_1` is the exponential moving average of the original data.
//     - `EMA_2` is the exponential moving average of `EMA_1`.
//     - `EMA_3` is the exponential moving average of `EMA_2`.
// - A true triple exponential moving average requires at least requires at least
//   `3 * n - 2` values. If not enough values exist to calculate the TEMA, it
//   returns a `NaN` value.
// - `tripleEMA()` inherits all `exponentialMovingAverage()` rules.
//
// ## Parameters
// - n: Number of points to use in the calculation.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate a three point triple exponential moving average
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> tripleEMA(n: 3)
// ```
//
// ## Metadata
// introduced: 0.38.0
// tags: transformations
//
tripleEMA = (n, tables=<-) =>
    tables
        |> exponentialMovingAverage(n: n)
        |> duplicate(column: "_value", as: "__ema1")
        |> exponentialMovingAverage(n: n)
        |> duplicate(column: "_value", as: "__ema2")
        |> exponentialMovingAverage(n: n)
        |> map(fn: (r) => ({r with _value: 3.0 * r.__ema1 - 3.0 * r.__ema2 + r._value}))
        |> drop(columns: ["__ema1", "__ema2"])

// truncateTimeColumn truncates all input time values in the `_time` to a
// specified unit.
//
// #### Truncate to weeks
// When truncating a time value to the week (`1w`), weeks are determined using the
// **Unix epoch (1970-01-01T00:00:00Z UTC)**. The Unix epoch was on a Thursday,
// so all calculated weeks begin on Thursday.
//
// ## Parameters
// - unit: Unit of time to truncate to.
//
//   **Example units:**
//   - 1ns (nanosecond)
//   - 1us (microsecond)
//   - 1ms (millisecond)
//   - 1s (second)
//   - 1m (minute)
//   - 1h (hour)
//   - 1d (day)
//   - 1w (week)
//   - 1mo (month)
//   - 1y (year)
//
// - timeColumn: Time column to truncate. Default is `_time`.
//
//   **Note:** Currently, assigning a custom value to this parameter will have
//   no effect.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Truncate all time values to the minute
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> truncateTimeColumn(unit: 1m)
// ```
//
// ## Metadata
// introduced: 0.37.0
// tags: transformations, date/time
//
truncateTimeColumn = (timeColumn="_time", unit, tables=<-) =>
    tables
        |> map(fn: (r) => ({r with _time: date.truncate(t: r._time, unit: unit)}))

// toString converts all values in the `_value` column to string types.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Convert the _value column to strings
// ```
// import "sampledata"
//
// < sampledata.float()
// >     |> toString()
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, type-conversions
//
toString = (tables=<-) => tables |> map(fn: (r) => ({r with _value: string(v: r._value)}))

// toInt converts all values in the `_value` column to integer types.
//
// #### Supported types and behaviors
// `toInt()` behavior depends on the `_value` column type:
//
// | _value type | Returned value                                  |
// | :---------- | :---------------------------------------------- |
// | string      | Integer equivalent of the numeric string        |
// | bool        | 1 (true) or 0 (false)                           |
// | duration    | Number of nanoseconds in the specified duration |
// | time        | Equivalent nanosecond epoch timestamp           |
// | float       | Value truncated at the decimal                  |
// | uint        | Integer equivalent of the unsigned integer      |
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Convert a float _value column to integers
// ```
// import "sampledata"
//
// < sampledata.float()
// >     |> toInt()
// ```
//
// ### Convert a boolean _value column to integers
// ```
// import "sampledata"
//
// < sampledata.bool()
// >     |> toInt()
// ```
//
// ### Convert a uinteger _value column to an integers
// ```
// import "sampledata"
//
// < sampledata.uint()
// >     |> toInt()
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, type-conversions
//
toInt = (tables=<-) => tables |> map(fn: (r) => ({r with _value: int(v: r._value)}))

// toUInt converts all values in the `_value` column to unsigned integer types.
//
// #### Supported types and behaviors
// `toUInt()` behavior depends on the `_value` column type:
//
// | _value type | Returned value                                  |
// | :---------- | :---------------------------------------------- |
// | string      | UInteger equivalent of the numeric string       |
// | bool        | 1 (true) or 0 (false)                           |
// | duration    | Number of nanoseconds in the specified duration |
// | time        | Equivalent nanosecond epoch timestamp           |
// | float       | Value truncated at the decimal                  |
// | int         | UInteger equivalent of the integer              |
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Convert a float _value column to uintegers
// ```
// import "sampledata"
//
// < sampledata.float()
// >     |> toUInt()
// ```
//
// ### Convert a boolean _value column to uintegers
// ```
// import "sampledata"
//
// < sampledata.bool()
// >     |> toUInt()
// ```
//
// ### Convert a uinteger _value column to an uintegers
// ```
// import "sampledata"
//
// < sampledata.uint()
// >     |> toUInt()
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, type-conversions
//
toUInt = (tables=<-) => tables |> map(fn: (r) => ({r with _value: uint(v: r._value)}))

// toFloat converts all values in the `_value` column to float types.
//
// #### Supported data types
// - string (numeric, scientific notation, ±Inf, or NaN)
// - boolean
// - int
// - uint
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Convert an integer _value column to floats
// ```
// import "sampledata"
//
// < sampledata.int()
// >     |> toFloat()
// ```
//
// ### Convert a boolean _value column to floats
// ```
// import "sampledata"
//
// < sampledata.bool()
// >     |> toFloat()
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, type-conversions
//
toFloat = (tables=<-) => tables |> map(fn: (r) => ({r with _value: float(v: r._value)}))

// toBool converts all values in the `_value` column to boolean types.
//
// #### Supported data types
// - **string**: `true` or `false`
// - **int**: `1` or `0`
// - **uint**: `1` or `0`
// - **float**: `1.0` or `0.0`
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Convert an integer _value column to booleans
// ```
// import "sampledata"
//
// < sampledata.numericBool()
// >     |> toBool()
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, type-conversions
//
toBool = (tables=<-) => tables |> map(fn: (r) => ({r with _value: bool(v: r._value)}))

// toTime converts all values in the `_value` column to time types.
//
// #### Supported data types
// - string (RFC3339 timestamp)
// - int
// - uint
//
// `toTime()` treats all numeric input values as nanosecond epoch timestamps.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Convert an integer _value column to times
// ```
// # import "sampledata"
// #
// # data = sampledata.int()
// #     |> map(fn: (r) => ({r with _value: r._value * 10000000000000000}))
// #
// < data
// >     |> toTime()
// ```
//
// ## Metadata
// introduced: 0.7.0
// tags: transformations, type-conversions
//
toTime = (tables=<-) => tables |> map(fn: (r) => ({r with _value: time(v: r._value)}))

// today returns the now() timestamp truncated to the day unit.
//
// ## Examples
//
// ### Return a timestamp representing today
// ```no_run
// option now = () => 2022-01-01T13:45:28Z
//
// today()
// // Returns 2022-01-01T00:00:00.000000000Z
// ```
//
// ### Query data from today
// ```no_run
// from(bucket: "example-bucket")
//     |> range(start: today())
// ```
//
// ## Metadata
// introduced: 0.116.0
// tags: date/time
//
today = () => date.truncate(t: now(), unit: 1d)
