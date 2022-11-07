// Package csv provides functions for retrieving annotated CSV.
//
// ## Metadata
// introduced: 0.64.0
// deprecated: 0.173.0
// tags: csv
//
package csv


import c "csv"
import "experimental/http"

// from retrieves [annotated CSV](https://docs.influxdata.com/influxdb/latest/reference/syntax/annotated-csv/) **from a URL**.
//
// **Deprecated**: Experimental `csv.from()` is deprecated in favor of a combination of [`requests.get()`](https://docs.influxdata.com/flux/v0.x/stdlib/http/requests/get/) and [`csv.from()`](https://docs.influxdata.com/flux/v0.x/stdlib/csv/from/).
//
// **Note:** Experimental `csv.from()` is an alternative to the standard
// `csv.from()` function.
//
// ## Parameters
// - url: URL to retrieve annotated CSV from.
//
// ## Examples
//
// ### Query annotated CSV data from a URL using the requests package
//
// ```no_run
// import "csv"
// import "http/requests"
//
// response = requests.get(url: "http://example.com/csv/example.csv")
// csv.from(csv: string(v: response.body))
// ```
//
// ### Query annotated CSV data from a URL
// ```no_run
// import "experimental/csv"
//
// csv.from(url: "http://example.com/csv/example.csv")
// ```
//
from = (url) => c.from(csv: string(v: http.get(url: url).body))
