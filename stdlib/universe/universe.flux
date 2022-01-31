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
// Each instance of system.time() in a Flux script returns a unique value.
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

// distinct ...
builtin distinct : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin drop : (<-tables: [A], ?fn: (column: string) => bool, ?columns: [string]) => [B] where A: Record, B: Record
builtin duplicate : (<-tables: [A], column: string, as: string) => [B] where A: Record, B: Record
builtin elapsed : (<-tables: [A], ?unit: duration, ?timeColumn: string, ?columnName: string) => [B]
    where
    A: Record,
    B: Record
builtin exponentialMovingAverage : (<-tables: [{B with _value: A}], n: int) => [{B with _value: A}] where A: Numeric
builtin fill : (<-tables: [A], ?column: string, ?value: B, ?usePrevious: bool) => [C] where A: Record, C: Record
builtin filter : (<-tables: [A], fn: (r: A) => bool, ?onEmpty: string) => [A] where A: Record
builtin first : (<-tables: [A], ?column: string) => [A] where A: Record
builtin group : (<-tables: [A], ?mode: string, ?columns: [string]) => [A] where A: Record
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

builtin hourSelection : (<-tables: [A], start: int, stop: int, ?timeColumn: string) => [A] where A: Record
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

builtin join : (<-tables: A, ?method: string, ?on: [string]) => [B] where A: Record, B: Record
builtin kaufmansAMA : (<-tables: [A], n: int, ?column: string) => [B] where A: Record, B: Record
builtin keep : (<-tables: [A], ?columns: [string], ?fn: (column: string) => bool) => [B] where A: Record, B: Record
builtin keyValues : (<-tables: [A], ?keyColumns: [string]) => [{C with _key: string, _value: B}]
    where
    A: Record,
    C: Record
builtin keys : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin last : (<-tables: [A], ?column: string) => [A] where A: Record
builtin limit : (<-tables: [A], n: int, ?offset: int) => [A]
builtin map : (<-tables: [A], fn: (r: A) => B, ?mergeKey: bool) => [B]
builtin max : (<-tables: [A], ?column: string) => [A] where A: Record
builtin mean : (<-tables: [A], ?column: string) => [B] where A: Record, B: Record
builtin min : (<-tables: [A], ?column: string) => [A] where A: Record
builtin mode : (<-tables: [A], ?column: string) => [{C with _value: B}] where A: Record, C: Record
builtin movingAverage : (<-tables: [{B with _value: A}], n: int) => [{B with _value: float}] where A: Numeric
builtin quantile : (
        <-tables: [A],
        ?column: string,
        q: float,
        ?compression: float,
        ?method: string,
    ) => [A]
    where
    A: Record

builtin pivot : (<-tables: [A], rowKey: [string], columnKey: [string], valueColumn: string) => [B]
    where
    A: Record,
    B: Record
builtin range : (
        <-tables: [{A with _time: time}],
        start: B,
        ?stop: C,
    ) => [{A with _time: time, _start: time, _stop: time}]

builtin reduce : (<-tables: [A], fn: (r: A, accumulator: B) => B, identity: B) => [C]
    where
    A: Record,
    B: Record,
    C: Record
builtin relativeStrengthIndex : (<-tables: [A], n: int, ?columns: [string]) => [B] where A: Record, B: Record
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

// kaufmansER computes Kaufman's Efficiency Ratios of the `_value` column
kaufmansER = (n, tables=<-) =>
    tables
        |> chandeMomentumOscillator(n: n)
        |> map(fn: (r) => ({r with _value: math.abs(x: r._value) / 100.0}))
toString = (tables=<-) => tables |> map(fn: (r) => ({r with _value: string(v: r._value)}))
toInt = (tables=<-) => tables |> map(fn: (r) => ({r with _value: int(v: r._value)}))
toUInt = (tables=<-) => tables |> map(fn: (r) => ({r with _value: uint(v: r._value)}))
toFloat = (tables=<-) => tables |> map(fn: (r) => ({r with _value: float(v: r._value)}))
toBool = (tables=<-) => tables |> map(fn: (r) => ({r with _value: bool(v: r._value)}))
toTime = (tables=<-) => tables |> map(fn: (r) => ({r with _value: time(v: r._value)}))

// today returns the now() timestamp truncated to the day unit
today = () => date.truncate(t: now(), unit: 1d)
