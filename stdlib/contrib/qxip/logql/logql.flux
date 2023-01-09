// Package logql provides functions for using [LogQL](https://grafana.com/docs/loki/latest/logql/) to query a [Loki](https://grafana.com/oss/loki/) data source.
//
// The primary function in this package is `logql.query_range()`
//
// ## Metadata
// introduced: 0.192.0
//
package logql


import "csv"
import "date"
import "experimental"
import "experimental/http/requests"

// defaultURL is the default LogQL HTTP API URL.
option defaultURL = "http://127.0.0.1:3100"

// defaultAPI is the default LogQL Query Range API Path.
option defaultAPI = "/loki/api/v1/query_range"

// query_range queries data from a specified LogQL query within given time bounds,
// filters data by query, timerange, and optional limit expressions.
// All values are returned as string values (using `raw` mode in `csv.from`)
//
// ## Parameters
// - url: LogQL/qryn URL and port. Default is `http://qryn:3100`.
// - path: LogQL query_range API path.
// - limit: Query limit. Default is 100.
// - query: LogQL query to execute.
// - start: Earliest time to include in results. Default is `-1h`.
//
//   Results include points that match the specified start time.
//   Use a relative duration or absolute time.
//   For example, `-1h` or `2022-01-01T22:00:00.801064Z`.
//
// - end: Latest time to include in results. Default is `now()`.
//
//   Results exclude points that match the specified stop time.
//   Use a relative duration or absolute time.
//   For example, `-1h` or `2022-01-01T22:00:00.801064Z`.
//
// - step: Query resolution step width in seconds. Default is 10.
//
//   Only applies to query types which produce a matrix response.
// - orgid: Optional Loki organization ID for partitioning. Default is `""`.
//
// ## Examples
// ### Query specific fields in a measurement from LogQL/qryn
// ```no_run
// import "contrib/qxip/logql"
//
// option logql.defaultURL = "http://qryn:3100"
//
// logql.query_range(
//     query: "{job=\"dummy-server\"}",
//     start: -1h,
//     end: now(),
//     limit: 100,
// )
// ```
//
// ## Metadata
// tags: inputs
//
query_range = (
        url=defaultURL,
        path=defaultAPI,
        query,
        limit=100,
        step=10,
        start=-1h,
        end=now(),
        orgid="",
    ) =>
    {
        dstart = date.time(t: start)
        dend = date.time(t: end)
        response =
            requests.get(
                url: url + path,
                params:
                    [
                        "query": [query],
                        "limit": ["${limit}"],
                        "start": [string(v: uint(v: dstart))],
                        "end": [string(v: uint(v: dend))],
                        "step": ["${step}"],
                        "csv": ["1"],
                    ],
                headers: if orgid != "" then ["X-Scope-OrgID": orgid] else [:],
                body: bytes(v: query),
            )

        return csv.from(csv: string(v: response.body), mode: "raw")
    }
