package tickscript

import "experimental"
import "influxdata/influxdb"
import "influxdata/influxdb/monitor"
import "influxdata/influxdb/schema"
import "universe"

// alert
alert = (
    check,
    id=(r)=>"${r._check_id}",
    message=(r)=>"Check: ${r._check_name} is: ${r._level}",
    details=(r)=>"",
    crit=(r) => false,
    warn=(r) => false,
    info=(r) => false,
    ok=(r) => true,
    tables=<-) =>
  tables
    |> drop(fn: (column) => column =~ /_start.*/ or column =~ /_stop.*/)
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
// it is meant to be a convenience function, it ensures _measurement column exists and also removes some useless columns
join = (tables, on=["_time"], measurement) =>
    universe.join(tables: tables, on: on)
      |> map(fn: (r) => ({ r with _measurement: measurement }))
      |> experimental.group(columns: ["_measurement"], mode: "extend") // required by monitor.check
