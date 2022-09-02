// Package experimental includes experimental functions and packages.
//
// ### Experimental packages are subject to change
// Please note that experimental packages and functions may:
//
// - be moved or promoted to a permanent location
// - undergo API changes
// - stop working with no planned fixes
// - be removed without warning or explanation
//
// ## Metadata
// introduced: 0.39.0
package experimental


import "influxdata/influxdb"
import "date"

// addDuration adds a duration to a time value and returns the resulting time value.
//
// **Deprecated**: `experimental.addDuration()` is deprecated in favor of [`date.add()`](https://docs.influxdata.com/flux/v0.x/stdlib/date/add/).
//
// ## Parameters
// - d: Duration to add.
// - to: Time to add the duration to.
// - location: Location to use for the time value.
//
//   Use an absolute time or a relative duration.
//   Durations are relative to `now()`.
//
// ## Examples
//
// ### Add six hours to a timestamp
// ```no_run
// import "experimental"
//
// experimental.addDuration(
//     d: 6h,
//     to: 2019-09-16T12:00:00Z,
// )
//
// // Returns 2019-09-16T18:00:00.000000000Z
// ```
//
// ### Add one month to yesterday
//
// A time may be represented as either an explicit timestamp
// or as a relative time from the current `now` time. addDuration can
// support either type of value.
//
// ```no_run
// import "experimental"
//
// option now = () => 2021-12-10T16:27:40Z
//
// experimental.addDuration(d: 1mo, to: -1d)
//
// // Returns 2022-01-09T16:27:40Z
// ```
//
// ### Add six hours to a relative duration
// ```no_run
// import "experimental"
//
// option now = () => 2022-01-01T12:00:00Z
//
// experimental.addDuration(d: 6h, to: 3h)
//
// // Returns 2022-01-01T21:00:00.000000000Z
// ```
//
// ## Metadata
// introduced: 0.39.0
// deprecated: 0.162.0
// tags: date/time
//
addDuration = (d, to, location=location) => date.add(d, to, location)

// subDuration subtracts a duration from a time value and returns the resulting time value.
//
// **Deprecated**: `experimental.subDuration()` is deprecated in favor of [`date.sub()`](https://docs.influxdata.com/flux/v0.x/stdlib/date/sub/).
//
// ## Parameters
// - from: Time to subtract the duration from.
//
//   Use an absolute time or a relative duration.
//   Durations are relative to `now()`.
//
// - d: Duration to subtract.
// - location: Location to use for the time value.
//
// ## Examples
//
// ### Subtract six hours from a timestamp
// ```no_run
// import "experimental"
//
// experimental.subDuration(from: 2019-09-16T12:00:00Z, d: 6h)
//
// // Returns 2019-09-16T06:00:00.000000000Z
// ```
//
// ### Subtract six hours from a relative duration
// ```no_run
// import "experimental"
//
// option now = () => 2022-01-01T12:00:00Z
//
// experimental.subDuration(d: 6h, from: -3h)
//
// // Returns 2022-01-01T03:00:00.000000000Z
// ```
//
// ### Subtract two days from one hour ago
//
// A time may be represented as either an explicit timestamp
// or as a relative time from the current `now` time. subDuration can
// support either type of value.
//
// ```no_run
// import "experimental"
//
// option now = () => 2021-12-10T16:27:40Z
//
// experimental.subDuration(
//     from: -1h,
//     d: 2d,
// )
//
// // Returns 2021-12-08T15:27:40Z
// ```
//
// ## Metadata
// introduced: 0.39.0
// deprecated: 0.162.0
// tags: date/time
//
subDuration = (d, from, location=location) => date.sub(d, from, location)

// group introduces an `extend` mode to the existing `group()` function.
//
// ## Parameters
// - columns: List of columns to use in the grouping operation. Default is `[]`.
// - mode: Grouping mode. `extend` is the only mode available to `experimental.group()`.
//
//   #### Grouping modes
//   - **extend**: Appends columns defined in the `columns` parameter to group keys.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Add a column to the group key
// ```
// # import "array"
// import "experimental"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, host: "host1", region: "east", _value: "41"},
// #         {_time: 2021-01-01T00:01:00Z, host: "host1", region: "east", _value: "48"},
// #         {_time: 2021-01-01T00:00:00Z, host: "host1", region: "west", _value: "34"},
// #         {_time: 2021-01-01T00:01:00Z, host: "host1", region: "west", _value: "12"},
// #         {_time: 2021-01-01T00:00:00Z, host: "host2", region: "east", _value: "56"},
// #         {_time: 2021-01-01T00:01:00Z, host: "host2", region: "east", _value: "72"},
// #         {_time: 2021-01-01T00:00:00Z, host: "host2", region: "west", _value: "43"},
// #         {_time: 2021-01-01T00:01:00Z, host: "host2", region: "west", _value: "22"},
// #     ],
// # )
// #     |> group(columns: ["host"])
//
// < data
// >     |> experimental.group(columns: ["region"], mode: "extend")
// ```
//
// ## Metadata
// tags: transformations
//
builtin group : (<-tables: stream[A], mode: string, columns: [string]) => stream[A] where A: Record

// objectKeys returns an array of property keys in a specified record.
//
// ## Parameters
// - o: Record to return property keys from.
//
// ## Examples
// ### Return all property keys in a record
// ```no_run
// import "experimental"
//
// user = {
//     firstName: "John",
//     lastName: "Doe",
//     age: 42,
// }
//
// experimental.objectKeys(o: user)
// // Returns [firstName, lastName, age]
// ```
//
// ## Metadata
// introduced: 0.40.0
//
builtin objectKeys : (o: A) => [string] where A: Record

// set sets multiple static column values on all records.
//
// If a column already exists, the function updates the existing value.
// If a column does not exist, the function adds it with the specified value.
//
// ## Parameters
// - o: Record that defines the columns and values to set.
//
//   The key of each key-value pair defines the column name.
//   The value of each key-value pair defines the column value.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Set values for multiple columns
// ```
// # import "array"
// import "experimental"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2019-09-16T12:00:00Z, _field: "temp", _value: 71.2},
// #         {_time: 2019-09-17T12:00:00Z, _field: "temp", _value: 68.4},
// #         {_time: 2019-09-18T12:00:00Z, _field: "temp", _value: 70.8},
// #     ],
// # )
//
// < data
//     |> experimental.set(
//         o: {
//             _field: "temperature",
//             unit: "°F",
//             location: "San Francisco",
//         },
// >     )
// ```
//
// ## Metadata
// introduced: 0.40.0
// tags: transformations
//
builtin set : (<-tables: stream[A], o: B) => stream[C] where A: Record, B: Record, C: Record

// to writes _pivoted_ data to an InfluxDB 2.x or InfluxDB Cloud bucket.
//
// **Deprecated**: `experimental.to()` is deprecated in favor of [`wideTo()`](/flux/v0.x/stdlib/influxdata/influxdb/wideto/),
// which is an equivalent function.
//
// #### Requirements and behavior
// - Requires both a `_time` and a `_measurement` column.
// - All columns in the group key (other than `_measurement`) are written as tags
//   with the column name as the tag key and the column value as the tag value.
// - All columns **not** in the group key (other than `_time`) are written as
//   fields with the column name as the field key and the column value as the field value.
//
// If using `from()` to query data from InfluxDB, use `pivot()` to transform
// data into the structure `experimental.to()` expects.
//
// ## Parameters
// - bucket: Name of the bucket to write to.
//   _`bucket` and `bucketID` are mutually exclusive_.
// - bucketID: String-encoded bucket ID to to write to.
//   _`bucket` and `bucketID` are mutually exclusive_.
// - host: URL of the InfluxDB instance to write to.
//
//     See [InfluxDB Cloud regions](https://docs.influxdata.com/influxdb/cloud/reference/regions/)
//     or [InfluxDB OSS URLs](https://docs.influxdata.com/influxdb/latest/reference/urls/).
//
//     `host` is required when writing to a remote InfluxDB instance.
//     If specified, `token` is also required.
//
// - org: Organization name.
//   _`org` and `orgID` are mutually exclusive_.
// - orgID: String-encoded organization ID to query.
//   _`org` and `orgID` are mutually exclusive_.
// - token: InfluxDB API token.
//
//     **InfluxDB 1.x or Enterprise**: If authentication is disabled, provide an
//     empty string (`""`). If authentication is enabled, provide your InfluxDB
//     username and password using the `<username>:<password>` syntax.
//
//     `token` is required when writing to another organization or when `host`
//     is specified.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
//
// ## Examples
//
// ### Pivot and write data to InfluxDB
// ```no_run
// import "experimental"
//
// from(bucket: "example-bucket")
//     |> range(start: -1h)
//     |> pivot(
//         rowKey: ["_time"],
//         columnKey: ["_field"],
//         valueColumn: "_value",
//     )
//     |> experimental.to(bucket: "example-target-bucket")
// ```
//
// ## Metadata
// introduced: 0.40.0
// deprecated: 0.174.0
// tags: outputs
//
to = influxdb.wideTo

// join joins two streams of tables on the **group key and `_time` column**.
//
// **Deprecated**: `experimental.join()` is deprecated in favor of [`join.time()`](https://docs.influxdata.com/flux/v0.x/stdlib/join/time/).
// The [`join` package](https://docs.influxdata.com/flux/v0.x/stdlib/join/) provides support
// for multiple join methods.
//
// Use the `fn` parameter to map new output tables using values from input tables.
//
// **Note**: To join streams of tables with different fields or measurements,
// use `group()` or `drop()` to remove `_field` and `_measurement` from the
// group key before joining.
//
// ## Parameters
// - left: First of two streams of tables to join.
// - right: Second of two streams of tables to join.
// - fn: Function with left and right arguments that maps a new output record
//   using values from the `left` and `right` input records.
//   The return value must be a record.
//
// ## Examples
// ### Join two streams of tables
// ```
// import "array"
// import "experimental"
//
// left = array.from(
//     rows: [
//         {_time: 2021-01-01T00:00:00Z, _field: "temp", _value: 80.1},
//         {_time: 2021-01-01T01:00:00Z, _field: "temp", _value: 80.6},
//         {_time: 2021-01-01T02:00:00Z, _field: "temp", _value: 79.9},
//         {_time: 2021-01-01T03:00:00Z, _field: "temp", _value: 80.1},
//     ],
// )
// right = array.from(
//     rows: [
//         {_time: 2021-01-01T00:00:00Z, _field: "temp", _value: 75.1},
//         {_time: 2021-01-01T01:00:00Z, _field: "temp", _value: 72.6},
//         {_time: 2021-01-01T02:00:00Z, _field: "temp", _value: 70.9},
//         {_time: 2021-01-01T03:00:00Z, _field: "temp", _value: 71.1},
//     ],
// )
//
// experimental.join(
//     left: left,
//     right: right,
//     fn: (left, right) => ({left with
//         lv: left._value,
//         rv: right._value,
//         diff: left._value - right._value,
//     }),
// > )
// ```
//
// ### Join two streams of tables with different fields and measurements
// ```no_run
// import "experimental"
//
// s1 = from(bucket: "example-bucket")
//     |> range(start: -1h)
//     |> filter(fn: (r) => r._measurement == "foo" and r._field == "bar")
//     |> group(columns: ["_time", "_measurement", "_field", "_value"], mode: "except")
//
// s2 = from(bucket: "example-bucket")
//     |> range(start: -1h)
//     |> filter(fn: (r) => r._measurement == "baz" and r._field == "quz")
//     |> group(columns: ["_time", "_measurement", "_field", "_value"], mode: "except")
//
// experimental.join(
//     left: s1,
//     right: s2,
//     fn: (left, right) => ({left with
//         bar_value: left._value,
//         quz_value: right._value,
//     }),
// )
// ```
//
// ## Metadata
// introduced: 0.65.0
// deprecated: 0.172.0
// tags: transformations
//
builtin join : (left: stream[A], right: stream[B], fn: (left: A, right: B) => C) => stream[C]
    where
    A: Record,
    B: Record,
    C: Record

// chain runs two queries in a single Flux script sequentially and outputs the
// results of the second query.
//
// Flux typically executes multiple queries in a single script in parallel.
// Running the queries sequentially ensures any dependencies the second query
// has on the results of the first query are met.
//
// ##### Applicable use cases
// - Write to an InfluxDB bucket and query the written data in a single Flux script.
//
//   _**Note:** `experimental.chain()` does not gaurantee that data written to
//   InfluxDB is immediately queryable. A delay between when data is written and
//   when it is queryable may cause a query using `experimental.chain()` to fail.
//
// - Execute queries sequentially in testing scenarios.
//
// ## Parameters
// - first: First query to execute.
// - second: Second query to execute.
//
// ## Examples
// ### Write to a bucket and query the written data
// ```no_run
// import "experimental"
//
// downsampled_max = from(bucket: "example-bucket-1")
//     |> range(start: -1d)
//     |> filter(fn: (r) => r._measurement == "example-measurement")
//     |> aggregateWindow(every: 1h, fn: max)
//     |> to(bucket: "downsample-1h-max", org: "example-org")
//
// average_max = from(bucket: "downsample-1h-max")
//     |> range(start: -1d)
//     |> filter(fn: (r) => r.measurement == "example-measurement")
//     |> mean()
//
// experimental.chain(
//     first: downsampled_max,
//     second: average_max,
// )
// ```
//
// ## Metadata
// introduced: 0.68.0
//
builtin chain : (first: stream[A], second: stream[B]) => stream[B] where A: Record, B: Record

// alignTime shifts time values in input tables to all start at a common start time.
//
// ## Parameters
// - alignTo: Time to align tables to. Default is `1970-01-01T00:00:00Z`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Compare month-over-month values
// 1. Window data by calendar month creating two separate tables (one for January and one for February).
// 2. Align tables to `2021-01-01T00:00:00Z`.
//
// Each output table represents data from a calendar month.
// When visualized, data is still grouped by month, but timestamps are aligned
// to a common start time and values can be compared by time.
//
// ```
// # import "array"
// import "experimental"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, _value: 32.1},
// #         {_time: 2021-01-02T00:00:00Z, _value: 32.9},
// #         {_time: 2021-01-03T00:00:00Z, _value: 33.2},
// #         {_time: 2021-01-04T00:00:00Z, _value: 34.0},
// #         {_time: 2021-02-01T00:00:00Z, _value: 38.3},
// #         {_time: 2021-02-02T00:00:00Z, _value: 38.4},
// #         {_time: 2021-02-03T00:00:00Z, _value: 37.8},
// #         {_time: 2021-02-04T00:00:00Z, _value: 37.5},
// #     ],
// # )
// #     |> range(start: 2021-01-01T00:00:00Z, stop: 2021-03-01T00:00:00Z)
//
// < data
//     |> window(every: 1mo)
// >     |> experimental.alignTime(alignTo: 2021-01-01T00:00:00Z)
// ```
//
// ## Metadata
// introduced: 0.66.0
// tags: transformations, date/time
//
alignTime = (tables=<-, alignTo=time(v: 0)) =>
    tables
        |> stateDuration(fn: (r) => true, column: "timeDiff", unit: 1ns)
        |> map(fn: (r) => ({r with _time: time(v: int(v: alignTo) + r.timeDiff)}))
        |> drop(columns: ["timeDiff"])

builtin _window : (
        <-tables: stream[{T with _start: time, _stop: time, _time: time}],
        every: duration,
        period: duration,
        offset: duration,
        location: {zone: string, offset: duration},
        createEmpty: bool,
    ) => stream[{T with _start: time, _stop: time, _time: time}]

// window groups records based on time.
//
// `_start` and `_stop` columns are updated to reflect the bounds of
// the window the row's time value is in.
// Input tables must have `_start`, `_stop`, and `_time` columns.
//
// A single input record can be placed into zero or more output tables depending
// on the specific windowing function.
//
// By default the start boundary of a window will align with the Unix epoch
// modified by the offset of the `location` option.
//
// #### Calendar months and years
// `every`, `period`, and `offset` support all valid duration units, including
// calendar months (`1mo`) and years (`1y`).
//
// ## Parameters
// - every: Duration of time between windows. Default is the `0s`.
// - period: Duration of the window. Default is `0s`.
//
//   Period is the length of each interval.
//   It can be negative, indicating the start and stop boundaries are reversed.
//
// - offset: Duration to shift the window boundaries by. Default is 0s.
//
//   `offset` can be negative, indicating that the offset goes backwards in time.
//
// - location: Location used to determine timezone. Default is the `location` option.
// - createEmpty: Create empty tables for empty windows. Default is `false`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Window data into thirty second intervals
// ```
// import "experimental"
// # import "sampledata"
// #
// # data = sampledata.int()
// #     |> range(start: sampledata.start, stop: sampledata.stop)
//
// < data
// >     |> experimental.window(every: 30s)
// ```
//
// ### Window by calendar month
// ```
// # import "array"
// import "experimental"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, _value: 32.1},
// #         {_time: 2021-01-02T00:00:00Z, _value: 32.9},
// #         {_time: 2021-01-03T00:00:00Z, _value: 33.2},
// #         {_time: 2021-02-01T00:00:00Z, _value: 38.3},
// #         {_time: 2021-02-02T00:00:00Z, _value: 38.4},
// #         {_time: 2021-02-03T00:00:00Z, _value: 37.8},
// #     ],
// # )
// #     |> range(start: 2021-01-01T00:00:00Z, stop: 2021-03-01T00:00:00Z)
//
// < data
// >     |> experimental.window(every: 1mo)
// ```
//
// ## Metadata
// introduced: 0.106.0
// tags: transformations,date/time
//
window = (
    tables=<-,
    every=0s,
    period=0s,
    offset=0s,
    location=location,
    createEmpty=false,
) =>
    tables
        |> _window(
            every,
            period,
            offset,
            location,
            createEmpty,
        )

// integral computes the area under the curve per unit of time of subsequent non-null records.
//
// The curve is defined using `_time` as the domain and record values as the range.
//
// Input tables must have `_start`, _stop`, `_time`, and `_value` columns.
// `_start` and `_stop` must be part of the group key.
//
// ## Parameters
// - unit: Time duration used to compute the integral.
// - interpolate: Type of interpolation to use. Default is `""` (no interpolation).
//
//   Use one of the following interpolation options:
//
//   - empty string (`""`) for no interpolation
//   - linear
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Calculate the integral
// ```
// import "experimental"
// import "sampledata"
//
// data = sampledata.int()
//     |> range(start: sampledata.start, stop: sampledata.stop)
//
// < data
// >     |> experimental.integral(unit: 20s)
// ```
//
// ### Calculate the integral with linear interpolation
// ```
// import "experimental"
// import "sampledata"
//
// data = sampledata.int()
//     |> range(start: sampledata.start, stop: sampledata.stop)
//
// < data
//     |> experimental.integral(
//         unit: 20s,
//         interpolate: "linear",
// >     )
// ```
//
// ## Metadata
// introduced: 0.106.0
// tags: transformations, aggregates
//
builtin integral : (
        <-tables: stream[{T with _time: time, _value: B}],
        ?unit: duration,
        ?interpolate: string,
    ) => stream[{T with _value: B}]

// count returns the number of records in each input table.
//
// The count is returned in the `_value` column and counts both null and non-null records.
//
// #### Counts on empty tables
// `experimental.count()` returns 0 for empty tables.
// To keep empty tables in your data, set the following parameters when using
// the following functions:
//
// ```
// filter(onEmpty: "keep")
// window(createEmpty: true)
// aggregateWindow(createEmpty: true)
// ```
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Count the number of rows in a table
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.int()
// >     |> experimental.count()
// ```
//
// ## Metadata
// introduced: 0.107.0
// tags: transformations,aggregates
//
builtin count : (<-tables: stream[{T with _value: A}]) => stream[{T with _value: int}]

// histogramQuantile approximates a quantile given a histogram with the
// cumulative distribution of the dataset.
//
// Each input table represents a single histogram.
// Input tables must have two columns: a count column (`_value`) and an upper bound
// column (`le`). Neither column can be part of the group key.
//
// The count is the number of values that are less than or equal to the upper bound value (`le`).
// Input tables can have an unlimited number of records; each record represents an entry in the histogram.
// The counts must be monotonically increasing when sorted by upper bound (`le`).
// If any values in the `_value` or `le` columns are `null`, the function returns an error.
//
// Linear interpolation between the two closest bounds is used to compute the quantile.
// If the either of the bounds used in interpolation are infinite,
// then the other finite bound is used and no interpolation is performed.
//
// The output table has the same group key as the input table.
// The function returns the value of the specified quantile from the histogram in the
// `_value` column and drops all columns not part of the group key.
//
// ## Parameters
// - quantile: Quantile to compute (`[0.0 - 1.0]`).
// - minValue: Assumed minimum value of the dataset. Default is `0.0`.
//
//   When the quantile falls below the lowest upper bound, the function
//   interpolates values between `minValue` and the lowest upper bound.
//   If `minValue` is equal to negative infinity, the lowest upper bound is used.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Compute the 90th percentile of a histogram
// ```
// # import "array"
// import "experimental"
//
// # histogramData = array.from(
// #     rows: [
// #         {_field: "example_histogram", _time: 2021-01-01T00:00:00Z, _value: 6873.0, le: 0.005},
// #         {_field: "example_histogram", _time: 2021-01-01T00:00:00Z, _value: 9445.0, le: 0.01},
// #         {_field: "example_histogram", _time: 2021-01-01T00:00:00Z, _value: 9487.0, le: 0.025},
// #         {_field: "example_histogram", _time: 2021-01-01T00:00:00Z, _value: 9487.0, le: 0.05},
// #         {_field: "example_histogram", _time: 2021-01-01T00:00:00Z, _value: 9487.0, le: 0.1},
// #         {_field: "example_histogram", _time: 2021-01-01T00:00:00Z, _value: 9487.0, le: 0.25},
// #         {_field: "example_histogram", _time: 2021-01-01T00:00:00Z, _value: 9487.0, le: 0.5},
// #         {_field: "example_histogram", _time: 2021-01-01T00:00:00Z, _value: 9487.0, le: 1.0},
// #         {_field: "example_histogram", _time: 2021-01-01T00:00:00Z, _value: 9487.0, le: 2.5},
// #         {_field: "example_histogram", _time: 2021-01-01T00:00:00Z, _value: 9487.0, le: 5.0},
// #         {_field: "example_histogram", _time: 2021-01-01T00:00:00Z, _value: 9487.0, le: 10.0},
// #         {_field: "example_histogram", _time: 2021-01-01T00:00:00Z, _value: 9487.0, le: float(v: "+Inf")},
// #     ],
// # )
// #     |> group(columns: ["_field"])
// #
// < histogramData
// >    |> experimental.histogramQuantile(quantile: 0.9)
// ```
//
// ## Metadata
// introduced: 0.107.0
// tags: transformations,aggregates
//
builtin histogramQuantile : (
        <-tables: stream[{T with _value: float, le: float}],
        ?quantile: float,
        ?minValue: float,
    ) => stream[{T with _value: float}]

// mean computes the mean or average of non-null values in the `_value` column
// of each input table.
//
// Output tables contain a single row the with the calculated mean in the `_value` column.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Calculate the average value of input tables
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.float()
// >     |> experimental.mean()
// ```
//
// ## Metadata
// introduced: 0.107.0
// tags: transformations,aggregates
//
builtin mean : (<-tables: stream[{T with _value: float}]) => stream[{T with _value: float}]

// mode computes the mode or value that occurs most often in the `_value` column
// in each input table.
//
// `experimental.mode` only considers non-null values.
// If there are multiple modes, it returns all modes in a sorted table.
// If there is no mode, it returns _null_.
//
// #### Supported types
// - string
// - float
// - int
// - uint
// - bool
// - time
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Compute the mode of input tables
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.int()
// >     |> experimental.mode()
// ```
//
// ## Metadata
// introduced: 0.107.0
// tags: transformations,aggregates
//
builtin mode : (<-tables: stream[{T with _value: A}]) => stream[{T with _value: A}]

// quantile returns non-null records with values in the `_value` column that
// fall within the specified quantile or represent the specified quantile.
//
// The `_value` column must contain float values.
//
// ## Computation methods and behavior
// `experimental.quantile()` behaves like an **aggregate function** or a
// **selector function** depending on the `method` parameter.
// The following computation methods are available:
//
// ##### estimate_tdigest
// An aggregate method that uses a [t-digest data structure](https://github.com/tdunning/t-digest)
// to compute an accurate quantile estimate on large data sources.
// When used, `experimental.quantile()` outputs non-null records with values
// that fall within the specified quantile.
//
// ##### exact_mean
// An aggregate method that takes the average of the two points closest to the quantile value.
// When used, `experimental.quantile()` outputs non-null records with values
// that fall within the specified quantile.
//
// ##### exact_selector
// A selector method that returns the data point for which at least `q` points are less than.
// When used, `experimental.quantile()` outputs the non-null record with the
// value that represents the specified quantile.
//
// ## Parameters
// - q: Quantile to compute (`[0 - 1]`).
// - method: Computation method. Default is `estimate_tdigest`.
//
//   **Supported methods**:
//   - estimate_tdigest
//   - exact_mean
//   - exact_selector
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
// ### Return values in the 50th percentile of each input table
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.float()
// >     |> experimental.quantile(q: 0.5)
// ```
//
// ### Return a value representing the 50th percentile of each input table
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.float()
//     |> experimental.quantile(
//           q: 0.5,
//           method: "exact_selector",
// >     )
// ```
//
// ## Metadata
// introduced: 0.107.0
// tags: transformations,aggregates,selectors
//
builtin quantile : (
        <-tables: stream[{T with _value: float}],
        q: float,
        ?compression: float,
        ?method: string,
    ) => stream[{T with _value: float}]

// skew returns the skew of non-null values in the `_value` column for each
// input table as a float.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Return the skew of input tables
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.float()
// >     |> experimental.skew()
// ```
//
// ## Metadata
// introduced: 0.107.0
// tags: transformations,aggregates
//
builtin skew : (<-tables: stream[{T with _value: float}]) => stream[{T with _value: float}]

// spread returns the difference between the minimum and maximum values in the
// `_value` column for each input table.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Return the difference between minimum and maximum values
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.float()
// >     |> experimental.spread()
// ```
//
// ## Metadata
// introduced: 0.107.0
// tags: transformations,aggregates
//
builtin spread : (<-tables: stream[{T with _value: A}]) => stream[{T with _value: A}]
    where
    A: Numeric

// stddev returns the standard deviation of non-null values in the `_value`
// column for each input table.
//
// ## Standard deviation modes
// The following modes are avaialable when calculating the standard deviation of data.
//
// ##### sample
// Calculate the sample standard deviation where the data is considered to be
// part of a larger population.
//
// ##### population
// Calculate the population standard deviation where the data is considered a
// population of its own.
//
// ## Parameters
// - mode: Standard deviation mode or type of standard deviation to calculate.
//   Default is `sample`.
//
//   **Available options**:
//   - sample
//   - population
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Return the standard deviation in input tables
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.float()
// >     |> experimental.stddev()
// ```
//
// ## Metadata
// introduced: 0.107.0
// tags: transformations,aggregates
//
builtin stddev : (
        <-tables: stream[{T with _value: float}],
        ?mode: string,
    ) => stream[{T with _value: float}]

// sum returns the sum of non-null values in the `_value` column for each input table.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Return the sum of each input table
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.int()
// >     |> experimental.sum()
// ```
//
// ## Metadata
// introduced: 0.107.0
// tags: transformations,aggregates
//
builtin sum : (<-tables: stream[{T with _value: A}]) => stream[{T with _value: A}] where A: Numeric

// kaufmansAMA calculates the Kaufman's Adaptive Moving Average (KAMA) of input
// tables using the `_value` column in each table.
//
// Kaufman's Adaptive Moving Average is a trend-following indicator designed to
// account for market noise or volatility.
//
// ## Parameters
// - n: Period or number of points to use in the calculation.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Calculate the KAMA of input tables
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.int()
// >     |> experimental.kaufmansAMA(n: 3)
// ```
//
// ## Metadata
// introduced: 0.107.0
// tags: transformations
//
builtin kaufmansAMA : (
        <-tables: stream[{T with _value: A}],
        n: int,
    ) => stream[{T with _value: float}]
    where
    A: Numeric

// distinct returns unique values from the `_value` column.
//
// The `_value` of each output record is set to a distinct value in the specified column.
// `null` is considered a distinct value.
//
// `experimental.distinct()` drops all columns **not** in the group key and
// drops empty tables.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Return distinct values from each input table
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> experimental.distinct()
// ```
//
// ## Metadata
// introduced: 0.112.0
// tags: transformations,selectors
//
builtin distinct : (<-tables: stream[{T with _value: A}]) => stream[{T with _value: A}]

// fill replaces all null values in the `_value` column with a non-null value.
//
// ## Parameters
// - value: Value to replace null values with.
//   Data type must match the type of the `_value` column.
// - usePrevious: Replace null values with the value of the previous non-null row.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Fill null values with a specified non-null value
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> experimental.fill(value: 0)
// ```
//
// ### Fill null values with the previous non-null value
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> experimental.fill(usePrevious: true)
// ```
//
// ## Metadata
// introduced: 0.112.0
// tags: transformations
//
builtin fill : (
        <-tables: stream[{T with _value: A}],
        ?value: A,
        ?usePrevious: bool,
    ) => stream[{T with _value: A}]

// first returns the first record with a non-null value in the `_value` column
// for each input table.
//
// `experimental.first()` drops empty tables.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Return the first non-null value in each input table
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> experimental.first()
// ```
//
// ## Metadata
// introduced: 0.112.0
// tags: transformations,selectors
//
builtin first : (<-tables: stream[{T with _value: A}]) => stream[{T with _value: A}]

// last returns the last record with a non-null value in the `_value` column
// for each input table.
//
// `experimental.last()` drops empty tables.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Return the last non-null value in each input table
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> experimental.last()
// ```
//
// ## Metadata
// introduced: 0.112.0
// tags: transformations,selectors
//
builtin last : (<-tables: stream[{T with _value: A}]) => stream[{T with _value: A}]

// max returns the record with the highest value in the `_value` column for each
// input table.
//
// // `experimental.max()` drops empty tables.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Return the row with the maximum value in each input table
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.int()
// >     |> experimental.max()
// ```
//
// ## Metadata
// introduced: 0.112.0
// tags: transformations,selectors
//
builtin max : (<-tables: stream[{T with _value: A}]) => stream[{T with _value: A}]

// min returns the record with the lowest value in the `_value` column for each
// input table.
//
// `experimental.min()` drops empty tables.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Return the row with the lowest value in each input table
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.int()
// >     |> experimental.min()
// ```
//
// ## Metadata
// introduced: 0.112.0
// tags: transformations,selectors
//
builtin min : (<-tables: stream[{T with _value: A}]) => stream[{T with _value: A}]

// unique returns all records containing unique values in the `_value` column.
//
// `null` is considered a unique value.
//
// #### Function behavior
// - Outputs a single table for each input table.
// - Outputs a single record for each unique value in an input table.
// - Leaves group keys, columns, and values unmodified.
// - Drops emtpy tables.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Return rows with unique values in each input table
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.int(includeNull: true)
// >     |> experimental.unique()
// ```
//
// ## Metadata
// introduced: 0.112.0
// tags: transformations,selectors
//
builtin unique : (<-tables: stream[{T with _value: A}]) => stream[{T with _value: A}]

// histogram approximates the cumulative distribution of a dataset by counting
// data frequencies for a list of bins.
//
// A bin is defined by an upper bound where all data points that are less than
// or equal to the bound are counted in the bin.
// Bin counts are cumulative.
//
// #### Function behavior
// - Outputs a single table for each input table.
// - Each output table represents a unique histogram.
// - Output tables have the same group key as the corresponding input table.
// - Drops columns that are not part of the group key.
// - Adds an `le` column to store upper bound values.
// - Stores bin counts in the `_value` column.
//
// ## Parameters
// - bins: List of upper bounds to use when computing histogram frequencies,
//   including the maximum value of the data set.
//
//   This value can be set to positive infinity (`float(v: "+Inf")`) if no maximum is known.
//
//   ##### Bin helper functions
//   The following helper functions can be used to generated bins.
//
//   - `linearBins()`
//   - `logarithmicBins()`
//
// - normalize: Convert count values into frequency values between 0 and 1.
//   Default is `false`.
//
//   **Note**: Normalized histograms cannot be aggregated by summing their counts.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Create a histgram from input data
// ```
// import "experimental"
// import "sampledata"
//
// < sampledata.float()
//     |> experimental.histogram(bins: [
//         0.0,
//         5.0,
//         10.0,
//         15.0,
//         20.0,
// >     ])
// ```
//
// ## Metadata
// introduced: 0.112.0
// tags: transformations
//
builtin histogram : (
        <-tables: stream[{T with _value: float}],
        bins: [float],
        ?normalize: bool,
    ) => stream[{T with _value: float, le: float}]

// preview limits the number of rows and tables in the stream.
//
// Included group keys are not deterministic and depends on the order
// that the engine sends them.
//
// ## Parameters
// - nrows: Maximum number of rows per table to return. Default is `5`.
//
// - ntables: Maximum number of tables to return.
//   Default is `5`.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Preview data output
// ```
// import "experimental"
// import "sampledata"
//
// sampledata.int()
//     |> experimental.preview()
// ```
//
// ## Metadata
// introduced: 0.167.0
// tags: transformations
//
builtin preview : (<-tables: stream[A], ?nrows: int, ?ntables: int) => stream[A] where A: Record

// unpivot creates `_field` and `_value` columns pairs using all columns (other than `_time`)
// _not_ in the group key.
// The `_field` column contains the original column label and the `_value` column
// contains the original column value.
//
// The output stream retains the group key and all group key columns of the input stream.
// `_field` is added to the output group key.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
// - otherColumns: List of column names that are not in the group key but are also not field columns. Default is `["_time"]`.
//
// ## Examples
//
// ### Unpivot data into _field and _value columns
//
// ```
// # import "array"
// import "experimental"
//
// # data =
// #     array.from(
// #         rows: [
// #             {_time: 2022-01-01T00:00:00Z, location: "New York", temp: 50.1, hum: 99.2},
// #             {_time: 2022-01-02T00:00:00Z, location: "New York", temp: 55.8, hum: 97.7},
// #             {_time: 2022-01-01T00:00:00Z, location: "Denver", temp: 10.2, hum: 81.5},
// #             {_time: 2022-01-02T00:00:00Z, location: "Denver", temp: 12.4, hum: 41.3},
// #         ],
// #     )
// #         |> group(columns: ["location"])
// #
// < data
// >     |> experimental.unpivot()
// ```
//
// ## Metadata
// introduced: 0.172.0
// tags: transformations
builtin unpivot : (
        <-tables: stream[{A with _time: time}],
        ?otherColumns: [string],
    ) => stream[{B with _field: string, _value: C}]
    where
    A: Record,
    B: Record

// catch calls a function and returns any error as a string value.
// If the function does not error the returned value is made into a string and returned.
//
// ## Parameters
// - fn: Function to call.
//
// ## Examples
//
// ### Catch an explicit error
// ```no_run
// import "experimental"
//
// experimental.catch(fn: () => die(msg:"error message")) // Returns "error message"
// ```
//
// ## Metadata
// introduced: 0.174.0
builtin catch : (fn: () => A) => {value: A, code: uint, msg: string}

// diff takes two table streams as input and produces a diff.
//
// `experimental.diff()` compares tables with the same group key.
// If compared tables are different, the function returns a table for that group key with one or more rows.
// If there are no differences, the function does not return a table for that group key.
//
// **Note:** `experimental.diff()` cannot tell the difference between an empty table and a non-existent table.
//
// **Important:** The output format of the diff is not considered stable and the algorithm used to produce the diff may change.
// The only guarantees are those mentioned above.
//
// ## Parameters
// - want: Input stream for the `-` side of the diff.
// - got: Input stream for the `+` side of the diff.
//
// ## Examples
//
// ### Output a diff between two streams of tables
// ```
// import "sampledata"
// import "experimental"
//
// want = sampledata.int()
// got = sampledata.int()
//     |> map(fn: (r) => ({r with _value: if r._value > 15 then r._value + 1 else r._value }))
//
// < experimental.diff(got: got, want: want)
// ```
//
// ### Return a diff between a stream of tables and the expected output
// ```no_run
// import "experimental"
//
// want = from(bucket: "backup-example-bucket") |> range(start: -5m)
//
// from(bucket: "example-bucket")
//     |> range(start: -5m)
//     |> experimental.diff(want: want)
// ```
//
// ## Metadata
// introduced: 0.175.0
//
builtin diff : (<-got: stream[A], want: stream[A]) => stream[{A with _diff: string}]
