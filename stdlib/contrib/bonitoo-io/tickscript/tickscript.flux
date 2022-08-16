// Package tickscript provides functions to help migrate
// Kapacitor [TICKscripts](https://docs.influxdata.com/kapacitor/latest/tick/) to Flux tasks.
//
// ## Metadata
// introduced: 0.111.0
// contributors: **GitHub**: [@alespour](https://github.com/alespour), [@bonitoo-io](https://github.com/bonitoo-io) | **InfluxDB Slack**: [@Ales Pour](https://influxdata.com/slack)
//
package tickscript


import "experimental"
import "experimental/array"
import "influxdata/influxdb"
import "influxdata/influxdb/monitor"
import "influxdata/influxdb/schema"
import "universe"

// defineCheck creates custom check data required by `alert()` and `deadman()`.
//
// ## Parameters
//
// - id: InfluxDB check ID.
// - name: InfluxDB check name.
// - type: InfluxDB check type. Default is `custom`.
//
//   **Valid values**:
//   - `threshold`
//   -`deadman`
//   -`custom`
//
// ## Examples
// ### Generate InfluxDB check data
// ```no_run
// import "contrib/bonitoo-io/tickscript"
//
// tickscript.defineCheck(
//     id: "000000000000",
//     name: "Example check name",
// )
//
// // Returns:
// // {
// //     _check_id: 000000000000,
// //     _check_name: Example check name,
// //     _type: custom,
// //     tags: {}
// // }
// ```
defineCheck = (id, name, type="custom") => {
    return {_check_id: id, _check_name: name, _type: type, tags: {}}
}

// alert identifies events of varying severity levels
// and writes them to the `statuses` measurement in the InfluxDB `_monitoring`
// system bucket.
//
// This function is comparable to
// TICKscript [`alert()`](https://docs.influxdata.com/kapacitor/v1.6/nodes/alert_node/).
//
// ## Parameters
//
// - check: InfluxDB check data.
//   See `tickscript.defineCheck()`.
// - id: Function that returns the InfluxDB check ID provided by the check record.
//   Default is `(r) => "${r._check_id}"`.
// - details: Function to return the InfluxDB check details using data from input rows.
//   Default is `(r) => ""`.
// - message: Function to return the InfluxDB check message using data from input rows.
//   Default is `(r) => "Threshold Check: ${r._check_name} is: ${r._level}"`.
// - crit: Predicate function to determine `crit` status. Default is `(r) => false`.
// - warn: Predicate function to determine `warn` status. Default is `(r) => false`.
// - info: Predicate function to determine `info` status. Default is `(r) => false`.
// - ok: Predicate function to determine `ok` status. Default is `(r) => true`.
// - topic: Check topic. Default is `""`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Store alert statuses for error counts
// ```no_run
// import "contrib/bonitoo-io/tickscript"
//
// option task = {name: "Example task", every: 1m}
//
// check = tickscript.defineCheck(
//     id: "000000000000",
//     name: "Errors",
//     type: "threshold",
// )
//
// from(bucket: "example-bucket")
//     |> range(start: -task.every)
//     |> filter(fn: (r) => r._measurement == "errors" and r._field == "value")
//     |> count()
//     |> tickscript.alert(
//         check: {check with _check_id: "task/${r.service}"},
//         message: "task/${r.service} is ${r._level} value: ${r._value}",
//         crit: (r) => r._value > 30,
//         warn: (r) => r._value > 20,
//         info: (r) => r._value > 10,
//     )
// ```
//
// ## Metadata
// tags: transformations,outputs
alert = (
        check,
        id=(r) => "${r._check_id}",
        details=(r) => "",
        message=(r) => "Threshold Check: ${r._check_name} is: ${r._level}",
        crit=(r) => false,
        warn=(r) => false,
        info=(r) => false,
        ok=(r) => true,
        topic="",
        tables=<-,
    ) =>
    {
        _addTopic =
            if topic != "" then
                (tables=<-) =>
                    tables
                        |> set(key: "_topic", value: topic)
                        |> experimental.group(mode: "extend", columns: ["_topic"])
            else
                (tables=<-) => tables

        return
            tables
                |> drop(fn: (column) => column =~ /_start.*/ or column =~ /_stop.*/)
                |> map(
                    fn: (r) =>
                        ({r with _check_id: check._check_id, _check_name: check._check_name}),
                )
                |> map(fn: (r) => ({r with id: id(r: r)}))
                |> map(fn: (r) => ({r with details: details(r: r)}))
                |> _addTopic()
                |> monitor.check(
                    crit: crit,
                    warn: warn,
                    info: info,
                    ok: ok,
                    messageFn: message,
                    data: check,
                )
    }

// deadman detects low data throughput and writes a point with a critical status to
// the InfluxDB `_monitoring` system bucket.
//
// For each input table containing a number of rows less than or equal to the specified threshold,
// the function assigns a `crit` value to the` _level` column.
//
// This function is comparable to [Kapacitor AlertNode deadman](https://docs.influxdata.com/kapacitor/latest/nodes/stream_node/#deadman).
//
// ## Parameters
//
// - check: InfluxDB check data. See `tickscript.defineCheck()`.
// - measurement: Measurement name. Should match the queried measurement.
// - threshold: Count threshold. Default is `0`.
//
//   The function assigns a `crit` status to input tables with a number of rows less than or equal to the threshold.
//
// - id: Function that returns the InfluxDB check ID provided by the check record.
//   Default is `(r) => "${r._check_id}"`.
// - message: Function that returns the InfluxDB check message using data from input rows.
//   Default is `(r) => "Deadman Check: ${r._check_name} is: " + (if r.dead then "dead" else "alive")`.
// - topic: Check topic. Default is `""`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Detect when a series stops reporting
// ```no_run
// import "contrib/bonitoo-io/tickscript"
//
// option task = {name: "Example task", every: 1m}
//
// from(bucket: "example-bucket")
//     |> range(start: -task.every)
//     |> filter(fn: (r) => r._measurement == "pulse" and r._field == "value")
//     |> tickscript.deadman(
//         check: tickscript.defineCheck(id: "000000000000", name: "task/${r.service}"),
//         measurement: "pulse",
//         threshold: 2,
//     )
//```
//
// ## Metadata
// tags: transformations
//
deadman =
    (
            check,
            measurement,
            threshold=0,
            id=(r) => "${r._check_id}",
            message=(r) =>
                "Deadman Check: ${r._check_name} is: " + (if r.dead then "dead" else "alive"),
            topic="",
            tables=<-,
        ) =>
        {
            // In order to detect empty stream (without tables), we concatenate input with dummy stream and count the result,
            // because count() returns nothing for empty stream. If the input stream is empty, then dummy stream with empty
            // table is used as input for actual threshold check in order to get 0.
            _dummy =
                array.from(rows: [{_time: 2000-01-01T00:00:00Z, _field: "unknown", _value: 0}])
                    |> set(key: "_measurement", value: measurement)
                    // required by monitor.check
                    |> experimental.group(columns: ["_measurement"], mode: "extend")
                    // input tables are expected to be pivoted already
                    |> schema.fieldsAsCols()
            _counts =
                union(tables: [_dummy, tables])
                    |> keep(columns: ["_measurement", "_time"])
                    // _measurement column is always present
                    |> duplicate(column: "_measurement", as: "__value__")
                    |> count(column: "__value__")
                    |> findColumn(fn: (key) => key._measurement == measurement, column: "__value__")
            _tables =
                // only dummy table is in the concatenated stream
                if _counts[0] == 1 then
                    _dummy
                        |> drop(columns: ["unknown"])
                        // need empty table
                        |> limit(n: 0)
                else
                    tables

            return
                _tables
                    // _measurement column is always present
                    |> duplicate(column: "_measurement", as: "__value__")
                    |> count(column: "__value__")
                    // recreate _time column after aggregation
                    |> map(fn: (r) => ({r with _time: now()}))
                    // same tag that monitor.deadman() adds
                    |> map(fn: (r) => ({r with dead: r.__value__ <= threshold}))
                    // drop dummy field
                    |> drop(columns: ["__value__"])
                    |> alert(
                        check: check,
                        id: id,
                        message: message,
                        crit: (r) => r.dead,
                        topic: topic,
                    )
        }

// select changes a column’s name and optionally applies an aggregate or selector
// function to values in the column.
//
// ## TICKscript helper function
//
// `tickscript.select()` is a helper function meant to replicate TICKscript operations like the following:
//
// ```
// // Rename
// query("SELECT x AS y")
//
// // Aggregate and rename
// query("SELECT f(x) AS y")
// ```
//
// ## Parameters
//
// - column: Column to operate on. Default is `_value`.
// - fn: Aggregate or selector function to apply.
// - as: New column name.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Change the name of the value column
// ```
// import "contrib/bonitoo-io/tickscript"
// import "sampledata"
//
// < sampledata.int()
// >     |> tickscript.select(as: "example-name")
// ```
//
// ### Change the name of the value column and apply an aggregate function
// ```
// import "contrib/bonitoo-io/tickscript"
// import "sampledata"
//
// < sampledata.int()
//     |> tickscript.select(
//         as: "sum",
//         fn: sum,
// >     )
// ```
//
// ### Change the name of the value column and apply a selector function
// ```
// import "contrib/bonitoo-io/tickscript"
// import "sampledata"
//
// < sampledata.int()
//     |> tickscript.select(
//         as: "max",
//         fn: max,
// >     )
// ```
//
// ## Metadata
// tags: transformations
//
select = (column="_value", fn=(column, tables=<-) => tables, as, tables=<-) => {
    _column = column
    _as = as

    return
        tables
            |> fn(column: _column)
            |> rename(fn: (column) => if column == _column then _as else column)
}

// selectWindow changes a column’s name, windows rows by time, and then applies an
// aggregate or selector function the specified column for each window of time.
//
// ## TICKscript helper function
// `tickscript.selectWindow` is a helper function meant to replicate TICKscript operations like the following:
//
// ```
// // Rename, window, and aggregate
// query("SELECT f(x) AS y")
//   .groupBy(time(t), ...)
// ```
//
// ## Parameters
//
// - column: Column to operate on. Default is _value.
// - fn: Aggregate or selector function to apply.
// - as: New column name.
// - every: Duration of windows.
// - defaultValue: Default fill value for null values in column.
//   Must be the same data type as column.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Change the name of, window, and then aggregate the value column
// ```
// import "contrib/bonitoo-io/tickscript"
// # import "sampledata"
//
// # data = sampledata.int()
// #    |> range(start: sampledata.start, stop: sampledata.stop)
//
// < data
//     |> tickscript.selectWindow(
//         fn: sum,
//         as: "example-name",
//         every: 1h,
//         defaultValue: 0,
// >     )
// ```
//
// ## Metadata
// tags: transformations
selectWindow = (
        column="_value",
        fn,
        as,
        every,
        defaultValue,
        tables=<-,
    ) =>
    {
        _column = column
        _as = as

        return
            tables
                |> aggregateWindow(every: every, fn: fn, column: _column, createEmpty: true)
                |> fill(column: _column, value: defaultValue)
                |> rename(fn: (column) => if column == _column then _as else column)
    }

// compute is an alias for `tickscript.select()` that changes a column’s name and
// optionally applies an aggregate or selector function.
//
// ## Parameters
//
// - as: New column name.
// - column: Column to operate on. Default is `_value`.
// - fn: Aggregate or selector function to apply.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Metadata
// tags: transformations
compute = select

// groupBy groups results by the `_measurement` column and other specified columns.
//
// This function is comparable to [Kapacitor QueryNode .groupBy](https://docs.influxdata.com/kapacitor/latest/nodes/query_node/#groupby).
//
// **Note**: To group by time intervals, use `window()` or `tickscript.selectWindow()`.
//
// ## Parameters
// - columns: List of columns to group by.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Group by host and region
// ```
// # import "array"
// import "contrib/bonitoo-io/tickscript"
//
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, _measurement: "m", host: "h1", region: "east", _field: "foo", _value: 1.2},
// #         {_time: 2021-01-01T00:01:00Z, _measurement: "m", host: "h1", region: "east", _field: "foo", _value: 3.4},
// #         {_time: 2021-01-01T00:00:00Z, _measurement: "m", host: "h2", region: "east", _field: "foo", _value: 2.3},
// #         {_time: 2021-01-01T00:01:00Z, _measurement: "m", host: "h2", region: "east", _field: "foo", _value: 5.6},
// #         {_time: 2021-01-01T00:00:00Z, _measurement: "m", host: "h3", region: "west", _field: "foo", _value: 1.2},
// #         {_time: 2021-01-01T00:01:00Z, _measurement: "m", host: "h3", region: "west", _field: "foo", _value: 3.4},
// #         {_time: 2021-01-01T00:00:00Z, _measurement: "m", host: "h4", region: "west", _field: "foo", _value: 2.3},
// #         {_time: 2021-01-01T00:01:00Z, _measurement: "m", host: "h4", region: "west", _field: "foo", _value: 5.6},
// #     ],
// # )
//
// < data
// >     |> tickscript.groupBy(columns: ["host", "region"])
// ```
//
// ## Metadata
// tags: transformations
groupBy = (columns, tables=<-) =>
    tables
        |> group(columns: columns)
        // required by monitor.check
        |> experimental.group(columns: ["_measurement"], mode: "extend")

// join merges two input streams into a single output stream
// based on specified columns with equal values and appends a new measurement name.
//
// This function is comparable to [Kapacitor JoinNode](https://docs.influxdata.com/kapacitor/latest/nodes/join_node/).
//
// ## Parameters
//
// - tables: Map of two streams to join.
// - on: List of columns to join on. Default is `["_time"]`.
// - measurement: Measurement name to use in results.
//
// ## Examples
// ### Join two streams of data
//
// ```
// import "array"
// import "contrib/bonitoo-io/tickscript"
//
// metrics = array.from(
//     rows: [
//         {_time: 2021-01-01T00:00:00Z, host: "host1", _value: 1.2},
//         {_time: 2021-01-01T01:00:00Z, host: "host1", _value: 0.8},
//         {_time: 2021-01-01T02:00:00Z, host: "host1", _value: 3.2},
//         {_time: 2021-01-01T00:00:00Z, host: "host2", _value: 8.4},
//         {_time: 2021-01-01T01:00:00Z, host: "host2", _value: 7.3},
//         {_time: 2021-01-01T02:00:00Z, host: "host2", _value: 7.9},
//     ],
// )
//     |> group(columns: ["host"])
//
// states = array.from(
//     rows: [
//         {_time: 2021-01-01T00:00:00Z, host: "host1", _value: "dead"},
//         {_time: 2021-01-01T01:00:00Z, host: "host1", _value: "dead"},
//         {_time: 2021-01-01T02:00:00Z, host: "host1", _value: "alive"},
//         {_time: 2021-01-01T00:00:00Z, host: "host2", _value: "alive"},
//         {_time: 2021-01-01T01:00:00Z, host: "host2", _value: "alive"},
//         {_time: 2021-01-01T02:00:00Z, host: "host2", _value: "alive"},
//     ],
// )
//     |> group(columns: ["host"])
//
// tickscript.join(
//     tables: {metric: metrics, state: states},
//     on: ["_time", "host"],
//     measurement: "example-m",
// > )
//```
//
// ## Metadata
// tags: transformations
join = (tables, on=["_time"], measurement) =>
    universe.join(tables: tables, on: on)
        |> map(fn: (r) => ({r with _measurement: measurement}))
        // required by monitor.check
        |> experimental.group(columns: ["_measurement"], mode: "extend")
