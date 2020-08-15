package influxdb

import "influxdata/influxdb"
import "influxdata/influxdb/v1"

// _mask will hide the given columns from downstream
// transformations. It will not perform any copies and
// it will not regroup. This should only be used when
// the user knows it can't cause a key conflict.
builtin _mask : (<-tables: [A], columns: [string]) => [B] where A: Record, B: Record

// from will retrieve data from a bucket between the start and stop time.
// This version of from is the equivalent of doing from |> range
// as a single call.
from = (bucket, start, stop=now(), org="", host="", token="") => {
    source =
        if org != "" and host != "" and token != "" then
            influxdb.from(bucket, org, host, token)
        else if org != "" and token != "" then
            influxdb.from(bucket, org, token)
        else if org != "" and host != "" then
            influxdb.from(bucket, org, host)
        else if host != "" and token != "" then
            influxdb.from(bucket, host, token)
        else if org != "" then
            influxdb.from(bucket, org)
        else if host != "" then
            influxdb.from(bucket, host)
        else if token != "" then
            influxdb.from(bucket, token)
        else
            influxdb.from(bucket)

    return source |> range(start, stop)
}

// _from allows us to reference the from function from
// within the select call which has a function parameter
// with the same name.
_from = from

// select will select data from an influxdb instance within
// the range between `start` and `stop` from the bucket specified by
// the `from` parameter. It will select the specific measurement
// and it will only include fields that are included in the list of
// `fields`.
//
// In order to filter by tags, the `where` function can be used to further
// limit the amount of data selected.
select = (from, start, stop=now(), m, fields=[], org="", host="", token="", where=(r) => true) => {
    bucket = from
    tables = _from(bucket, start, stop, org, host, token)
        |> filter(fn: (r) => r._measurement == m)
        |> filter(fn: where)

    nfields = length(arr: fields)
    fn =
        if nfields == 0 then
            (r) => true
        else if nfields == 1 then
            (r) => r._field == fields[0]
        else if nfields == 2 then
            (r) => r._field == fields[0] or r._field == fields[1]
        else if nfields == 3 then
            (r) => r._field == fields[0] or r._field == fields[1] or r._field == fields[2]
        else if nfields == 4 then
            (r) => r._field == fields[0] or r._field == fields[1] or r._field == fields[2] or r._field == fields[3]
        else if nfields == 5 then
            (r) => r._field == fields[0] or r._field == fields[1] or r._field == fields[2] or r._field == fields[3] or r._field == fields[4]
        else
            (r) => contains(value: r._field, set: fields)

    return tables
        |> filter(fn)
        |> v1.fieldsAsCols()
        |> _mask(columns: ["_measurement", "_start", "_stop"])
}
