package tickscript

import "experimental"
import "experimental/array"
import "influxdata/influxdb"
import "influxdata/influxdb/monitor"
import "influxdata/influxdb/schema"
import "universe"

// alert
alert = (
    check,
    id=(r)=>"${r._check_id}",
    details=(r)=>"",
    message=(r)=>"Threshold Check: ${r._check_name} is: ${r._level}",
    crit=(r) => false,
    warn=(r) => false,
    info=(r) => false,
    ok=(r) => true,
    tables=<-) =>
  tables
    |> drop(fn: (column) => column =~ /_start.*/ or column =~ /_stop.*/)
    |> map(fn: (r) => ({r with
        _check_id: check._check_id,
        _check_name: check._check_name,
    }))
    |> map(fn: (r) => ({ r with id: id(r: r) }))
    |> map(fn: (r) => ({ r with details: details(r: r) }))
    |> monitor.check(
        crit: crit,
        warn: warn,
        info: info,
        ok: ok,
        messageFn: message,
        data: check
    )

// deadman
deadman = (
    check,
    measurement, threshold=0,
    id=(r)=>"${r._check_id}",
    message=(r)=>"Deadman Check: ${r._check_name} is: " + (if r.dead then "dead" else "alive"),
    tables=<-) => {
  _dummy = array.from(rows: [{_time: 2000-01-01T00:00:00Z, _field: "unknown", _value: 0}])
    |> map(fn: (r) => ({ r with _measurement: measurement }))
    |> experimental.group(columns: ["_measurement"], mode: "extend") // required by monitor.check
  _counts = union(tables: [_dummy, tables])
    |> keep(columns: ["_time"])
    |> map(fn: (r) => ({ r with __value__: 0 }))
    |> count(column: "__value__")
    |> findColumn(fn: (key) => true, column: "__value__")
  _tables =
    if _counts[0] == 1 then // only dummy record is in the unioned stream
      _dummy
        |> limit(n: 0) // need empty table
    else
      tables

  return _tables
    |> duplicate(column: "_measurement", as: "__value__")        // _measurement column is always present
    |> count(column: "__value__")
    |> map(fn: (r) => ({r with _time: now()}))                   // recreate _time column after aggregation
    |> map(fn: (r) => ({r with dead: r.__value__ <= threshold})) // same tag that monitor.deadman() adds
    |> drop(columns: ["__value__"])
    |> alert(
      check: check,
      id: id,
      message: message,
      crit: (r) => r.dead
    )
}

// routes alerts to topic
topic = (name, tables=<-) =>
  tables
    |> set(key: "_topic", value: name )
    |> experimental.group(mode: "extend", columns: ["_topic"])
    |> monitor.write()

//
// TICKscript -> Flux helper functions
//

// selects a column and optionally computes aggregated value
// it is meant to be a convenience function to be used for:
//
//   query("SELECT x AS y")
//   query("SELECT f(x) AS y") without time grouping
//
select = (column="_value", fn=(column, tables=<-) => tables, as, tables=<-) => {
  _column = column
  _as = as
  return
    tables
      |> fn(column: _column)
      |> rename(fn: (column) => if column == _column then _as else column)
}

// selects column with time grouping and computes aggregated values
// it is meant to be a convenience function to be used for:
//
//   query("SELECT f(x) AS y")
//     .groupBy(time(t), ...)
//
selectWindow = (column="_value", fn, as, every, defaultValue, tables=<-) => {
  _column = column
  _as = as
  return
    tables
      |> aggregateWindow(every: every, fn: fn, column: _column, createEmpty: true)
      |> fill(column: _column, value: defaultValue)
      |> rename(fn: (column) => if column == _column then _as else column)
}

// computes aggregated value of tha data
// it is meant to be a convenience function to be used for:
//
//   |median('x)'
//      .as(y)
//
compute = select

// groups by specified columns
// it is meant to be a convenience function, it adds _measurement column which is required by monitor.check() (in alert())
groupBy = (columns, tables=<-) =>
  tables
    |> group(columns: columns)
    |> experimental.group(columns: ["_measurement"], mode:"extend") // required by monitor.check

// joins the streams using standard join()
// it is meant to be a convenience function, it ensures _measurement column exists and is in the group key
join = (tables, on=["_time"], measurement) =>
    universe.join(tables: tables, on: on)
      |> map(fn: (r) => ({ r with _measurement: measurement }))
      |> experimental.group(columns: ["_measurement"], mode: "extend") // required by monitor.check
