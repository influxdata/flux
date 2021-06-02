package tickscript


import "experimental"
import "experimental/array"
import "influxdata/influxdb"
import "influxdata/influxdb/monitor"
import "influxdata/influxdb/schema"
import "universe"

// defineCheck creates custom check data required by alert() and deadman()
defineCheck = (id, name, type="custom") => {
    return {
        _check_id: id,
        _check_name: name,
        _type: type,
        tags: {},
    }
}

// alert is a helper function similar to TICKscript alert.
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
) => {
    _addTopic = if topic != "" then
        (tables=<-) => tables
            |> set(key: "_topic", value: topic)
            |> experimental.group(mode: "extend", columns: ["_topic"])
    else
        (tables=<-) => tables

    return tables
        |> drop(fn: (column) => column =~ /_start.*/ or column =~ /_stop.*/)
        |> map(
            fn: (r) => ({r with
                _check_id: check._check_id,
                _check_name: check._check_name,
            }),
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

// deadman is a helper function similar to TICKscript deadman.
deadman = (
        check,
        measurement,
        threshold=0,
        id=(r) => "${r._check_id}",
        message=(r) => "Deadman Check: ${r._check_name} is: " + (if r.dead then "dead" else "alive"),
        topic="",
        tables=<-,
) => {
    // In order to detect empty stream (without tables), we concatenate input with dummy stream and count the result,
    // because count() returns nothing for empty stream. If the input stream is empty, then dummy stream with empty
    // table is used as input for actual threshold check in order to get 0.
    _dummy = array.from(rows: [{_time: 2000-01-01T00:00:00Z, _field: "unknown", _value: 0}])
        |> set(key: "_measurement", value: measurement)
        // required by monitor.check
        |> experimental.group(columns: ["_measurement"], mode: "extend")
        // input tables are expected to be pivoted already    
        |> schema.fieldsAsCols()
    _counts = union(tables: [_dummy, tables])
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

    return _tables
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

// select selects a column and optionally computes aggregated value.
// It is meant to be a convenience function to be used for:
//
//   query("SELECT x AS y")
//   query("SELECT f(x) AS y") without time grouping
//
select = (column="_value", fn=(column, tables=<-) => tables, as, tables=<-) => {
    _column = column
    _as = as

    return tables
        |> fn(column: _column)
        |> rename(fn: (column) => if column == _column then _as else column)
}

// selectWindow selects a column with time grouping and computes aggregated values.
// It is a convenience function to be used as
//
//   query("SELECT f(x) AS y")
//     .groupBy(time(t), ...)
//
selectWindow = (
        column="_value",
        fn,
        as,
        every,
        defaultValue,
        tables=<-,
) => {
    _column = column
    _as = as

    return tables
        |> aggregateWindow(every: every, fn: fn, column: _column, createEmpty: true)
        |> fill(column: _column, value: defaultValue)
        |> rename(fn: (column) => if column == _column then _as else column)
}

// compute computes aggregated value of the input data.
// It is a convenience function to be used as
//
//   |median('x)'
//      .as(y)
//
compute = select

// groupBy groups by specified columns.
// It is a convenience function, it adds _measurement column which is required by monitor.check().
groupBy = (columns, tables=<-) => tables
    |> group(columns: columns)
    // required by monitor.check
    |> experimental.group(columns: ["_measurement"], mode: "extend")

// join merges two streams using standard join().
// It is meant a convenience function, it ensures _measurement column exists and is in the group key.
join = (tables, on=["_time"], measurement) => universe.join(tables: tables, on: on)
    |> map(fn: (r) => ({r with _measurement: measurement}))
    // required by monitor.check
    |> experimental.group(columns: ["_measurement"], mode: "extend")
