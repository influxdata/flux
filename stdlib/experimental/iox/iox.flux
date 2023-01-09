// Package iox provides functions for querying data from IOx.
//
// ## Metadata
// introduced: 0.152.0
package iox


import "regexp"
import "strings"

// from reads from the selected bucket and measurement in an IOx storage node.
//
// This function creates a source that reads data from IOx. Output data is
// "pivoted" on the time column and includes columns for each returned
// tag and field per time value.
//
// ## Parameters
// - bucket: IOx bucket to read data from.
// - measurement: Measurement to read data from.
//
// ## Examples
//
// ### Use Flux to query data from IOx
// ```no_run
// import "experimental/iox"
//
// iox.from(bucket: "example-bucket", measurement: "example-measurement")
//     |> range(start: -1d)
//     |> filter(fn: (r) => r._field == "example-field")
// ```
//
// ## Metadata
// tags: inputs
builtin from : (bucket: string, measurement: string) => stream[{A with _time: time}] where A: Record

// sql executes an SQL query against a bucket in an IOx storage node.
//
// This function creates a source that reads data from IOx.
//
// ## Parameters
// - bucket: IOx bucket to read data from.
// - query: SQL query to execute.
//
// ## Examples
//
// ### Use SQL to query data from IOx
// ```no_run
// import "experimental/iox"
//
// iox.sql(bucket: "example-bucket", query: "SELECT * FROM measurement")
// ```
//
// ## Metadata
// introduced: 0.186.0
// tags: inputs
builtin sql : (bucket: string, query: string) => stream[A] where A: Record

// sqlInterval converts a duration value to a SQL interval string.
//
// SQL interval strings support down to millisecond precision.
// Any microsecond or nanosecond duration units are dropped from the duration value.
// If the duration only consists of microseconds or nanosecond units,
// `iox.sqlInterval()` returns `1 millisecond`.
// Duration values must be positive to work as a SQL interval string.
//
// ## Parameters
// - d: Duration value to convert to SQL interval string.
//
// ## Examples
//
// ### Convert a duration to a SQL interval
// ```no_run
// import "experimental/iox"
//
// iox.sqlInterval(d: 1y2mo3w4d5h6m7s8ms)
// // Returns 1 years 2 months 3 weeks 4 days 5 hours 6 minutes 7 seconds 8 milliseconds
// ```
//
// ### Use a Flux duration to define a SQL interval
// ```no_run
// import "experimental/iox"
//
// windowInterval = 1d12h
// sqlQuery = "
// SELECT
//   DATE_BIN(INTERVAL '${iox.sqlInterval(d: windowInterval)}', time, TIMESTAMP '2023-01-01T00:00:00Z')
//   COUNT(field1)
// FROM
//   measurement
// GROUP BY
//   time
// "
//
// iox.sql(bucket: "example-bucket", query: sqlQuery)
// ```
//
// ## Metadata
// introduced: 0.192.0
// tags: sql, type-conversions
sqlInterval = (d) => {
    _durationString = string(v: d)
    _pipeRegex = (v=<-, r, t) => regexp.replaceAllString(v: v, r: r, t: t)
    _intervalString =
        _pipeRegex(v: _durationString, r: /[\d]+(us|ns)/, t: "")
            |> _pipeRegex(r: /([^\d]+)/, t: " $1 ")
            |> _pipeRegex(r: / ms /, t: " milliseconds ")
            |> _pipeRegex(r: / s /, t: " seconds ")
            |> _pipeRegex(r: / m /, t: " minutes ")
            |> _pipeRegex(r: / h /, t: " hours ")
            |> _pipeRegex(r: / d /, t: " days ")
            |> _pipeRegex(r: / w /, t: " weeks ")
            |> _pipeRegex(r: / mo /, t: " months ")
            |> _pipeRegex(r: / y /, t: " years ")
    _output = if _intervalString == "" then "1 millisecond" else _intervalString

    return strings.trimSpace(v: _output)
}
