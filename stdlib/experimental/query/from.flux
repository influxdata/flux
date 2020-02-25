package query

import "strings"

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

splitData = (v, t=",") => strings.split(v:v, t:t)
sanitizeList = (v) => strings.replaceAll(v:v, t: ", ", u: ",")

select = (
    src,
    fields,
    start,
    stop=now(),
    where=(r) => true
) => {
    srcData = splitData(v: src, t: ":")
    bucket = srcData[0]
    measurements = sanitizeList(v: string(v: srcData[1]))
    measurementsArr = splitData(v: measurements)
    fieldsList = sanitizeList(v: fields)
    fieldsArr = splitData(v: fieldsList)

    data = from(bucket: bucket)
        |> range(start: start, stop: stop)
        |> filter(fn: (r) =>
          (if measurements == "*" then true else contains(value: r._measurement, set: measurementsArr)) and
          (if fields == "*" then true else contains(value: r._field, set: fieldsArr)))
        |> filter(fn: where)

    return data
}
