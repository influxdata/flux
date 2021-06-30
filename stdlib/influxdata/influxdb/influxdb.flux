// Package influxdb provides functions for analyzing InfluxDB metadata.
package influxdb


// cardinality returns the series cardinality of data stored in InfluxDB Cloud.
//
// ## Parameters
// - `bucket` is the bucket to query cardinality
// - `bucketID` is the String-encoded bucket ID to query cardinality from.
// - `org` is the organization name
// - `orgID` is the String-encoded organization ID to query cardinality from.
// - `host` is the URL of the InfluxDB instance to query.
//
//      See InfluxDB Cloud regions or InfluxDB OSS URLs.
//
// - `token` is the InfluxDB authentication token.
// - `start` is the earliest time to include when calculating cardinality.
//
//      The cardinality calculation includes points that match the specified start time.
//      Use a relative duration or absolute time. For example, -1h or 2019-08-28T22:00:00Z.
//      Durations are relative to now().
//
// - `stop` is the latest time to include when calculating cardinality.
//
//      The cardinality calculation excludes points that match the specified start time.
//      Use a relative duration or absolute time. For example, -1h or 2019-08-28T22:00:00Z.
//      Durations are relative to now(). Defaults to now().
// - `predicate` is the predicate function that filters records. Defaults to (r) => true.
//
//
// ## Query series cardinality in a bucket
//
// ```
// import "influxdata/influxdb
//
// influxdb.cardinality(
//    bucket: "example-bucket",
//    start: -1y
// )
// ```
//
// ## Query series cardinality in a measurement
//
// ```
// import "influxdata/influxdb
//
// influxdb.cardinality(
//    bucket: "example-bucket",
//    start: -1y
//    predicate: (r) => r._measurement == "example-measurement"
// )
// ```
//
// ## Query series cardinality for a specific tag
//
// ```
// import "influxdata/influxdb
//
// influxdb.cardinality(
//    bucket: "example-bucket",
//    start: -1y
//    predicate: (r) => r.exampleTag == "foo"
// )
// ```
//
builtin cardinality : (
    ?bucket: string,
    ?bucketID: string,
    ?org: string,
    ?orgID: string,
    ?host: string,
    ?token: string,
    start: A,
    ?stop: B,
    ?predicate: (r: {T with _measurement: string, _field: string, _value: S}) => bool,
) => [{_start: time, _stop: time, _value: int}] where
    A: Timeable,
    B: Timeable

builtin from : (
    ?bucket: string,
    ?bucketID: string,
    ?org: string,
    ?orgID: string,
    ?host: string,
    ?token: string,
) => [{B with _measurement: string, _field: string, _time: time, _value: A}]

builtin to : (
    <-tables: [A],
    ?bucket: string,
    ?bucketID: string,
    ?org: string,
    ?orgID: string,
    ?host: string,
    ?token: string,
    ?timeColumn: string,
    ?measurementColumn: string,
    ?tagColumns: [string],
    ?fieldFn: (r: A) => B,
) => [A] where
    A: Record,
    B: Record

builtin buckets : (
    ?org: string,
    ?orgID: string,
    ?host: string,
    ?token: string,
) => [{
    name: string,
    id: string,
    organizationID: string,
    retentionPolicy: string,
    retentionPeriod: int,
}]