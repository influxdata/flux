// Package clickhouse provides functions to query [ClickHouse](https://clickhouse.com/) using the ClickHouse HTTP API.
//
// ## Metadata
// introduced: 0.192.0
//
package clickhouse


import "csv"
import "experimental"
import "experimental/http/requests"

// defaultURL is the default ClickHouse HTTP API URL.
option defaultURL = "http://127.0.0.1:8123"

// query queries data from ClickHouse using specified parameters.
//
// ## Parameters
// - url: ClickHouse HTTP API URL. Default is `http://127.0.0.1:8123`.
// - query: ClickHouse query to execute.
// - limit: Query rows limit. Defaults is `100`.
// - cors: Request remote CORS headers. Defaults is `1`.
// - max_bytes: Query bytes limit. Default is `10000000`.
// - format: Query format. Default is `CSVWithNames`.
//
//   _For information about available formats, see [ClickHouse formats](https://clickhouse.com/docs/en/interfaces/formats/)._
//
// ## Examples
//
// ### Query ClickHouse
// ```no_run
// import "contrib/qxip/clickhouse"
//
// option clickhouse.defaultURL = "https://play@play.clickhouse.com"
//
// clickhouse.query(query: "SELECT version()")
// ```
//
// ## Metadata
// tags: inputs
//
query = (
        url=defaultURL,
        query,
        limit=100,
        cors="1",
        max_bytes=10000000,
        format="CSVWithNames",
    ) =>
    {
        response =
            requests.get(
                url: url,
                params:
                    [
                        "query": [query],
                        "default_format": [format],
                        "max_result_rows": ["${limit}"],
                        "max_result_bytes": ["${max_bytes}"],
                        "add_http_cors_header": [cors],
                    ],
                headers: ["X-ClickHouse-Format": string(v: format)],
            )

        return csv.from(csv: string(v: response.body), mode: "raw")
    }
