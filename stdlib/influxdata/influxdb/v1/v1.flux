// Package v1 provides tools for managing data from an InfluxDB v1.x database or
// structured using the InfluxDB v1 data structure.
//
// ### Deprecated functions
// In Flux 0.88.0, the following v1 package functions moved to
// the InfluxDB schema package. These functions are still available in the v1
// package for backwards compatibility, but are deprecated in favor of the
// schema package.
//
// - `v1.fieldKeys()`
// - `v1.fieldsAsCols()`
// - `v1.measurementFieldKeys()`
// - `v1.measurements()`
// - `v1.measurementTagKeys()`
// - `v1.measurementTagValues()`
// - `v1.tagKeys()`
// - `v1.tagValues()`
//
// ## Metadata
// introduced: 0.16.0
//
package v1


import "influxdata/influxdb/schema"

// json parses an InfluxDB 1.x JSON result into a stream of tables.
//
// ## Parameters
// - json: InfluxDB 1.x query results in JSON format.
//
//   _`json` and `file` are mutually exclusive._
//
// - file: File path to file containing InfluxDB 1.x query results in JSON format.
//
//   The path can be absolute or relative.
//   If relative, it is relative to the working directory of the `fluxd` process.
//   The JSON file must exist in the same file system running the `fluxd` process.
//
//   **Note**: InfluxDB OSS and InfluxDB Cloud do not support the `file` parameter.
//   Neither allow access to the underlying filesystem.
//
// ## Examples
//
// ### Convert a InfluxDB 1.x JSON query output string to a stream of tables
// ```
// import "influxdata/influxdb/v1"
//
// jsonData = "{
//     \"results\": [
//         {
//             \"statement_id\": 0,
//             \"series\": [
//                 {
//                     \"name\": \"cpu_load_short\",
//                     \"columns\": [
//                         \"time\",
//                         \"value\"
//                     ],
//                     \"values\": [
//                         [
//                             \"2021-01-01T00:00:00.000000000Z\",
//                             2
//                         ],
//                         [
//                             \"2021-01-01T00:01:00.000000000Z\",
//                             0.55
//                         ],
//                         [
//                             \"2021-01-01T00:02:00.000000000Z\",
//                             0.64
//                         ]
//                     ]
//                 }
//             ]
//         }
//     ]
// }"
//
// > v1.json(json: jsonData)
// ```
//
// ### Convert a InfluxDB 1.x JSON query output file to a stream of tables
// ```no_run
// import "influxdata/influxdb/v1"
//
// v1.json(file: "/path/to/results.json")
// ```
//
// ## Metadata
// tags: inputs
//
builtin json : (?json: string, ?file: string) => stream[A] where A: Record

// databases returns a list of databases in an InfluxDB 1.x (1.7+) instance.
//
// Output includes the following columns:
//
// - **databaseName**: Database name (string)
// - **retentionPolicy**: Retention policy name (string)
// - **retentionPeriod**: Retention period in nanoseconds (integer)
// - **default**: Default retention policy for the database (boolean)
//
// ## Parameters
// - org: Organization name.
// - orgID: Organization ID.
// - host: InfluxDB URL. Default is `http://localhost:8086`.
// - token: InfluxDB API token.
//
// ## Examples
//
// ### List databases from an InfluxDB instance
// ```no_run
// import "influxdata/influxdb/v1"
//
// v1.databases()
// ```
//
// ## Metadata
// tags: metadata
//
builtin databases : (
        ?org: string,
        ?orgID: string,
        ?host: string,
        ?token: string,
    ) => stream[{
        organizationID: string,
        databaseName: string,
        retentionPolicy: string,
        retentionPeriod: int,
        default: bool,
        bucketID: string,
    }]

// fieldsAsCols is a special application of `pivot()` that pivots input data
// on `_field` and `_time` columns to align fields within each input table that
// have the same timestamp.
//
// **Deprecated**: See influxdata/influxdata/schema.fieldsAsCols.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Pivot InfluxDB fields into columns
// ```
// # import "array"
// import "influxdata/influxdb/v1"
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
// >     |> v1.fieldsAsCols()
// ```
//
// ## Metadata
// deprecated: 0.88.0
// tags: transformations
//
fieldsAsCols = schema.fieldsAsCols

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
// import "influxdata/influxdb/v1"
//
// v1.tagValues(
//     bucket: "example-bucket",
//     tag: "host",
// )
// ```
//
// ## Metadata
// deprecated: 0.88.0
// tags: metadata
//
tagValues = schema.tagValues

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
// deprecated: 0.88.0
//
measurementTagValues = schema.measurementTagValues

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
// import "influxdata/influxdb/v1"
//
// v1.tagKeys(bucket: "example-bucket")
// ```
//
// ## Metadata
// tags: metadata
// deprecated: 0.88.0
//
tagKeys = schema.tagKeys

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
// deprecated: 0.88.0
//
measurementTagKeys = schema.measurementTagKeys

// fieldKeys returns field keys in a bucket.
//
// Results include a single table with a single column, `_value`.
//
// **Note**: FieldKeys is a special application of `tagValues that returns field
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
// deprecated: 0.88.0
//
fieldKeys = schema.fieldKeys

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
// deprecated: 0.88.0
//
measurementFieldKeys = schema.measurementFieldKeys

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
// deprecated: 0.88.0
//
measurements =
    schema.measurements// Maintain backwards compatibility by mapping the functions into the schema package.

