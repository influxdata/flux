// Package query provides functions meant to simplify common InfluxDB queries.
//
// The primary function in this package is `query.inBucket()`, which uses all
// other functions in this package.
//
// ## Metadata
// introduced: 0.60.0
//
package query


// fromRange returns all data from a specified bucket within given time bounds.
//
// ## Parameters
// - bucket: InfluxDB bucket name.
// - start: Earliest time to include in results.
//
//   Results include points that match the specified start time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//   For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//   Durations are relative to `now()`.
//
// - stop: Latest time to include in results. Default is `now()`.
//
//   Results exclude points that match the specified stop time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in seconds).For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//   Durations are relative to `now()`.
//
// ## Examples
// ### Query data from InfluxDB in a specified time range
// ```no_run
// import "experimental/query"
//
// query.fromRange(bucket: "example-bucket", start: -1h)
// ```
//
// ## Metadata
// tags: transformations,filters
//
fromRange = (bucket, start, stop=now()) =>
    from(bucket: bucket)
        |> range(start: start, stop: stop)

// filterMeasurement filters input data by measurement.
//
// ## Parameters
// - measurement: InfluxDB measurement name to filter by.
// - table: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Query data from InfluxDB in a specific measurement
// ```no_run
// import "experimental/query"
//
// query.fromRange(bucket: "example-bucket", start: -1h)
//     |> query.filterMeasurement(measurement: "example-measurement")
// ```
//
// ## Metadata
// tags: transformations,filters
//
filterMeasurement = (table=<-, measurement) =>
    table |> filter(fn: (r) => r._measurement == measurement)

// filterFields filters input data by field.
//
// ## Parameters
// - fields: Fields to filter by. Default is `[]`.
// - table: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Query specific fields from InfluxDB
// ```no_run
// import "experimental/query"
//
// query.fromRange(bucket: "telegraf", start: -1h)
//     |> query.filterFields(fields: ["used_percent", "available_percent"])
// ```
//
// ## Metadata
// tags: transformations,filters
//
filterFields = (table=<-, fields=[]) =>
    if length(arr: fields) == 0 then
        table
    else
        table |> filter(fn: (r) => contains(value: r._field, set: fields))

// inBucket queries data from a specified InfluxDB bucket within given time bounds,
// filters data by measurement, field, and optional predicate expressions.
//
// ## Parameters
// - bucket: InfluxDB bucket name.
// - measurement: InfluxDB measurement name to filter by.
// - start: Earliest time to include in results.
//
//   Results include points that match the specified start time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//   For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//   Durations are relative to `now()`.
//
// - stop: Latest time to include in results. Default is `now()`.
//
//   Results exclude points that match the specified stop time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in seconds).For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//   Durations are relative to `now()`.
//
// - fields: Fields to filter by. Default is `[]`.
// - predicate: Predicate function that evaluates column values and returns `true` or `false`.
//   Default is `(r) => true`.
//
//   Records (`r`) are passed to the function.
//   Those that evaluate to `true` are included in the output tables.
//   Records that evaluate to null or `false` are not included in the output tables.
//
// ## Examples
// ### Query specific fields in a measurement from InfluxDB
// ```no_run
// import "experimental/query"
//
// query.inBucket(
//     bucket: "example-buckt",
//     start: -1h,
//     measurement: "mem",
//     fields: ["field1", "field2"],
//     predicate: (r) => r.host == "host1",
// )
// ```
//
// ## Metadata
// tags: inputs
//
inBucket = (
    bucket,
    measurement,
    start,
    stop=now(),
    fields=[],
    predicate=(r) => true,
) =>
    fromRange(bucket: bucket, start: start, stop: stop)
        |> filterMeasurement(measurement)
        |> filter(fn: predicate)
        |> filterFields(fields)
