// Package iox provides additional functions for querying data from InfluxDB IOx.
//
// ## Metadata
// introduced: 0.195.0
// contributors: **GitHub**: [@qxip](https://github.com/qxip)
//
package iox


import "sql"
import "date"
import "strings"

builtin _mask : (<-tables: stream[A], columns: [string]) => stream[B] where A: Record, B: Record

// from retrieves data from an IOx bucket between the `start` and `stop` times.
//
// This version of `from` is equivalent to `from() |> range()` in a single call.
//
// ## Parameters
//
// - bucket: Name of the IOx bucket to query.
//
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
//   Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//   For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//   Durations are relative to `now()`.
//
// - host: URL of the IOx instance to query.
// - org: Organization name.
// - token: [API token](https://docs.influxdata.com/influxdb/latest/security/tokens/).
// - table: Table used in the FlightSQL query.
// - limit: Limit for the FlightSQL query. Default is `1000`.
// - columns: Columns selected by the FlightSQL query. Default is `*`.
// - secure: Secure connection to IOx instance. Default is `true`.
//
// ## Examples
//
// ### Query using the bucket name
//
// ```no_run
// import "contrib/qxip/iox"
//
// iox.from(bucket: "sensors", org: "company", table: "cpu")
// ```
//
// ### Query a remote InfluxDB Cloud instance
//
// ```no_run
// import "contrib/qxip/iox"
// import "influxdata/influxdb/secrets"
//
// token = secrets.get(key: "INFLUXDB_CLOUD_TOKEN")
//
// from(
//     bucket: "sensors",
//     host: "https://eu-central-1-1.aws.cloud2.influxdata.com:443",
//     org: "company",
//     token: token,
//     table: "cpu",
//     start: -1h,
// )
// ```
//
// ## Metadata
// tags: inputs
from =
    (
            bucket,
            start=-1h,
            stop=now(),
            org="",
            host="",
            token="",
            table="",
            columns="*",
            limit="1000",
            secure="true",
        ) =>
        {

            dataSourceName =
                if org != "" and host != "" and token != "" then
                    "iox://${host}/${bucket}?secure=${secure}&token=${token}"
                else if org != "" and token != "" then
                    "iox://${host}/${org}_${bucket}?secure=${secure}&token=${token}"
                else if org != "" and host != "" then
                    "iox://${host}/${org}_${bucket}?secure=${secure}"
                else if host != "" and token != "" then
                    "iox://${host}/${bucket}?secure=${secure}&token=${token}"
                else if org != "" then
                    "iox://${host}/${org}_${bucket}?secure=${secure}"
                else if host != "" then
                    "iox://${host}/${bucket}?secure=${secure}&token=${token}"
                else
                    "iox://${host}/${bucket}?secure=${secure}"

            qStart = date.time(t: start)
            qStop = date.time(t: stop)
            query =
                "SELECT ${columns} FROM ${table} WHERE ( time >= '${qStart}' AND  time <= '${qStop}') LIMIT ${limit}"

            source =
                sql.from(
                    driverName: "influxdb-iox",
                    dataSourceName: "${dataSourceName}",
                    query: "${query}",
                )

            return source |> rename(columns: {time: "_time"})
        }

_from = from
