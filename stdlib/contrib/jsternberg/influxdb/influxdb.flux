// Package influxdb provides additional functions for querying data from InfluxDB.
//
// introduced: 0.77.0
package influxdb


import "influxdata/influxdb"
import "influxdata/influxdb/v1"

// _mask does TODO
//
// _mask will hide the given columns from downstream
// transformations. It will not perform any copies and
// it will not regroup. This should only be used when
// the user knows it can't cause a key conflict.
// 
// ## Parameters
// 
// - columns:  TODO
builtin _mask : (<-tables: [A], columns: [string]) => [B] where A: Record, B: Record

// from TODO
// 
// from will retrieve data from a bucket between the start and stop time.
// This version of from is the equivalent of doing from |> range
// as a single call.
//
// ## Parameters
// 
// - bucket: TODO
// - start: TODO
// - stop: TODO
// - org: TODO
// - host: TODO
// - token: TODO
from = (
    bucket,
    start,
    stop=now(),
    org="",
    host="",
    token="",
) => {
    source = if org != "" and host != "" and token != "" then
        influxdb.from(bucket, org, host, token)
    else if org != "" and token != "" then
        influxdb.from(bucket, org, token)
    else if org != "" and host != "" then
        influxdb.from(bucket, org, host)
    else if host != "" and token != "" then
        influxdb.from(bucket, host, token)
    else if org != "" then
        influxdb.from(bucket, org)
    else if host != "" then
        influxdb.from(bucket, host)
    else if token != "" then
        influxdb.from(bucket, token)
    else
        influxdb.from(bucket)

    return source |> range(start, stop)
}

// _from allows us to reference the from function from
// within the select call which has a function parameter
// with the same name.
//
// ## Parameters
// 
// - bucket: TODO
// - start: TODO
// - stop: TODO
// - org: TODO
// - host: TODO
// - token: TODO
_from = from

// select is an alternate implementation of `from()`,
// `range()`, `filter()` and `pivot()` that returns pivoted query results and masks
// the `_measurement`, `_start`, and `_stop` columns. Results are similar to those
// returned by InfluxQL `SELECT` statements.
// 
// ## Parameters
// - from: (Required) Name of the bucket to query.
// - start: (Required) Earliest time to include in results.
//   Results include points that match the specified start time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//   For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//   Durations are relative to `now()`.
// - stop: Latest time to include in results.
//   Results exclude points that match the specified stop time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//   For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//   Durations are relative to `now()`.
//   Defaults to `now()`.
// - m: (Required) Name of the measurement to query.
// - fields: List of fields to query. Returns all fields when list is empty or unspecified.
//   Defaults to `[]`.
// - where: A single argument predicate function that evaluates true or false
//   and filters results based on tag values.
//   Records are passed to the function before fields are pivoted into columns.
//   Records that evaluate to true are included in the output tables.
//   Records that evaluate to null or false are not included in the output tables.
//   Defaults to `(r) => true`.
//   (Records evaluated in fn functions are represented by r, short for “record” or “row”.)
// - host: URL of the InfluxDB instance to query.
//   See [InfluxDB URLs](https://docs.influxdata.com/influxdb/v2.1/reference/urls/)
//   or [InfluxDB Cloud regions](https://docs.influxdata.com/influxdb/v2.1/reference/regions/).
// - org: Organization name.
// - token: InfluxDB [API token](https://docs.influxdata.com/influxdb/v2.1/security/tokens/).
// 
// ## Examples
// 
// ### Query a single field
// ```no_run
// import "contrib/jsternberg/influxdb"
// 
// influxdb.select(
//   from: "example-bucket",
//   start: -1d,
//   m: "example-measurement",
//   fields: ["field1"]
// )
// ```
// 
// ### Query multiple fields
// ```no_run
// import "contrib/jsternberg/influxdb"
// 
// influxdb.select(
//   from: "example-bucket",
//   start: -1d,
//   m: "example-measurement",
//   fields: ["field1", "field2", "field3"]
// )
// ```
// 
// ### Query all fields and filter by tags
// ```no_run
// import "contrib/jsternberg/influxdb"
// 
// influxdb.select(
//   from: "example-bucket",
//   start: -1d,
//   m: "example-measurement",
//   where: (r) => r.host == "host1" and r.region == "us-west"
// )
// ```
// 
// ### Query data from a remote InfluxDB Cloud instance
// ```no_run
// import "contrib/jsternberg/influxdb"
// import "influxdata/influxdb/secrets"
// 
// token = secrets.get(key: "INFLUXDB_CLOUD_TOKEN")
// 
// influxdb.select(
//   from: "example-bucket",
//   start: -1d,
//   m: "example-measurement",
//   fields: ["field1", "field2"],
//   host: "https://us-west-2-1.aws.cloud2.influxdata.com",
//   org: "example-org",
//   token: token
// )
// ```
select = (
    from,
    start,
    stop=now(),
    m,
    fields=[],
    org="",
    host="",
    token="",
    where=(r) => true,
) => {
    bucket = from
    tables = _from(
        bucket,
        start,
        stop,
        org,
        host,
        token,
    )
        |> filter(fn: (r) => r._measurement == m)
        |> filter(fn: where)
    nfields = length(arr: fields)
    fn = if nfields == 0 then
        (r) => true
    else if nfields == 1 then
        (r) => r._field == fields[0]
    else if nfields == 2 then
        (r) => r._field == fields[0] or r._field == fields[1]
    else if nfields == 3 then
        (r) => r._field == fields[0] or r._field == fields[1] or r._field == fields[2]
    else if nfields == 4 then
        (r) => r._field == fields[0] or r._field == fields[1] or r._field == fields[2] or r._field == fields[3]
    else if nfields == 5 then
        (r) => r._field == fields[0] or r._field == fields[1] or r._field == fields[2] or r._field == fields[3] or r._field == fields[4]
    else
        (r) => contains(value: r._field, set: fields)

    return tables
        |> filter(fn)
        |> v1.fieldsAsCols()
        |> _mask(columns: ["_measurement", "_start", "_stop"])
}
