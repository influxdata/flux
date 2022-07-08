// Package events provides tools for analyzing event-based data.
//
// ## Metadata
// introduced: 0.91.0
// contributors: **GitHub**: [@tomhollingworth](https://github.com/tomhollingworth) | **InfluxDB Slack**: [@Tom Hollingworth](https://influxdata.com/slack)
//
package events


// duration calculates the duration of events.
//
// The function determines the time between a record and the subsequent record
// and associates the duration with the first record (start of the event).
// To calculate the duration of the last event,
// the function compares the timestamp of the final record
// to the timestamp in the `stopColumn` or the specified stop time.
//
// ### Similar functions
// `events.duration()` is similar to `elapsed()` and `stateDuration()`, but differs in important ways:
//
// - `elapsed()` drops the first record. `events.duration()` does not.
// - `stateDuration()` calculates the total time spent in a state (determined by a predicate function).
//   `events.duration()` returns the duration between all records and their subsequent records.
//
// See the example [below](#compared-to-similar-functions).
//
// ## Parameters
// - unit: Duration unit of the calculated state duration.
//   Default is `1ns`.
// - columnName: Name of the result column.
//   Default is `"duration"`.
// - timeColumn: Name of the time column.
//   Default is `"_time"`.
// - stopColumn: Name of the stop column.
//   Default is `"_stop"`.
// - stop: The latest time to use when calculating results.
//
//   If provided, `stop` overrides the time value in the `stopColumn`.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Calculate the duration of states
//
// ```
// import "array"
// import "contrib/tomhollingworth/events"
//
// # data = array.from(
// #     rows: [
// #         {_time: 2020-01-01T00:00:00Z, state: "ok"},
// #         {_time: 2020-01-01T00:12:34Z, state: "warn"},
// #         {_time: 2020-01-01T00:25:01Z, state: "ok"},
// #         {_time: 2020-01-01T16:07:55Z, state: "crit"},
// #         {_time: 2020-01-01T16:54:21Z, state: "warn"},
// #         {_time: 2020-01-01T18:20:45Z, state: "ok"},
// #     ],
// # )
// #
// < data
//     |> events.duration(
//         unit: 1m,
//         stop: 2020-01-02T00:00:00Z,
// >     )
// ```
//
// ### Compared to similar functions
//
// The example below includes output values of
// `events.duration()`, `elapsed()`, and `stateDuration()`
// related to the `_time` and `state` values of input data.
//
// ```
// import "array"
// import "contrib/tomhollingworth/events"
//
// # data = array.from(
// #     rows: [
// #         {_time: 2020-01-01T00:00:00Z, state: "ok"},
// #         {_time: 2020-01-01T00:12:34Z, state: "warn"},
// #         {_time: 2020-01-01T00:25:01Z, state: "ok"},
// #         {_time: 2020-01-01T16:07:55Z, state: "crit"},
// #         {_time: 2020-01-01T16:54:21Z, state: "warn"},
// #         {_time: 2020-01-01T18:20:45Z, state: "ok"},
// #     ],
// # )
// #
// union(
//     tables: [
//         data |> events.duration(unit: 1m, stop: 2020-01-02T00:00:00Z) |> map(fn: (r) => ({_time: r._time, state: r.state, function: "events.Duration()", value: r.duration})),
//         data |> elapsed(unit: 1m) |> map(fn: (r) => ({_time: r._time, state: r.state, function: "elapsed()", value: r.elapsed})),
//         data |> stateDuration(unit: 1m, fn: (r) => true) |> map(fn: (r) => ({_time: r._time, state: r.state, function: "stateDuration()", value: r.stateDuration})),
//     ],
// )
// >     |> pivot(rowKey: ["_time", "state"], columnKey: ["function"], valueColumn: "value")
// ```
//
// ## Metadata
// tags: transformations,events
//
builtin duration : (
        <-tables: stream[A],
        ?unit: duration,
        ?timeColumn: string,
        ?columnName: string,
        ?stopColumn: string,
        ?stop: time,
    ) => stream[B]
    where
    A: Record,
    B: Record
