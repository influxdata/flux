package v1

// Json parses an InfluxDB 1.x json result into a table stream.
builtin json

// Databases returns the list of available databases, it has no parameters.
builtin databases

// fieldsAsCols is a special application of pivot that will automatically align fields within each measurement that have the same timestamp.
fieldsAsCols = (tables=<-) =>
    tables
        |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")

// TagValues returns the unique values for a given tag.
// The return value is always a single table with a single column "_value".
tagValues = (bucket, tag, predicate=(r) => true, start=-30d) =>
    from(bucket: bucket)
      |> range(start: start)
      |> filter(fn: predicate)
      |> keep(columns: [tag])
      |> group()
      |> distinct(column: tag)

// MeasurementTagValues returns a single table with a single column "_value" that contains the
// The return value is always a single table with a single column "_value".
measurementTagValues = (bucket, measurement, tag) =>
    tagValues(bucket: bucket, tag: tag, predicate: (r) => r._measurement == measurement)

// TagKeys returns the list of tag keys for all series that match the predicate.
// The return value is always a single table with a single column "_value".
tagKeys = (bucket, predicate=(r) => true, start=-30d) =>
    from(bucket: bucket)
        |> range(start: start)
        |> filter(fn: predicate)
        |> keys()
        |> keep(columns: ["_value"])
        |> distinct()

// MeasurementTagKeys returns the list of tag keys for a specific measurement.
measurementTagKeys = (bucket, measurement) =>
    tagKeys(bucket: bucket, predicate: (r) => r._measurement == measurement)

// Measurements returns the list of measurements in a specific bucket.
measurements = (bucket) =>
    tagValues(bucket: bucket, tag: "_measurement")

