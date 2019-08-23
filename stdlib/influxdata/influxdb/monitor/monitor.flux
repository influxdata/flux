package monitor

import "experimental"
import "influxdata/influxdb/v1"
import "influxdata/influxdb"

bucket = "_monitoring"

// write optionally persists the check statuses
option write = (tables=<-) => tables |> drop(columns: ["_start", "_stop"])

// From retrieves the check statuses that have been stored.
from = (start, stop=now(), fn) =>
    influxdb.from(bucket: bucket)
        |> range(start: start, stop: stop)
        |> filter(fn: fn)
        |> v1.fieldsAsCols()

// Log records notification events
log = (tables=<-) => tables |> experimental.to(bucket: bucket)

// Notify will call the endpoint and log the results.
notify = (tables=<-, endpoint, data={}) =>
    tables
        |> experimental.set(o: data)
        |> experimental.group(mode: "extend", columns: experimental.objectKeys(o: data))
        |> duplicate(column: "_time", as: "_status_timestamp")
        |> endpoint()
        |> log()

// Logs retrieves notification events that have been logged.
logs = (start, stop=now(), fn) =>
    influxdb.from(bucket: bucket)
        |> range(start: start, stop: stop)
        |> filter(fn: fn)
        |> v1.fieldsAsCols()

// Deadman takes in a stream of tables and reports which tables
// were observed strictly before t and which were observed after.
//
deadman = (t, tables=<-) => tables
    |> max(column: "_time")
    |> map(fn: (r) => ( {r with dead: r._time < t} ))

// levels describing the result of a check
levelOK = "ok"
levelInfo = "info"
levelWarn = "warn"
levelCrit = "crit"
levelUnknown = "unknown"

// Check performs a check against its input using the given ok, info, warn and crit functions
// and writes the result to a system bucket.
check = (
    tables=<-,
    data={},
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
            _source_timestamp: r._time,
            _time: now(),
        }))
        |> map(fn: (r) => ({r with
            _message: messageFn(r: r),
        }))
        |> experimental.group(mode: "extend", columns: ["_source_measurement", "_type", "_check_id", "_check_name", "_level"])
        |> write()
