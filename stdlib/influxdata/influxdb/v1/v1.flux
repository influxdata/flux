// Package v1 provides an API for working with an InfluxDB v1.x instance.
// >NOTE: Must functions in this package are now deprecated see influxdata/influxdb/schema.
package v1


import "influxdata/influxdb/schema"

// Json parses an InfluxDB 1.x json result into a table stream.
builtin json : (?json: string, ?file: string) => [A] where A: Record

// databases is a function that returns a list of a database in
//  an InfluxDB 1.7+ instances.
//
// Output includes the following columns:
// - databaseName: Database name (string)
// - retentionPolicy: Retention policy name (string)
// - retentionPeriod: Retention period in nanoseconds (integer)
// - default: Default retention policy for database (boolean)
//
// ## Example
//
// ```
// import "influxdata/influxdb/v1"
//
// v1.database()
// ```
builtin databases : (
    ?org: string,
    ?orgID: string,
    ?host: string,
    ?token: string,
) => [{
    organizationID: string,
    databaseName: string,
    retentionPolicy: string,
    retentionPeriod: int,
    default: bool,
    bucketID: string,
}]

// fieldsAsCols is a function that is a special application of the pivot()
//  function that pivots on _field and _time columns to align fields within
//  each input table that have the same timestamp, and ressemble InfluxDB 1.x
//  query output.
//
// Deprecated: See influxdata/influxdata/schema.fieldsAsCols.
//
// ## Example
//
// ```
// import "influxdata/influxdb/v1"
//
// from(bucket:"example-bucket")
//   |> range(start: -1h)
//   |> filter(fn: (r) => r._measurement == "cpu")
//   |> v1.fieldsAsCols()
//   |> keep(columns: ["_time", "cpu", "usage_idle", "usage_user"])
// ```
fieldsAsCols = schema.fieldsAsCols

// tagValues is a function that returns a list of unique values for a
//  given tag.
//
//  The return value is always a single table with a single column _value.
//
// Deprecated: See influxdata/influxdata/schema.tagValues.
//
// ## Parameters
// - `bucket` is the bucket to return unique tag values from.
// - `tag` is the tag to return unique values from.
// - `predicate` is the predecate function that filters tag values.
//
//   Defaults to (r) => true.
//
// - `start` is the oldest time to include in results.
//
//   Defaults to -30d. Relative start times are defined using negative
//   durations. Negative durations are relative to now. Absolute start
//   times are defined using time values.
//
// ## Example
//
// ```
// import "influxdata/influxdb/v1"
//
// v1.tagValues(
//   bucket: "my-bucket",
//   tag: "host",
// )
// ```
tagValues = schema.tagValues

// measurementTagValues is a function that returns a list of tag values for
//  a specified measurement.
//
//  The return value is always a single table with a single column, _value.
//
// Deprecated: See influxdata/influxdata/schema.measurementTagValues.
//
// ## Parameters
// - `bucket` is the bucket to return tag values from a specific measurement.
// - `measurement` is the measurement to return tag values from.
// - `tag` is the tag to return all unique values from.
//
// ## Example
//
// ```
// import "influxdata/influxdb/v1"
//
// v1.measurementTagValues(
//   bucket: "example-bucket",
//   measurement: "cpu",
//   tag: "host"
// )
// ```
measurementTagValues = schema.measurementTagValues

// tagKeys is a function that returns a list of tag keys for all series that
//  match the predicate.
//
//  The return value is always a single table with a single column, _value.
//
// Deprecated: See influxdata/influxdata/schema.tagKeys.
//
// ## Parameters
// - `bucket` is the bucket to return tag keys from.
// - `predicate` is the predicate function that filters tag keys.
//
//   Defaults to (r) => true.
//
// - `start` is the oldest time to include in the results.
//
//   Defaults to -30d. Relative start times are defined using negative durations.
//   Absolute start times are defined using time values.
//
// ## Example
//
// ```
// import "influxdata/influxdb/v1"
//
// v1.tagKeys(
//   bucket: "example-bucket",
//   predicate: (r) => true,
//   start: -30d
// )
// ```
tagKeys = schema.tagKeys

// measurementTagKeys is a function that returns a list of tag keys for a specific
//  measurement.
//
//  The return value is always a single table with a single column, _value.
//
// Deprecated: See influxdata/influxdata/schema.measurementTagKeys.
//
// ## Parameters
// - `bucket` is the bucket to return the tag keys from a specific measurement.
// - `measurement` is the measurement to return tag key from.
//
// ## Example
//
// ```
// import "influxdata/influxdb/v1"
//
// v1.measurementTagKeys(
//   bucket: "example-bucket",
//   measurement: "cpu"
// )
// ```
measurementTagKeys = schema.measurementTagKeys

// fieldKeys is the function the returns field keys in a bucket.
//
//  The return value is always a single table with a single
//  column, _value.
//
// Deprecated: See influxdata/influxdata/schema.fieldKeys.
//
// ## Parameters
// - `bucket` is the bucket to list field keys from.
// - `predicate` is the predicate function that filters field keys.
//
//   Defaults to (r) => true.
//
// - `start` is the oldest time to include in results.
//
//   defaults to -30d. Relative start times are defined using negative
//   durations are relative to now. Absolute start times are defined
//   using time values.
//
// ## Example
//
// ```
// import "influxdata/influxdb/v1"
//
// v1.fieldKeys(
//   bucket: "example-bucket",
//   predicate: (r) => true,
//   start: -30d
// )
// ```
fieldKeys = schema.fieldKeys

// measurementFieldKeys is a function that returns a list of fields in a measurements.
//
//  The return value is always a single table with a single column, _value.
//
// Deprecated: See influxdata/influxdata/schema.measurementFieldKeys.
//
// ## Parameters
// - `bucket` is the bucket to retrieve field keys from.
// - `measurement` is the measurement to list field keys from.
// - `start` is is the oldest time to include in results.
//
//   Defaults to -30d. Relative start times are defined using negative durations. Negative
//   durations are relative to now. Absolute start times are defined using time values.
//
// ## Example
//
// ```
// import "influxdata/influxdb/v1"
//
// v1.measurementFieldKeys(
//   bucket: "example-bucket",
//   measurement: "example-measurement",
//   start: -30d
// )
// ```
measurementFieldKeys = schema.measurementFieldKeys

// measurements is a function that returns a list of measurements in a specific bucket.
//
//  The return value is always a single table with a single column, _value.
//
// Deprecated: See influxdata/influxdata/schema.measurements.
//
// ## Parameters
// - `bucket` is the bucket to recieves measurements from.
//
// ## Example
//
// ```
// import "influxdata/influxdb/v1"
//
// v1.measurements(bucket: "example-bucket")
// ```
measurements = schema.measurements
// Maintain backwards compatibility by mapping the functions into the schema package.
