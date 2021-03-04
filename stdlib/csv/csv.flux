// CSV provides an API for working with [annotated CSV](https://github.com/influxdata/flux/blob/master/docs/SPEC.md#csv) files.
package csv


// From parses an annotated CSV and produces a stream of tables.
builtin from : (?csv: string, ?file: string, ?mode: string) => [A] where A: Record
