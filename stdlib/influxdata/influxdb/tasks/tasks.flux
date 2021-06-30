// Package tasks is an experimental package.
// The API for this package is not stable and should not
// be counted on for production code.
package tasks


// _zeroTime is a sentinel value for the zero time.
// This is used to mark that the lastSuccessTime has not been set.
builtin _zeroTime : time

// lastSuccessTime is the last time this task had run successfully.
option lastSuccessTime = _zeroTime

// _lastSuccess will return the time set on the option lastSuccessTime
// or it will return the orTime.
builtin _lastSuccess : (orTime: T, lastSuccessTime: time) => time where T: Timeable

// lastSuccess is a function that returns the time of the last successful run
//  of the InfluxDb task or the value of the orTime parameter if the task
//  has never successfully run.
//
// ## Parameters
// - `orTime` is the defualt time value returned if the task has never
//   successfully run.
//
// ## Example
//
// ```
// import "influxdata/influxdb/tasks"
//
// tasks.lastSuccess(orTime: 2020-01-01T00:00:00Z)
// ```
//
// ## Query data since the last successful task run
//
// ```
// import "influxdata/influxdb/tasks"
//
// option task = {
//   name: "Example task",
//   every: 30m
// }
//
// from(bucket: "example-bucket")
//   |> range(start: tasks.lastSuccess(orTime: -task.every))
//   // ...
// ```
lastSuccess = (orTime) => _lastSuccess(orTime, lastSuccessTime)
