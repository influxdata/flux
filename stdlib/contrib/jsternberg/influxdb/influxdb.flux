package influxdb

import "influxdata/influxdb"
import "influxdata/influxdb/v1"

// select will select data from an influxdb instance within
// the range between `start` and `stop` from the bucket specified by
// the `from` parameter. It will select the specific measurement
// and it will only include fields that are included in the list of
// `fields`.
//
// In order to filter by tags, the `where` function can be used to further
// limit the amount of data selected.
select = (from, start, stop=now(), m, fields, where=(r) => true) => {
    tables = influxdb.from(bucket: from)
        |> range(start, stop)
        |> filter(fn: (r) => r._measurement == m)
        |> filter(fn: where)

    nfields = length(arr: fields)
    filtered = if nfields == 1 then
            tables |> filter(fn: (r) => r._field == fields[0])
        else if nfields == 2 then
            tables |> filter(fn: (r) => r._field == fields[0] or r._field == fields[1])
        else if nfields == 3 then
            tables |> filter(fn: (r) => r._field == fields[0] or r._field == fields[1] or r._field == fields[2])
        else if nfields == 4 then
            tables |> filter(fn: (r) => r._field == fields[0] or r._field == fields[1] or r._field == fields[2] or r._field == fields[3])
        else if nfields == 5 then
            tables |> filter(fn: (r) => r._field == fields[0] or r._field == fields[1] or r._field == fields[2] or r._field == fields[3] or r._field == fields[4])
        else
            tables |> filter(fn: (r) => contains(value: r._field, set: fields))
    return filtered |> v1.fieldsAsCols()
}
