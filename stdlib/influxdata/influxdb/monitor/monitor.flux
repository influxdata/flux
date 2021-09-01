// The Flux monitor package provides tools for monitoring and alerting with InfluxDB.
package monitor


import "experimental"
import "influxdata/influxdb/v1"
import "influxdata/influxdb"

bucket = "_monitoring"

// Write persists the check statuses
option write = (tables=<-) => tables |> experimental.to(bucket: bucket)

// Log records notification events
option log = (tables=<-) => tables |> experimental.to(bucket: bucket)

// logs retrieves notification events stored in the notifications measurement in the _monitoring bucket.
//
// ## Parameters
// - `start` is the earliest time to include in results
//
//      Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//      For example, -1h, 2019-08-28T22:00:00Z, or 1567029600. Durations are relative to now().
//
// - `stop` is the latest time to include in results
//
//      Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//      For example, -1h, 2019-08-28T22:00:00Z, or 1567029600. Durations are relative to now().
//
// - `fn` is a single argument predicate function that evaluates true or false.
//
//      Records or rows (r) that evaluate to true are included in output tables.
//      Records that evaluate to null or false are not included in output tables.
//
// ## Query notification events from the last hour
// ```
// import "influxdata/influxdb/monitor"
//
// monitor.logs(start: -2h, fn: (r) => true)
// ```
//
logs = (start, stop=now(), fn) => influxdb.from(bucket: bucket)
    |> range(start: start, stop: stop)
    |> filter(fn: (r) => r._measurement == "notifications")
    |> filter(fn: fn)
    |> v1.fieldsAsCols()

// from retrieves check statuses stored in the statuses measurement in the _monitoring bucket.
//
// ## Parameters
// - `start` is the earliest time to include in results
//
//      Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//      For example, -1h, 2019-08-28T22:00:00Z, or 1567029600. Durations are relative to now().
//
// - `stop` is the latest time to include in results
//
//      Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//      For example, -1h, 2019-08-28T22:00:00Z, or 1567029600. Durations are relative to now().
//
// - `fn` is a single argument predicate function that evaluates true or false.
//
//      Records or rows (r) that evaluate to true are included in output tables.
//      Records that evaluate to null or false are not included in output tables.
//
// ## View critical check statuses from the last hour
// ```
// import "influxdata/influxdb/monitor"
//
// monitor.from(
//  start: -1h,
//  fn: (r) => r._level == "crit"
// )
// ```
//
from = (start, stop=now(), fn=(r) => true) => influxdb.from(bucket: bucket)
    |> range(start: start, stop: stop)
    |> filter(fn: (r) => r._measurement == "statuses")
    |> filter(fn: fn)
    |> v1.fieldsAsCols()

// levels describing the result of a check
levelOK = "ok"
levelInfo = "info"
levelWarn = "warn"
levelCrit = "crit"
levelUnknown = "unknown"
_stateChanges = (fromLevel="any", toLevel="any", tables=<-) => {
    toLevelFilter = if toLevel == "any" then
        (r) => r._level != fromLevel and exists r._level
    else
        (r) => r._level == toLevel
    fromLevelFilter = if fromLevel == "any" then
        (r) => r._level != toLevel and exists r._level
    else
        (r) => r._level == fromLevel

    return tables
        |> map(
            fn: (r) => ({r with
                level_value: if toLevelFilter(r: r) then
                    1
                else if fromLevelFilter(r: r) then
                    0
                else
                    -10,
            }),
        )
        |> duplicate(column: "_level", as: "____temp_level____")
        |> drop(columns: ["_level"])
        |> rename(columns: {"____temp_level____": "_level"})
        |> sort(columns: ["_source_timestamp"], desc: false)
        |> difference(columns: ["level_value"])
        |> filter(fn: (r) => r.level_value == 1)
        |> drop(columns: ["level_value"])
        |> experimental.group(mode: "extend", columns: ["_level"])
}

// Notify will call the endpoint and log the results.
notify = (tables=<-, endpoint, data) => tables
    |> experimental.set(o: data)
    |> experimental.group(mode: "extend", columns: experimental.objectKeys(o: data))
    |> map(
        fn: (r) => ({r with
            _measurement: "notifications",
            _status_timestamp: int(v: r._time),
            _time: now(),
        }),
    )
    |> endpoint()
    |> experimental.group(mode: "extend", columns: ["_sent"])
    |> log()

// stateChangesOnly takes a stream of tables that contains a _level column
// and returns a stream of tables where each record represents a state change.
//
// ## Return records representing state changes
// ```
// import "influxdata/influxdb/monitor"
//
// monitor.from(start: -1h)
//  |> monitor.stateChangesOnly()
//
// ```
//
stateChangesOnly = (tables=<-) => {
    return tables
        |> map(
            fn: (r) => ({r with
                level_value: if r._level == levelCrit then
                    4
                else if r._level == levelWarn then
                    3
                else if r._level == levelInfo then
                    2
                else if r._level == levelOK then
                    1
                else
                    0,
            }),
        )
        |> duplicate(column: "_level", as: "____temp_level____")
        |> drop(columns: ["_level"])
        |> rename(columns: {"____temp_level____": "_level"})
        |> sort(columns: ["_source_timestamp"], desc: false)
        |> difference(columns: ["level_value"])
        |> filter(fn: (r) => r.level_value != 0)
        |> drop(columns: ["level_value"])
        |> experimental.group(mode: "extend", columns: ["_level"])
}

// stateChanges detects state changes in a stream of data with a _level column and outputs records that change from fromLevel to toLevel.
//
// ## Parameters
// - `fromLevel` is the level to detect a change from. Defaults to "any".
// - `toLevel` is the level to detect a change to. The function output records that change to this level
//
// ## Detect when the state changes to critical
// ```
// monitor.from(start: -1h)
//    |> monitor.stateChanges(toLevel: "crit")
// ```
//
stateChanges = (fromLevel="any", toLevel="any", tables=<-) => {
    return if fromLevel == "any" and toLevel == "any" then
        tables |> stateChangesOnly()
    else
        tables |> _stateChanges(fromLevel: fromLevel, toLevel: toLevel)
}

// deadman detects when a group stops reporting data.
// It takes a stream of tables and reports if groups have been observed since time t.
//
//      monitor.deadman() retains the most recent row from each input table and adds a dead column.
//      If a record appears after time t, monitor.deadman() sets dead to false. Otherwise, dead is set to true.
//
// ## Parameters
// - `t` is the time threshold for the deadman check.
//
// ## Detect if a host hasn’t reported in the last five minutes
// ```
// import "influxdata/influxdb/monitor"
// import "experimental"
//
// from(bucket: "example-bucket")
//   |> range(start: -10m)
//   |> group(columns: ["host"])
//   |> monitor.deadman(t: experimental.subDuration(d: 5m, from: now() ))
// ```
//
deadman = (t, tables=<-) => tables
    |> max(column: "_time")
    |> map(fn: (r) => ({r with dead: r._time < t}))

// check checks input data and assigns a level (ok, info, warn, or crit) to each row based on predicate functions.
//
//      monitor.check() stores statuses in the _level column and writes results to the statuses measurement in the _monitoring bucket.
//
// ## Parameters
// - `crit` is the predicate function that determines crit status. Default is (r) => false.
// - `warn` is the predicate function that determines warn status. Default is (r) => false.
// - `info` is the predicate function that determines info status. Default is (r) => false.
// - `ok` is the predicate function that determines ok status. Default is (r) => false.
// - `messagefn` is the predicate function that constructs a message to append to each row. The message is stored in the _message column.
// - `data` is meta data used to identify this check.
//
// ## Monitor disk usage
// ```
// import "influxdata/influxdb/monitor"
//
// from(bucket: "telegraf")
//   |> range(start: -1h)
//   |> filter(fn: (r) =>
//       r._measurement == "disk" and
//       r._field == "used_percent"
//   )
//   |> group(columns: ["_measurement"])
//   |> monitor.check(
//     crit: (r) => r._value > 90.0,
//     warn: (r) => r._value > 80.0,
//     info: (r) => r._value > 70.0,
//     ok:   (r) => r._value <= 60.0,
//     messageFn: (r) =>
//       if r._level == "crit" then "Critical alert!! Disk usage is at ${r._value}%!"
//       else if r._level == "warn" then "Warning! Disk usage is at ${r._value}%."
//       else if r._level == "info" then "Disk usage is at ${r._value}%."
//       else "Things are looking good.",
//     data: {
//       _check_name: "Disk Utilization (Used Percentage)",
//       _check_id: "disk_used_percent",
//       _type: "threshold",
//       tags: {}
//     }
//   )
// ```
//
check = (
        tables=<-,
        data,
        messageFn,
        crit=(r) => false,
        warn=(r) => false,
        info=(r) => false,
        ok=(r) => true,
) => tables
    |> experimental.set(o: data.tags)
    |> experimental.group(mode: "extend", columns: experimental.objectKeys(o: data.tags))
    |> map(
        fn: (r) => ({r with
            _measurement: "statuses",
            _source_measurement: r._measurement,
            _type: data._type,
            _check_id: data._check_id,
            _check_name: data._check_name,
            _level: if crit(r: r) then
                levelCrit
            else if warn(r: r) then
                levelWarn
            else if info(r: r) then
                levelInfo
            else if ok(r: r) then
                levelOK
            else
                levelUnknown,
            _source_timestamp: int(v: r._time),
            _time: now(),
        }),
    )
    |> map(
        fn: (r) => ({r with
            _message: messageFn(r: r),
        }),
    )
    |> experimental.group(
        mode: "extend",
        columns: [
            "_source_measurement",
            "_type",
            "_check_id",
            "_check_name",
            "_level",
        ],
    )
    |> write()
