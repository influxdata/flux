// Package csv provides tools for working with data in annotated CSV format.
//
// ## Metadata
// introduced: 0.14.0
// tags: csv
package csv


// from retrieves data from a comma separated value (CSV) data source and
// returns a stream of tables.
//
// ## Parameters
//
// - csv: CSV data.
//
//   Supports anonotated CSV or raw CSV. Use `mode` to specify the parsing mode.
//
// - file: File path of the CSV file to query.
//
//   The path can be absolute or relative.
//   If relative, it is relative to the working directory of the `fluxd` process.
//   The CSV file must exist in the same file system running the `fluxd` process.
//
// - mode: is the CSV parsing mode. Default is `annotations`.
//
//     **Available annotation modes**
//
//     - **annotations**: Use CSV notations to determine column data types.
//     - **raw**: Parse all columns as strings and use the first row as the
//       header row and all subsequent rows as data.
//
// ## Examples
//
// ### Query anotated CSV data from file
//
// ```no_run
// import "csv"
//
// csv.from(file: "path/to/data-file.csv")
// ```
//
// ### Query raw data from CSV file
//
// ```no_run
// import "csv"
//
// csv.from(
//     file: "/path/to/data-file.csv",
//     mode: "raw",
// )
// ```
//
// ### Query an annotated CSV string
//
// ```
// import "csv"
//
// csvData = "
// #datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double
// #group,false,false,false,false,false,true,true,false
// #default,,,,,,,,
// ,result,table,_start,_stop,_time,region,host,_value
// ,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43
// ,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:51:00Z,east,A,65.15
// ,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25
// ,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:51:20Z,east,B,18.67
// ,mean,2,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62
// ,mean,2,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:51:40Z,east,C,82.16
// "
//
// > csv.from(csv: csvData)
// ```
//
// ### Query a raw CSV string
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
//     csv: csvData,
//     mode: "raw",
// > )
// ```
//
// ## Metadata
// tags: csv,inputs
builtin from : (?csv: string, ?file: string, ?mode: string) => stream[A] where A: Record
