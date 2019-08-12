package alerts

import "influxdata/influxdb/v1"

bucket = "_monitoring"

// Write persists the check statuses
write = (tables=<-) => tables |> to(bucket: bucket)

// From retrieves the check statuses that have been stored.
statuses = (start, stop=now(), fn) =>
    from(bucket: bucket)
        |> range(start: start, stop: stop)
        |> filter(fn: fn)
        |> v1.fieldsAsCols()

// Log records notification events
log = (tables=<-) => tables |> to(bucket: bucket)

// Logs retrieves notification events that have been logged.
logs = (start, stop=now(), fn) =>
    from(bucket: bucket)
        |> range(start: start, stop: stop)
        |> filter(fn: fn)
        |> v1.fieldsAsCols()
