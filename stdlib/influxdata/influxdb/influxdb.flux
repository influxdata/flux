// Package influxdb provides functions designed for working with InfluxDB and
// analyzing InfluxDB metadata.
//
// ## Metadata
// introduced: 0.7.0
package influxdb


// cardinality returns the series cardinality of data stored in InfluxDB.
//
// ## Parameters
// - bucket: Bucket to query cardinality from.
// - bucketID: String-encoded bucket ID to query cardinality from.
// - org: Organization name.
// - orgID: String-encoded organization ID.
// - host: URL of the InfluxDB instance to query.
//
//      See [InfluxDB Cloud regions](https://docs.influxdata.com/influxdb/cloud/reference/regions/)
//      or [InfluxDB OSS URLs](https://docs.influxdata.com/influxdb/latest/reference/urls/).
//
// - token: InfluxDB API token.
// - start: Earliest time to include when calculating cardinality.
//
//      The cardinality calculation includes points that match the specified start time.
//      Use a relative duration or absolute time. For example, `-1h` or `2019-08-28T22:00:00Z`.
//      Durations are relative to `now()`.
//
// - stop: Latest time to include when calculating cardinality.
//
//      The cardinality calculation excludes points that match the specified start time.
//      Use a relative duration or absolute time. For example, `-1h` or `2019-08-28T22:00:00Z`.
//      Durations are relative to `now()`. Default is `now()`.
//
// - predicate: Predicate function that filters records.
//      Default is `(r) => true`.
//
// ## Examples
//
// ### Query series cardinality in a bucket
// ```no_run
// import "influxdata/influxdb"
//
// influxdb.cardinality(
//     bucket: "example-bucket",
//     start: -1y,
// )
// ```
//
// ### Query series cardinality in a measurement//
// ```no_run
// import "influxdata/influxdb"
//
// influxdb.cardinality(
//     bucket: "example-bucket",
//     start: -1y,
//     predicate: (r) => r._measurement == "example-measurement",
// )
// ```
//
// ### Query series cardinality for a specific tag
// ```no_run
// import "influxdata/influxdb"
//
// influxdb.cardinality(
//    bucket: "example-bucket",
//    start: -1y,
//    predicate: (r) => r.exampleTag == "foo",
// )
// ```
//
// ## Metadata
// introduced: 0.92.0
// tags: metadata
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
    ) => stream[{_start: time, _stop: time, _value: int}]
    where
    A: Timeable,
    B: Timeable

// from queries data from an InfluxDB data source.
//
// It returns a stream of tables from the specified bucket.
// Each unique series is contained within its own table.
// Each record in the table represents a single point in the series.
//
// #### Query remote InfluxDB data sources
// Use `from()` to query data from remote **InfluxDB OSS 1.7+**,
// **InfluxDB Enterprise 1.9+**, and **InfluxDB Cloud**.
// To query remote InfluxDB sources, include the `host`, `token`, and `org`
// (or `orgID`) parameters.
//
// #### from() does not require a package import
// `from()` is part of the `influxdata/influxdb` package, but is part of the
// Flux prelude and does not require an import statement or package namespace.
//
// ## Parameters
// - bucket: Name of the bucket to query.
//   _`bucket` and `bucketID` are mutually exclusive_.
//
//     **InfluxDB 1.x or Enterprise**: Provide an empty string (`""`).
//
// - bucketID: String-encoded bucket ID to query.
//   _`bucket` and `bucketID` are mutually exclusive_.
//
//     **InfluxDB 1.x or Enterprise**: Provide an empty string (`""`).
//
// - host: URL of the InfluxDB instance to query.
//
//     See [InfluxDB Cloud regions](https://docs.influxdata.com/influxdb/cloud/reference/regions/)
//     or [InfluxDB OSS URLs](https://docs.influxdata.com/influxdb/latest/reference/urls/).
//
// - org: Organization name.
//   _`org` and `orgID` are mutually exclusive_.
//
//     **InfluxDB 1.x or Enterprise**: Provide an empty string (`""`).
//
// - orgID: String-encoded organization ID to query.
//   _`org` and `orgID` are mutually exclusive_.
//
//     **InfluxDB 1.x or Enterprise**: Provide an empty string (`""`).
//
// - token: InfluxDB API token.
//
//     **InfluxDB 1.x or Enterprise**: If authentication is disabled, provide an
//     empty string (`""`). If authentication is enabled, provide your InfluxDB
//     username and password using the `<username>:<password>` syntax.
//
// ## Examples
//
// ### Query InfluxDB using the bucket name
// ```no_run
// from(bucket: "example-bucket")
// ```
//
// ### Query InfluxDB using the bucket ID
// ```no_run
// from(bucketID: "0261d8287f4d6000")
// ```
//
// ### Query a remote InfluxDB Cloud instance
// ```no_run
// import "influxdata/influxdb/secrets"
//
// token = secrets.get(key: "INFLUXDB_CLOUD_TOKEN")
//
// from(
//     bucket: "example-bucket",
//     host: "https://us-west-2-1.aws.cloud2.influxdata.com",
//     org: "example-org",
//     token: token,
// )
// ```
//
// ## Metadata
// tags: inputs
//
builtin from : (
        ?bucket: string,
        ?bucketID: string,
        ?org: string,
        ?orgID: string,
        ?host: string,
        ?token: string,
    ) => stream[{B with _measurement: string, _field: string, _time: time, _value: A}]

// to writes data to an InfluxDB Cloud or 2.x bucket and returns the written data.
//
// ### Output data requirements
// `to()` writes data structured using the standard InfluxDB Cloud and v2.x data
// structure that includes, at a minimum, the following columns:
//
// - `_time`
// - `_measurement`
// - `_field`
// - `_value`
//
// All other columns are written to InfluxDB as
// [tags](https://docs.influxdata.com/influxdb/cloud/reference/key-concepts/data-elements/#tags).
//
// **Note**: `to()` drops rows with null `_time` values and does not write them
// to InfluxDB.
//
// #### to() does not require a package import
// `to()` is part of the `influxdata/influxdb` package, but is part of the
// Flux prelude and does not require an import statement or package namespace.
//
// ## Parameters
// - bucket: Name of the bucket to write to.
//   _`bucket` and `bucketID` are mutually exclusive_.
// - bucketID: String-encoded bucket ID to to write to.
//   _`bucket` and `bucketID` are mutually exclusive_.
// - host: URL of the InfluxDB instance to write to.
//
//     See [InfluxDB Cloud regions](https://docs.influxdata.com/influxdb/cloud/reference/regions/)
//     or [InfluxDB OSS URLs](https://docs.influxdata.com/influxdb/latest/reference/urls/).
//
//     `host` is required when writing to a remote InfluxDB instance.
//     If specified, `token` is also required.
//
// - org: Organization name.
//   _`org` and `orgID` are mutually exclusive_.
// - orgID: String-encoded organization ID to query.
//   _`org` and `orgID` are mutually exclusive_.
// - token: InfluxDB API token.
//
//     **InfluxDB 1.x or Enterprise**: If authentication is disabled, provide an
//     empty string (`""`). If authentication is enabled, provide your InfluxDB
//     username and password using the `<username>:<password>` syntax.
//
//     `token` is required when writing to another organization or when `host`
//     is specified.
//
// - timeColumn: Time column of the output. Default is `"_time"`.
// - measurementColumn: Measurement column of the output. Default is `"_measurement"`.
// - tagColumns: Tag columns in the output. Defaults to all columns with type
//   `string`, excluding all value columns and columns identified by `fieldFn`.
// - fieldFn: Function that maps a field key to a field value and returns a record.
//   Default is `(r) => ({ [r._field]: r._value })`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Write data to InfluxDB
// ```no_run
// data = array.from(
//     rows: [
//         {_time: 2021-01-01T00:00:00Z, _measurement: "m", tag1: "a", _field: "temp", _value: 100.1},
//         {_time: 2021-01-01T00:01:00Z, _measurement: "m", tag1: "a", _field: "temp", _value: 99.8},
//         {_time: 2021-01-01T00:02:00Z, _measurement: "m", tag1: "a", _field: "temp", _value: 99.1},
//         {_time: 2021-01-01T00:03:00Z, _measurement: "m", tag1: "a", _field: "temp", _value: 98.6},
//     ],
// )
//
// data
//     |> to(
//         bucket: "example-bucket",
//         org: "example-org",
//         token: "mYSuP3rSecR37t0k3N",
//         host: "http://localhost:8086",
//     )
// ```
//
// The example above produces the following line protocol and sends it to the
// InfluxDB `/api/v2/write` endpoint:
//
// ```txt
// m,tag1=a temp=100.1 1609459200000000000
// m,tag1=a temp=99.8 1609459260000000000
// m,tag1=a temp=99.1 1609459320000000000
// m,tag1=a temp=98.6 1609459380000000000
// ```
//
// ### Customize measurement, tag, and field columns in the to() operation
// ```no_run
// data = array.from(
//     rows: [
//         {_time: 2021-01-01T00:00:00Z, tag1: "a", tag2: "b", hum: 53.3, temp: 100.1},
//         {_time: 2021-01-01T00:01:00Z, tag1: "a", tag2: "b", hum: 53.4, temp: 99.8},
//         {_time: 2021-01-01T00:02:00Z, tag1: "a", tag2: "b", hum: 53.6, temp: 99.1},
//         {_time: 2021-01-01T00:03:00Z, tag1: "a", tag2: "b", hum: 53.5, temp: 98.6},
//     ],
// )
//
// data
//     |> to(
//         bucket: "example-bucket",
//         measurementColumn: "tag1",
//         tagColumns: ["tag2"],
//         fieldFn: (r) => ({"hum": r.hum, "temp": r.temp}),
//     )
// ```
//
// The example above produces the following line protocol and sends it to the
// InfluxDB `/api/v2/write` endpoint:
//
// ```txt
// a,tag2=b hum=53.3,temp=100.1 1609459200000000000
// a,tag2=b hum=53.4,temp=99.8 1609459260000000000
// a,tag2=b hum=53.6,temp=99.1 1609459320000000000
// a,tag2=b hum=53.5,temp=98.6 1609459380000000000
// ```
//
// ### Write to multiple InfluxDB buckets
// The example below does the following:
//
// 1. Writes data to `bucket1` and returns the data as it is written.
// 2. Applies an empty group key to group all rows into a single table.
// 3. Counts the number of rows.
// 4. Maps columns required to write to InfluxDB.
// 5. Writes the modified data to `bucket2`.
//
// ```no_run
// data
//     |> to(bucket: "bucket1")
//     |> group()
//     |> count()
//     |> map(fn: (r) => ({r with _time: now(), _measurement: "writeStats", _field: "numPointsWritten"}))
//     |> to(bucket: "bucket2")
// ```
//
// ## Metadata
// tags: outputs
//
builtin to : (
        <-tables: stream[A],
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
    ) => stream[A]
    where
    A: Record,
    B: Record

// buckets returns a list of buckets in the specified organization.
//
// ## Parameters
// - org: Organization name. Default is the current organization.
//
//   _`org` and `orgID` are mutually exclusive_.
//
// - orgID: Organization ID. Default is the ID of the current organization.
//
//   _`org` and `orgID` are mutually exclusive_.
//
// - host: URL of the InfluxDB instance.
//
//     See [InfluxDB Cloud regions](https://docs.influxdata.com/influxdb/cloud/reference/regions/)
//     or [InfluxDB OSS URLs](https://docs.influxdata.com/influxdb/latest/reference/urls/).
//
//     _`host` is required when `org` or `orgID` are specified._
//
// - token: InfluxDB API token.
//
//     _`token` is required when `host`, `org, or `orgID` are specified._
//
// ## Examples
//
// ### List buckets in an InfluxDB organization
// ```no_run
// buckets(
//     org: "example-org",
//     host: "http://localhost:8086",
//     token: "mYSuP3rSecR37t0k3N",
// )
// ```
//
// ## Metadata
// introduced: 0.16.0
// tags: metadata
//
builtin buckets : (
        ?org: string,
        ?orgID: string,
        ?host: string,
        ?token: string,
    ) => stream[{
        name: string,
        id: string,
        organizationID: string,
        retentionPolicy: string,
        retentionPeriod: int,
    }]

// wideTo writes wide data to an InfluxDB 2.x or InfluxDB Cloud bucket.
// Wide data is _pivoted_ in that its fields are represented as columns making the table wider.
//
// #### Requirements and behavior
// - Requires both a `_time` and a `_measurement` column.
// - All columns in the group key (other than `_measurement`) are written as tags
//   with the column name as the tag key and the column value as the tag value.
// - All columns **not** in the group key (other than `_time`) are written as
//   fields with the column name as the field key and the column value as the field value.
//
// If using the `from()` to query data from InfluxDB, use pivot() to transform
// data into the structure `experimental.to()` expects.
//
// ## Parameters
// - bucket: Name of the bucket to write to.
//   _`bucket` and `bucketID` are mutually exclusive_.
// - bucketID: String-encoded bucket ID to to write to.
//   _`bucket` and `bucketID` are mutually exclusive_.
// - host: URL of the InfluxDB instance to write to.
//
//     See [InfluxDB Cloud regions](https://docs.influxdata.com/influxdb/cloud/reference/regions/)
//     or [InfluxDB OSS URLs](https://docs.influxdata.com/influxdb/latest/reference/urls/).
//
//     `host` is required when writing to a remote InfluxDB instance.
//     If specified, `token` is also required.
//
// - org: Organization name.
//   _`org` and `orgID` are mutually exclusive_.
// - orgID: String-encoded organization ID to query.
//   _`org` and `orgID` are mutually exclusive_.
// - token: InfluxDB API token.
//
//     **InfluxDB 1.x or Enterprise**: If authentication is disabled, provide an
//     empty string (`""`). If authentication is enabled, provide your InfluxDB
//     username and password using the `<username>:<password>` syntax.
//
//     `token` is required when writing to another organization or when `host`
//     is specified.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Pivot and write data to InfluxDB
//
// ```no_run
// import "influxdata/influxdb"
// import "influxdata/influxdb/schema"
//
// from(bucket: "example-bucket")
//     |> range(start: -1h)
//     |> schema.fieldsAsCols()
//     |> wideTo(bucket: "example-target-bucket")
// ```
//
// ## Metadata
// introduced: 0.174.0
// tags: outputs
//
builtin wideTo : (
        <-tables: stream[A],
        ?bucket: string,
        ?bucketID: string,
        ?org: string,
        ?orgID: string,
        ?host: string,
        ?token: string,
    ) => stream[A]
    where
    A: Record
