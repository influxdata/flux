// Package schema provides functions for exploring your InfluxDB data schema.
package schema


// fieldsAsCols is a special application of the pivot() function that pivots
// on _field and _time columns to aligns fields within each input table that have the same timestamp.
//
// ## Examples
// ```
// import "influxdata/influxdb/schema"
//
// from(bucket:"example-bucket")
//   |> range(start: -1h)
//   |> filter(fn: (r) => r._measurement == "cpu")
//   |> schema.fieldsAsCols()
//   |> keep(columns: ["_time", "cpu", "usage_idle", "usage_user"])
// ```
//
fieldsAsCols = (tables=<-) => tables
    |> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")

// tagValues returns a list of unique values for a given tag. The return value is always a single table with a single column, _value.
//
// ## Parameters
// - `bucket` is the bucket to return unique tag values from.
// - `tag` is the tag to return unique values from
// - `predicate` is the predicate function that filters tag values. Defaults to (r) => true.
// - `start` Oldest time to include in results. Defaults to -30d.
//
//      Relative start times are defined using negative durations. Negative durations are relative to now.
//      Absolute start times are defined using time values.
//
// ## Examples
// ```
// import "influxdata/influxdb/schema"
//
// schema.tagValues(
//    bucket: "my-bucket",
//    tag: "host",
//  )
// ```
//
tagValues = (bucket, tag, predicate=(r) => true, start=-30d) => from(bucket: bucket)
    |> range(start: start)
    |> filter(fn: predicate)
    |> keep(columns: [tag])
    |> group()
    |> distinct(column: tag)

// tagKeys returns a list of tag keys for all series that match the predicate. The return value is always a single table with a single column, _value.
//
// ## Parameters
// - `bucket` is the bucket to return tag keys from.
// - `predicate` is the predicate function that filters tag keys. Defaults to (r) => true.
// - `start` Oldest time to include in results. Defaults to -30d.
//
//      Relative start times are defined using negative durations. Negative durations are relative to now.
//      Absolute start times are defined using time values.
//
// ## Examples
// ```
// import "influxdata/influxdb/schema"
//
// schema.tagKeys(bucket: "my-bucket")
// ```
//
// The return value is always a single table with a single column "_value".
//
tagKeys = (bucket, predicate=(r) => true, start=-30d) => from(bucket: bucket)
    |> range(start: start)
    |> filter(fn: predicate)
    |> keys()
    |> keep(columns: ["_value"])
    |> distinct()

// measurementTagValues returns a list of tag values for a specific measurement. The return value is always a single table with a single column, _value.
//
// ## Parameters
// - `bucket` is the bucket to return tag values from for a specific measurement.
// - `measurement` is the measurement to return tag values from.
// - `tag` is the tag to return all unique values from.
//
// The return value is always a single table with a single column "_value".
measurementTagValues = (bucket, measurement, tag) => tagValues(bucket: bucket, tag: tag, predicate: (r) => r._measurement == measurement)

// MeasurementTagKeys returns the list of tag keys for a specific measurement.
//
// ## Parameters
// - `bucket` is the bucket to return tag keys from for a specific measurement.
// - `measurement` is the measurement to return tag keys from.
//
measurementTagKeys = (bucket, measurement) => tagKeys(bucket: bucket, predicate: (r) => r._measurement == measurement)

// fieldKeys returns field keys in a bucket. The return value is always a single table with a single column, _value.
//
// ## Parameters
// - `bucket` is the bucket to list field keys from.
// - `predicate` is the predicate function that filters field keys. Defaults to (r) => true.
// - `start` Oldest time to include in results. Defaults to -30d.
//
//      Relative start times are defined using negative durations. Negative durations are relative to now.
//      Absolute start times are defined using time values.
//
// ## Examples
// ```
// import "influxdata/influxdb/schema"
//
// schema.fieldKeys(bucket: "my-bucket")
// ```
//
// FieldKeys is a special application of tagValues that returns field keys in a given bucket.
//
fieldKeys = (bucket, predicate=(r) => true, start=-30d) => tagValues(bucket: bucket, tag: "_field", predicate: predicate, start: start)

// measurementFieldKeys returns a list of fields in a measurement.
// The return value is always a single table with a single column, "_value".
//
// ## Parameters
// - `bucket` is the bucket to retrieve field keys from.
// - `measurement` is the measurement to list field keys from.
// - `start` Oldest time to include in results. Defaults to -30d.
//
//      Relative start times are defined using negative durations. Negative durations are relative to now.
//      Absolute start times are defined using time values.
//
// ## Examples
// ```
// import "influxdata/influxdb/schema"
//
// schema.measurementFieldKeys(
//   bucket: "telegraf",
//   measurement: "cpu",
// )
// ```
//
measurementFieldKeys = (bucket, measurement, start=-30d) => fieldKeys(bucket: bucket, predicate: (r) => r._measurement == measurement, start: start)

// measurements returns a list of measurements in a specific bucket.
// The return value is always a single table with a single column, _value.
//
// ## Parameters
// - `bucket` is the bucket to retrieve field keys from.
//
measurements = (bucket) => tagValues(bucket: bucket, tag: "_measurement")
