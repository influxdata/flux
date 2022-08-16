// Package schema provides functions for exploring your InfluxDB data schema.
//
// ## Metadata
// introduced: 0.88.0
package schema


// internal only option to make testing possible
option _from = from

_startDefault = -30d
_stopDefault = now()

// fieldsAsCols is a special application of `pivot()` that pivots input data
// on `_field` and `_time` columns to align fields within each input table that
// have the same timestamp.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Pivot InfluxDB fields into columns
// ```
// # import "array"
// import "influxdata/influxdb/schema"
//
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T12:00:00Z, _measurement: "m", loc: "Seattle", _field: "temp", _value: "73.1"},
// #         {_time: 2021-01-02T12:00:00Z, _measurement: "m", loc: "Seattle", _field: "temp", _value: "68.2"},
// #         {_time: 2021-01-03T12:00:00Z, _measurement: "m", loc: "Seattle", _field: "temp", _value: "61.4"},
// #         {_time: 2021-01-01T12:00:00Z, _measurement: "m", loc: "Seattle", _field: "hum", _value: "89.2"},
// #         {_time: 2021-01-02T12:00:00Z, _measurement: "m", loc: "Seattle", _field: "hum", _value: "90.5"},
// #         {_time: 2021-01-03T12:00:00Z, _measurement: "m", loc: "Seattle", _field: "hum", _value: "81.0"},
// #     ],
// # )
// #     |> group(columns: ["_time", "_value"], mode: "except")
// #
// < data
// >     |> schema.fieldsAsCols()
// ```
//
// ## Metadata
// tags: transformations
//
fieldsAsCols = (tables=<-) =>
    tables
        |> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")

// tagValues returns a list of unique values for a given tag.
//
// Results include a single table with a single column, `_value`.
//
// ## Parameters
// - bucket: Bucket to return unique tag values from.
// - tag: Tag to return unique values from.
// - predicate: Predicate function that filters tag values.
//   Default is `(r) => true`.
// - start: Oldest time to include in results. Default is `-30d`.
// - stop: Newest time include in results.
//     The stop time is exclusive, meaning values with a time equal to stop time are excluded from the results.
//     Default is `now()`.
//
//   Relative start times are defined using negative durations.
//   Negative durations are relative to `now()`.
//   Absolute start times are defined using time values.
//
// ## Examples
//
// ### Query unique tag values from an InfluxDB bucket
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.tagValues(
//     bucket: "example-bucket",
//     tag: "host",
// )
// ```
//
// ## Metadata
// tags: metadata
//
tagValues = (
    bucket,
    tag,
    predicate=(r) => true,
    start=_startDefault,
    stop=_stopDefault,
) =>
    _from(bucket: bucket)
        |> range(start: start, stop: stop)
        |> filter(fn: predicate)
        |> keep(columns: [tag])
        |> group()
        |> distinct(column: tag)

// tagKeys returns a list of tag keys for all series that match the `predicate`.
//
// Results include a single table with a single column, `_value`.
//
// ## Parameters
// - bucket: Bucket to return tag keys from.
// - predicate: Predicate function that filters tag keys.
//   Default is `(r) => true`.
// - start: Oldest time to include in results. Default is `-30d`.
// - stop: Newest time include in results.
//     The stop time is exclusive, meaning values with a time equal to stop time are excluded from the results.
//     Default is `now()`.
//
//   Relative start times are defined using negative durations.
//   Negative durations are relative to `now()`.
//   Absolute start times are defined using time values.
//
// ## Examples
//
// ### Query tag keys in an InfluxDB bucket
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.tagKeys(bucket: "example-bucket")
// ```
//
// ## Metadata
// tags: metadata
//
tagKeys = (bucket, predicate=(r) => true, start=_startDefault, stop=_stopDefault) =>
    _from(bucket: bucket)
        |> range(start: start, stop: stop)
        |> filter(fn: predicate)
        |> keys()
        |> keep(columns: ["_value"])
        |> distinct()

// measurementTagValues returns a list of tag values for a specific measurement.
//
// Results include a single table with a single column, `_value`.
//
// ## Parameters
// - bucket: Bucket to return tag values from for a specific measurement.
// - measurement: Measurement to return tag values from.
// - tag: Tag to return all unique values from.
// - start: Oldest time to include in results. Default is `-30d`.
// - stop: Newest time include in results.
//     The stop time is exclusive, meaning values with a time equal to stop time are excluded from the results.
//     Default is `now()`.
//
// ## Examples
//
// ### Query unique tag values from an InfluxDB measurement
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.measurementTagValues(
//     bucket: "example-bucket",
//     measurement: "example-measurement",
//     tag: "example-tag",
// )
// ```
//
// ## Metadata
// tags: metadata
//
measurementTagValues = (
    bucket,
    measurement,
    tag,
    start=_startDefault,
    stop=_stopDefault,
) =>
    tagValues(
        bucket: bucket,
        tag: tag,
        predicate: (r) => r._measurement == measurement,
        start: start,
        stop: stop,
    )

// measurementTagKeys returns the list of tag keys for a specific measurement.
//
// Results include a single table with a single column, `_value`.
//
// ## Parameters
// - bucket: Bucket to return tag keys from for a specific measurement.
// - measurement: Measurement to return tag keys from.
// - start: Oldest time to include in results. Default is `-30d`.
// - stop: Newest time include in results.
//     The stop time is exclusive, meaning values with a time equal to stop time are excluded from the results.
//     Default is `now()`.
//
// ## Examples
//
// ### Query tag keys from an InfluxDB measurement
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.measurementTagKeys(
//     bucket: "example-bucket",
//     measurement: "example-measurement",
// )
// ```
//
// ## Metadata
// tags: metadata
//
measurementTagKeys = (bucket, measurement, start=_startDefault, stop=_stopDefault) =>
    tagKeys(
        bucket: bucket,
        predicate: (r) => r._measurement == measurement,
        start: start,
        stop: stop,
    )

// fieldKeys returns field keys in a bucket.
//
// Results include a single table with a single column, `_value`.
//
// **Note**: FieldKeys is a special application of `tagValues()` that returns field
// keys in a given bucket.
//
// ## Parameters
// - bucket: Bucket to list field keys from.
// - predicate: Predicate function that filters field keys.
//   Default is `(r) => true`.
// - start: Oldest time to include in results. Default is `-30d`.
// - stop: Newest time include in results.
//     The stop time is exclusive, meaning values with a time equal to stop time are excluded from the results.
//     Default is `now()`.
//
//   Relative start times are defined using negative durations.
//   Negative durations are relative to `now()`.
//   Absolute start times are defined using time values.
//
// ## Examples
//
// ### Query field keys from an InfluxDB bucket
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.fieldKeys(bucket: "example-bucket")
// ```
//
// ## Metadata
// tags: metadata
//
fieldKeys = (bucket, predicate=(r) => true, start=_startDefault, stop=_stopDefault) =>
    tagValues(
        bucket: bucket,
        tag: "_field",
        predicate: predicate,
        start: start,
        stop: stop,
    )

// measurementFieldKeys returns a list of fields in a measurement.
//
// Results include a single table with a single column, `_value`.
//
// ## Parameters
// - bucket: Bucket to retrieve field keys from.
// - measurement: Measurement to list field keys from.
// - start: Oldest time to include in results. Default is `-30d`.
// - stop: Newest time include in results.
//     The stop time is exclusive, meaning values with a time equal to stop time are excluded from the results.
//     Default is `now()`.
//
//   Relative start times are defined using negative durations.
//   Negative durations are relative to `now()`.
//   Absolute start times are defined using time values.
//
// ## Examples
//
// ### Query field keys from an InfluxDB measurement
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.measurementFieldKeys(
//     bucket: "example-bucket",
//     measurement: "example-measurement",
// )
// ```
//
// ## Metadata
// tags: metadata
//
measurementFieldKeys = (bucket, measurement, start=_startDefault, stop=_stopDefault) =>
    fieldKeys(
        bucket: bucket,
        predicate: (r) => r._measurement == measurement,
        start: start,
        stop: stop,
    )

// measurements returns a list of measurements in a specific bucket.
//
// Results include a single table with a single column, `_value`.
//
// ## Parameters
// - bucket: Bucket to retrieve measurements from.
// - start: Oldest time to include in results. Default is `-30d`.
// - stop: Newest time include in results.
//     The stop time is exclusive, meaning values with a time equal to stop time are excluded from the results.
//     Default is `now()`.
//
// ## Examples
//
// ### Return a list of measurements in an InfluxDB bucket
// ```no_run
// import "influxdata/influxdb/schema"
//
// schema.measurements(bucket: "example-bucket")
// ```
//
// ## Metadata
// tags: metadata
//
measurements = (bucket, start=_startDefault, stop=_stopDefault) =>
    tagValues(bucket: bucket, tag: "_measurement", start: start, stop: stop)
