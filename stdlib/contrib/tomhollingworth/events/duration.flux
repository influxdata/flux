// Package events provides tools for analyzing event-based data.
//
// introduced: 0.91.0
package events


// duration calculates the duration of events.
//
// The function determines the time between a record and the subsequent record
// and associates the duration with the first record (start of the event).
// To calculate the duration of the last event,
// the function compares the timestamp of the final record
// to the timestamp in the stopColumn or the specified stop time.
//
// > ## Similar functions
// > events.duration() is similar to elapsed() and stateDuration(), but differs in important ways:
// >
// >elapsed() drops the first record. events.duration() does not.
// >stateDuration() calculates the total time spent in a state (determined by a predicate function). events.duration() returns the duration between all records and their subsequent records.
// >
// >For examples, see below.
//
// ## Parameters
// - unit: Duration unit of the calculated state duration.
//   Default is `1ns`
// - columnName: Name of the result column.
//   Default is `"duration"`.
// - timeColumn: Name of the time column.
//   Default is `"_time"`.
// - stopColumn: Name of the stop column.
//   Default is `"_stop"`.
// - stop: The latest time to use when calculating results.
//   If provided, `stop` overrides the time value in the `stopColumn`.
//
// ## Examples
// ### Calculate the duration of states
//
// ```
// import "contrib/tomhollingworth/events"
//
// data
//   |> events.duration(
//     unit: 1m,
//     stop: 2020-01-02T00:00:00Z
//   )
// ```
//
// ## Compared to similar functions
//
// The example below includes output values of
// `events.duration()`, `elapsed()`, and `stateDuration()`
// related to the `_time` and `state` values of input data.
//
// [INPUT]
//
// ### Functions
//
// ```
// data |> events.duration(
//   unit: 1m,
//   stop: 2020-01-02T00:00:00Z
// )
//
// data |> elapsed(
//   unit: 1m
// )
//
// data |> stateDuration(
//   unit: 1m,
//   fn: (r) => true
// )
// ```
//
// [OUTPUT]
builtin duration : (
    <-tables: [A],
    ?unit: duration,
    ?timeColumn: string,
    ?columnName: string,
    ?stopColumn: string,
    ?stop: time,
) => [B] where A: Record, B: Record
