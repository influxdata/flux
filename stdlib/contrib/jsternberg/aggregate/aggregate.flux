// Package aggregate provides an API for computing multiple aggregates over multiple columns within the same table stream.
//
// ## Metadata
// introduced: 0.77.0
// contributors: **GitHub**: [@jsternberg](https://github.com/jsternberg) | **InfluxDB Slack**: [@Jonathan Sternberg](https://influxdata.com/slack)
//
package aggregate


import "contrib/jsternberg/math"

// table will aggregate columns and create tables with a single
// row containing the aggregated value.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
// - columns: Columns to aggregate and which aggregate method to use.
//
//      Columns is a record where the key is a column name and the value is an aggregate record.
//      The aggregate record is composed of at least the following required attributes:
//          - **column**: Input column name (string).
//          - **init**: A function to compute the initial state of the
//              output. This can return either the final aggregate or a
//              temporary state object that can be used to compute the
//              final aggregate. The `values` parameter will always be a
//              non-empty array of values from the specified column.
//              For example: `(values) => state`.
//          - **reduce**: A function that takes in another buffer of values
//              and the current state of the aggregate and computes
//              the updated state.
//              For example: `(values, state) => state`.
//          - **compute**: A function that takes the state and computes the final
//              aggregate. For example, `(state) => value`.
//          - **fill**: The value passed to `fill()`. If present, the fill value
//              determines what the aggregate does when there are no values.
//              This can either be a value or one of the predefined
//              identifiers of `null` or `none`.
//              This value must be the same type as the value return from
//              `compute`.
//
// ## Examples
//
// ### Compute the min of a specific column
//
// ```
// import "sampledata"
// import "contrib/jsternberg/aggregate"
//
// > sampledata.float()
//      |> aggregate.table(columns: {
//         "min_bottom_degrees": aggregate.min(column: "_value"),
// <    })
// ```
builtin table : (<-tables: stream[A], columns: C) => stream[B] where A: Record, B: Record, C: Record

// window aggregates columns and create tables by
// organizing incoming points into windows.
//
// Each table will have two additional columns: start and stop.
// These are the start and stop times for each interval.
// It is not possible to use start or stop as destination column
// names with this function. The start and stop columns are not
// added to the group key.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
// - columns: Columns to aggregate and which aggregate method to use. See `aggregate.table()` for details.
// - every: Duration between the start of each interval.
// - time: Column name for the time input. Defaults to `_time` or `time` (whichever is earlier in the list of columns).
// - period: Length of the interval. Defaults to the `every` duration.
builtin window : (
        <-tables: stream[A],
        ?time: string,
        every: duration,
        ?period: duration,
        columns: C,
    ) => stream[B]
    where
    A: Record,
    B: Record,
    C: Record

// null is a sentinel value for fill that will fill
// in a null value if there were no values for an interval.
builtin null : A

// none is a sentinel value for fill that will skip
// emitting a row if there are no values for an interval.
builtin none : A

// define constructs an aggregate function record.
//
// ## Parameters
// - init: Function to compute the initial state of the
//     output. This can return either the final aggregate or a
//     temporary state object that can be used to compute the
//     final aggregate. The `values` parameter will always be a
//     non-empty array of values from the specified column.
// - reduce: Function that takes in another buffer of values
//     and the current state of the aggregate and computes
//     the updated state.
// - compute: Function that takes the state and computes the final aggregate.
// - fill: Value passed to `fill()`. If present, the fill value determines what
//     the aggregate does when there are no values.
//     This can either be a value or one of the predefined
//     identifiers, `null` or `none`.
//     This value must be the same type as the value return from
//     compute.
define = (init, reduce, compute, fill=null) =>
    (column="_value", fill=fill) =>
        ({
            column: column,
            init: init,
            reduce: reduce,
            compute: compute,
            fill: fill,
        })
_make_selector = (fn) =>
    define(
        init: (values) => fn(values),
        reduce: (values, state) => {
            v = fn(values)

            return fn(values: [state, v])
        },
        compute: (state) => state,
    )

// min constructs a min aggregate or selector for the column.
//
// ## Parameters
// - column: Name of the column to aggregate.
// - fill: When set, value to replace missing values.
min = _make_selector(fn: math.min)

// max constructs a max aggregate or selector for the column.
//
// ## Parameters
// - column: Name of the column to aggregate.
// - fill: When set, value to replace missing values.
max = _make_selector(fn: math.max)

// sum constructs a sum aggregate for the column.
//
// ## Parameters
// - column: Name of the column to aggregate.
// - fill: When set, value to replace missing values.
sum =
    define(
        init: (values) => math.sum(values),
        reduce: (values, state) => {
            return state + math.sum(values)
        },
        compute: (state) => state,
    )

// count constructs a count aggregate for the column.
//
// ## Parameters
// - column: Name of the column to aggregate.
// - fill: When set, value to replace missing values.
count =
    define(
        init: (values) => length(arr: values),
        reduce: (values, state) => {
            return state + length(arr: values)
        },
        compute: (state) => state,
        fill: 0,
    )

// mean constructs a mean aggregate for the column.
//
// ## Parameters
// - column: Name of the column to aggregate.
// - fill: When set, value to replace missing values.
mean =
    define(
        init: (values) => ({sum: math.sum(values), count: length(arr: values)}),
        reduce: (values, state) => ({sum: state.sum + math.sum(values), count: state.count + length(arr: values)}),
        compute: (state) => float(v: state.sum) / float(v: state.count),
    )
