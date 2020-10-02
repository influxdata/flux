package schema

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

// FieldKeys is a special application of tagValues that returns field keys in a given bucket.
// The return value is always a single table with a single column, "_value".
fieldKeys = (bucket, predicate=(r) => true, start=-30d) =>
    tagValues(bucket: bucket, tag: "_field", predicate: predicate, start: start)

// MeasurementFieldKeys returns field keys in a given measurement.
// The return value is always a single table with a single column, "_value".
measurementFieldKeys = (bucket, measurement, start=-30d) =>
    fieldKeys(bucket: bucket, predicate: (r) => r._measurement == measurement, start: start)

// Measurements returns the list of measurements in a specific bucket.
measurements = (bucket) =>
    tagValues(bucket: bucket, tag: "_measurement")
