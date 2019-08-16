package alerts

import "experimental"
import "influxdata/influxdb/v1"
import "influxdata/influxdb"

bucket = "_monitoring"

// Write persists the check statuses
write = (tables=<-) => tables |> influxdb.to(bucket: bucket)

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
