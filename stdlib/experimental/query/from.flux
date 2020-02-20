package query

fromRange = (bucket, start, stop=now()) =>
    from(bucket: bucket)
        |> range(start: start, stop: stop)

filterMeasurement = (table=<-, measurement) => table |> filter(fn: (r) => r._measurement == measurement)

filterFields = (table=<-, fields=[]) =>
    if length(arr: fields) == 0 then
        table
    else
        table |> filter(fn: (r) => contains(value: r._field, set: fields))

inBucket = (
    bucket,
    measurement,
    start,
    stop=now(),
    fields=[],
    predicate=(r) => true
) =>
    fromRange(bucket: bucket, start: start, stop: stop)
        |> filterMeasurement(measurement)
        |> filter(fn: predicate)
        |> filterFields(fields)
