// CSV provides an API for working with [annotated CSV](https://github.com/influxdata/flux/blob/master/docs/SPEC.md#csv) files.
package csv


// csv.from retrieves data from a comma-separated value (CSV) data source.
// The function returns a stream of tables. Each unique series is contained
// within its own table. Each record in the table represents a single point
// in the series.
//
// - `?csv` the CSV data. Supports annotated CSV or raw CSV. use mode to specify.
// - `?file` the file path of the CSV file to query. The path can be absolute or relative.
// - `?mode` the CSV parsing mode. The default is annotations.
//
builtin from : (?csv: string, ?file: string, ?mode: string) => [A] where A: Record
