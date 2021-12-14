// Package csv provides functions for retrieving annotated CSV.
//
// introduced: 0.64.0
// tags: csv
//
package csv


import c "csv"
import "experimental/http"

// from retrieves [annotated CSV](https://docs.influxdata.com/influxdb/latest/reference/syntax/annotated-csv/) **from a URL**.
//
// **Note:** Experimental `csv.from()` is an alternative to the standard
// `csv.from()` function.
//
// ## Parameters
// - url: URL to retrieve annotated CSV from.
//
// ## Examples
// ### Query annotated CSV data from a URL
// ```no_run
// import "experimental/csv"
//
// csv.from(url: "http://example.com/csv/example.csv")
// ```
//
from = (url) => c.from(csv: string(v: http.get(url: url).body))
