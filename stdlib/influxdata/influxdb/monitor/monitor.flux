package monitor

import "experimental"
import "influxdata/influxdb/v1"
import "influxdata/influxdb"

bucket = "_monitoring"

// Write persists the check statuses
option write = (tables=<-) => tables |> experimental.to(bucket: bucket)

// Log records notification events
option log = (tables=<-) => tables |> experimental.to(bucket: bucket)

// From retrieves the check statuses that have been stored.
from = (start, stop=now(), fn=(r) => true) =>
    influxdb.from(bucket: bucket)
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
    toLevelFilter = if toLevel == "any" then (r) => r._level != fromLevel and exists r._level
                   else (r) => r._level == toLevel

    fromLevelFilter = if fromLevel == "any" then (r) => r._level != toLevel and exists r._level
                   else (r) => r._level == fromLevel

    return tables
        |> map(fn: (r) => ({r with level_value: if toLevelFilter(r: r) then 1
                                                else if fromLevelFilter(r: r) then 0
                                                else -10}))
        |> duplicate(column: "_level", as: "____temp_level____")
        |> drop(columns: ["_level"])
        |> rename(columns: {"____temp_level____": "_level"})
        |> sort(columns: ["_time"], desc: false)
        |> difference(columns: ["level_value"])
        |> filter(fn: (r) => r.level_value == 1)
        |> drop(columns: ["level_value"])
        |> experimental.group(mode: "extend", columns: ["_level"])
}

// stateChangesOnly takes a stream of tables that contains a _level column and
// returns a stream of tables where each record in a table represents a state change
// of the _level column.
stateChangesOnly = (tables=<-) => {
    return tables
        |> map(fn: (r) => ({r with level_value: if r._level == levelCrit then 4
                                                else if r._level == levelWarn then 3
                                                else if r._level == levelInfo then 2
                                                else if r._level == levelOK then 1
                                                else 0}))
        |> duplicate(column: "_level", as: "____temp_level____")
        |> drop(columns: ["_level"])
        |> rename(columns: {"____temp_level____": "_level"})
        |> sort(columns: ["_time"], desc: false)
        |> difference(columns: ["level_value"])
        |> filter(fn: (r) => r.level_value != 0)
        |> drop(columns: ["level_value"])
        |> experimental.group(mode: "extend", columns: ["_level"])
}

// StateChanges takes a stream of tables, fromLevel, and toLevel and returns
// a stream of tables where status has gone from fromLevel to toLevel.
//
// StateChanges only operates on data with data where r._level exists.
stateChanges = (fromLevel="any", toLevel="any", tables=<-) => {
    return if fromLevel == "any" and toLevel == "any" then tables |> stateChangesOnly()
           else tables |> _stateChanges(fromLevel: fromLevel, toLevel: toLevel)
}

// Notify will call the endpoint and log the results.
notify = (tables=<-, endpoint, data) =>
    tables
        |> experimental.set(o: data)
        |> experimental.group(mode: "extend", columns: experimental.objectKeys(o: data))
        |> map(fn: (r) => ({r with
            _measurement: "notifications",
            _status_timestamp: int(v: r._time),
            _time: now(),
        }))
        |> endpoint()
        |> experimental.group(mode: "extend", columns: ["_sent"])
        |> log()

// Logs retrieves notification events that have been logged.
logs = (start, stop=now(), fn) =>
    influxdb.from(bucket: bucket)
        |> range(start: start, stop: stop)
        |> filter(fn: (r) => r._measurement == "notifications")
        |> filter(fn: fn)
        |> v1.fieldsAsCols()

// Deadman takes in a stream of tables and reports which tables
// were observed strictly before t and which were observed after.
//
deadman = (t, tables=<-) => tables
    |> max(column: "_time")
    |> map(fn: (r) => ( {r with dead: r._time < t} ))

// Check performs a check against its input using the given ok, info, warn and crit functions
// and writes the result to a system bucket.
check = (
    tables=<-,
    data,
    messageFn,
    crit=(r) => false,
    warn=(r) => false,
    info=(r) => false,
    ok=(r) => true
) =>
    tables
        |> experimental.set(o: data.tags)
        |> experimental.group(mode: "extend", columns: experimental.objectKeys(o: data.tags))
        |> map(fn: (r) => ({r with
            _measurement: "statuses",
            _source_measurement: r._measurement,
            _type: data._type,
            _check_id:  data._check_id,
            _check_name: data._check_name,
            _level:
                if crit(r: r) then levelCrit
                else if warn(r: r) then levelWarn
                else if info(r: r) then levelInfo
                else if ok(r: r) then levelOK
                else levelUnknown,
            _source_timestamp: int(v:r._time),
            _time: now(),
        }))
        |> map(fn: (r) => ({r with
            _message: messageFn(r: r),
        }))
        |> experimental.group(mode: "extend", columns: ["_source_measurement", "_type", "_check_id", "_check_name", "_level"])
        |> write()
