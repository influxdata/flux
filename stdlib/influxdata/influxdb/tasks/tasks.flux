// Package tasks provides tools for working with InfluxDB tasks.
//
// ## Metadata
// introduced: 0.84.0
//
package tasks


// _zeroTime is a sentinel value for the zero time.
// This is used to mark that the **lastSuccessTime** has not been set.
builtin _zeroTime : time

// lastSuccessTime is the last time this task ran successfully.
option lastSuccessTime = _zeroTime

// _lastSuccess returns the time set on the option **lastSuccessTime**
// or the `orTime`.
builtin _lastSuccess : (orTime: T, lastSuccessTime: time) => time where T: Timeable

// lastSuccess returns the time of the last successful run of the InfluxDB task
// or the value of the `orTime` parameter if the task has never successfully run.
//
// ## Parameters
// - orTime: Defualt time value returned if the task has never successfully run.
//
// ## Examples
//
// ### Return the time an InfluxDB task last succesfully ran
// ```no_run
// import "influxdata/influxdb/tasks"
//
// tasks.lastSuccess(orTime: 2020-01-01T00:00:00Z)
// ```
//
// ## Query data since the last successful task run
//
// ```no_run
// import "influxdata/influxdb/tasks"
//
// option task = {
//     name: "Example task",
//     every: 30m,
// }
//
// from(bucket: "example-bucket")
//     |> range(start: tasks.lastSuccess(orTime: -task.every))
// ```
//
// ## Metadata
// tags: metadata
//
lastSuccess = (orTime) => _lastSuccess(orTime, lastSuccessTime)
