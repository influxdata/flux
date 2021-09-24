// Package csv provides tools for working with data in annotated CSV format.
package csv


// from is a function that retrieves data from a comma separated value (CSV) data source.
//
// A stream of tables are returned, each unique series contained within its own table.
// Each record in the table represents a single point in the series.
//
// ## Parameters
// - `csv` is CSV data.
//
//   Supports anonotated CSV or raw CSV. Use mode to specify the parsing mode.
//
// - `file` is the file path of the CSV file to query.
//
//   The path can be absolute or relative. If relative, it is relative to the working
//   directory of the `fluxd` process. The CSV file must exist in the same file
//   system running the `fluxd` process.
//
// - `mode` is the CSV parsing mode. Default is annotations.
//
//   Available annotation modes:
//    - annotations: Use CSV notations to determine column data types.
//    - raw: Parse all columns as strings and use the first row as the header row
//    - and all subsequent rows as data.
//
// ## Query anotated CSV data from file
//
// ```
// import "csv"
//
// csv.from(file: "path/to/data-file.csv")
// ```
//
// ## Query raw data from CSV file
//
// ```
// import "csv"
//
// csv.from(
//   file: "/path/to/data-file.csv",
//   mode: "raw"
// )
// ```
//
// ## Query an annotated CSV string
//
// ```
// import "csv"
//
// csvData = "
// #datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double
// #group,false,false,false,false,false,false,false,false
// #default,,,,,,,,
// ,result,table,_start,_stop,_time,region,host,_value
// ,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43
// ,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25
// ,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62
// "
//
// csv.from(csv: csvData)
//
// ```
//
// ## Query a raw CSV string
//
// ```
// import "csv"
//
// csvData = "
// _start,_stop,_time,region,host,_value
// 2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43
// 2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25
// 2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62
// "
//
// csv.from(
//   csv: csvData,
//   mode: "raw"
// )
// ```
builtin from : (?csv: string, ?file: string, ?mode: string) => [A] where A: Record
