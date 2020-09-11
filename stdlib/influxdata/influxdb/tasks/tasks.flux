// Package tasks is an experimental package.
// The API for this package is not stable and should not
// be counted on for production code.
package tasks

// _zeroTime is a sentinel value for the zero time.
// This is used to mark that the lastSuccessTime has not been set.
builtin _zeroTime: time

// lastSuccessTime is the last time this task had run successfully.
option lastSuccessTime = _zeroTime

// _lastSuccess will return the time set on the option lastSuccessTime
// or it will return the orTime.
builtin _lastSuccess: (orTime: time, lastSuccessTime: time) => time

// lastSuccess will return the last successful time a task ran
// within an influxdb task. If the task has not successfully run,
// the orTime will be returned.
lastSuccess = (orTime) => _lastSuccess(orTime, lastSuccessTime)
