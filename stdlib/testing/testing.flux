// Package testing functions test piped-forward data in specific ways and return errors if the tests fail.
package testing


import c "csv"

// assertEquals tests whether two streams have identical data.
//
//      If equal, the function outputs the tested data stream unchanged.
//      If unequal, the function returns an error.
//
// assertEquals can be used to perform in-line tests in a query.
//
// ## Parameters
// - `name` is the unique name given to the assertion.
// - `got` is the stream containing data to test. Defaults to piped-forward data (<-).
// - `want` is the stream that contains the expected data to test against.
//
// ## Assert of separate streams
// ```
// import "testing"
//
// want = from(bucket: "backup-example-bucket")
//   |> range(start: -5m)
//
// got = from(bucket: "example-bucket")
//   |> range(start: -5m)
//
// testing.assertEquals(got: got, want: want)
// ```
//
// ## Inline assertion
// ```
// import "testing"
//
// want = from(bucket: "backup-example-bucket")
//   |> range(start: -5m)
//
// from(bucket: "example-bucket")
//   |> range(start: -5m)
//   |> testing.assertEquals(want: want)
// ```
//
builtin assertEquals : (name: string, <-got: [A], want: [A]) => [A]

// assertEmpty tests if an input stream is empty. If not empty, the function returns an error.
// assertEmpty can be used to perform in-line tests in a query.
//
// ## Check if there is a difference between streams
//
//      This example uses the testing.diff() function which outputs the diff for the two streams.
//      The .testing.assertEmpty() function checks to see if the diff is empty.
//
// ```
// import "testing"
//
// got = from(bucket: "example-bucket")
//   |> range(start: -15m)
// want = from(bucket: "backup_example-bucket")
//   |> range(start: -15m)
// got
//   |> testing.diff(want: want)
//   |> testing.assertEmpty()
// ```
//
builtin assertEmpty : (<-tables: [A]) => [A]

// diff produces a diff between two streams.
//
// It matches tables from each stream with the same group keys.
//
//      For each matched table, it produces a diff. Any added or removed rows are added to the table as a row.
//      An additional string column with the name diff is created and contains a
//      - if the row was present in the got table and not in the want table or + if the opposite is true.
//
// The diff function is guaranteed to emit at least one row if the tables are different and no rows if the tables are the same. The exact diff produced may change.
// diff can be used to perform in-line diffs in a query.
//
// ## Parameters
// - `got` is the stream containing data to test. Defaults to piped-forward data (<-).
// - `want` is the stream that contains the expected data to test against.
// - `epsilon` specifies how far apart two float values can be, but still considered equal. Defaults to 0.000000001.
//
// ## Diff separate streams
// ```
// import "testing"
//
// want = from(bucket: "backup-example-bucket")
//   |> range(start: -5m)
// got = from(bucket: "example-bucket")
//   |> range(start: -5m)
// testing.diff(got: got, want: want)
// ```
//
// ## Inline diff
// ```
// import "testing"
//
// want = from(bucket: "backup-example-bucket") |> range(start: -5m)
// from(bucket: "example-bucket")
//   |> range(start: -5m)
//   |> testing.diff(want: want)
// ```
//
builtin diff : (
    <-got: [A],
    want: [A],
    ?verbose: bool,
    ?epsilon: float,
    ?nansEqual: bool,
) => [{A with _diff: string}]

// loadStorage loads annotated CSV test data as if it were queried from InfluxDB.
// This function ensures tests behave correctly in both the Flux and InfluxDB test suites.
//
// ## Function Requirements
// - Test data requires the _field, _measurement, and _time columns
//
// ## Parameters
// - `csv` is the annotated CSV data to load
//
// ## Examples
// ```
// import "testing"
//
// csvData = "
// #datatype,string,long,string,dateTime:RFC3339,string,double
// #group,false,false,true,false,true,false
// #default,_result,,,,,
// ,result,table,_measurement,_time,_field,_value
// ,,0,m,2021-01-01T00:00:00Z,t,1.2
// ,,0,m,2021-01-02T00:00:00Z,t,1.4
// ,,0,m,2021-01-03T00:00:00Z,t,2.2
// "
//
// testing.loadStorage(csv: csvData)
// ```
//
option loadStorage = (csv) => c.from(csv: csv)
    |> range(start: 1800-01-01T00:00:00Z, stop: 2200-12-31T11:59:59Z)
    |> map(
        fn: (r) => ({r with
            _field: if exists r._field then r._field else die(msg: "test input table does not have _field column"),
            _measurement: if exists r._measurement then r._measurement else die(msg: "test input table does not have _measurement column"),
            _time: if exists r._time then r._time else die(msg: "test input table does not have _time column"),
        }),
    )

// load loads tests data from a stream of tables.
//
// ## Parameters
// - `tables` is the input data. Default is piped-forward data (<-).
//
// ## Load a raw stream of tables in a test case
//
//      The following test uses array.from() to create two streams of tables to compare in the test.
//
// ```
// import "testing"
// import "array"
//
// got = array.from(rows: [
//   {_time: 2021-01-01T00:00:00Z, _measurement: "m", _field: "t", _value: 1.2},
//   {_time: 2021-01-01T01:00:00Z, _measurement: "m", _field: "t", _value: 0.8},
//   {_time: 2021-01-01T02:00:00Z, _measurement: "m", _field: "t", _value: 3.2}
// ])
//
// want = array.from(rows: [
//   {_time: 2021-01-01T00:00:00Z, _measurement: "m", _field: "t", _value: 1.2},
//   {_time: 2021-01-01T01:00:00Z, _measurement: "m", _field: "t", _value: 0.8},
//   {_time: 2021-01-01T02:00:00Z, _measurement: "m", _field: "t", _value: 3.1}
// ])
//
// testing.diff(got, want)
// ```
//
option load = (tables=<-) => tables

// loadMem loads annotated CSV test data from memory to emulate query results returned by Flux.
//
// ## Parameters
// - `csv` is the annotated CSV data to load.
//
// ## Examples
// ```
// import "testing"
//
// csvData = "
// #datatype,string,long,string,dateTime:RFC3339,string,double
// #group,false,false,true,false,true,false
// #default,_result,,,,,
// ,result,table,_measurement,_time,_field,_value
// ,,0,m,2021-01-01T00:00:00Z,t,1.2
// ,,0,m,2021-01-02T00:00:00Z,t,1.4
// ,,0,m,2021-01-03T00:00:00Z,t,2.2
// "
//
// testing.loadMem(csv: csvData)
// ```
option loadMem = (csv) => c.from(csv: csv)

// inspect returns information about a test case.
//
// ## Parameters
// - `case` is the test case to inspect.
//
// ## Define and inspect a test case
// ```
// import "testing"
//
// inData = "
// #datatype,string,long,string,dateTime:RFC3339,string,double
// #group,false,false,true,false,true,false
// #default,_result,,,,,
// ,result,table,_measurement,_time,_field,_value
// ,,0,m,2021-01-01T00:00:00Z,t,1.2
// ,,0,m,2021-01-02T00:00:00Z,t,1.4
// ,,0,m,2021-01-03T00:00:00Z,t,2.2
// "
//
// outData = "
// #datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,double
// #group,false,false,true,true,true,true,false
// #default,_result,,,,,,
// ,result,table,_start,_stop,_measurement,_field,_value
// ,,0,2021-01-01T00:00:00Z,2021-01-03T01:00:00Z,m,t,4.8
// "
//
// t_sum = (table=<-) =>
//   (table
//     |> range(start:2021-01-01T00:00:00Z, stop:2021-01-03T01:00:00Z)
//     |> sum())
//
// test _sum = () =>
//   ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_sum})
//
// testing.inpsect(case: _sum)
//
// // Returns: {
// //   fn: (<-table: [{_time: time | t10997}]) -> [t10996],
// //   input: fromCSV -> range -> map,
// //   want: fromCSV -> yield,
// //   got: fromCSV -> range -> map -> range -> sum -> yield,
// //   diff: ( fromCSV; fromCSV -> range -> map -> range -> sum;  ) -> diff -> yield
// // }
// ```
inspect = (case) => {
    tc = case()
    got = tc.input |> tc.fn()
    dif = got |> diff(want: tc.want)

    return {
        fn: tc.fn,
        input: tc.input,
        want: tc.want |> yield(name: "want"),
        got: got |> yield(name: "got"),
        diff: dif |> yield(name: "diff"),
    }
}

// run executes a specified test case.
//
// ## Parameters
// - `case` is the test case to run.
//
// ## Define and execute a test case
// ```
// import "testing"
//
// inData = "
// #datatype,string,long,string,dateTime:RFC3339,string,double
// #group,false,false,true,false,true,false
// #default,_result,,,,,
// ,result,table,_measurement,_time,_field,_value
// ,,0,m,2021-01-01T00:00:00Z,t,1.2
// ,,0,m,2021-01-02T00:00:00Z,t,1.4
// ,,0,m,2021-01-03T00:00:00Z,t,2.2
// "
//
// outData = "
// #datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,double
// #group,false,false,true,true,true,true,false
// #default,_result,,,,,,
// ,result,table,_start,_stop,_measurement,_field,_value
// ,,0,2021-01-01T00:00:00Z,2021-01-03T01:00:00Z,m,t,4.8
// "
//
// t_sum = (table=<-) =>
//   (table
//     |> range(start:2021-01-01T00:00:00Z, stop:2021-01-03T01:00:00Z)
//     |> sum())
//
// test _sum = () =>
//   ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_sum})
//
// testing.run(case: _sum)
// ```
run = (case) => {
    return inspect(case: case).diff |> assertEmpty()
}

// benchmark executes a test case without comparing test output with the expected test output.
// This lets you accurately benchmark a test case without the added overhead of comparing test output that occurs in testing.run().
//
// ## Parameters
// - `case` is the test case to benchmark.
//
// ## Define and benchmark a test case
//
//      The following script defines a test case for the sum() function and enables profilers to measure query performance.
//
// ```
// import "testing"
// import "profiler"
//
// option profiler.enabledProfilers = ["query", "operator"]
//
// inData = "
// #datatype,string,long,string,dateTime:RFC3339,string,double
// #group,false,false,true,false,true,false
// #default,_result,,,,,
// ,result,table,_measurement,_time,_field,_value
// ,,0,m,2021-01-01T00:00:00Z,t,1.2
// ,,0,m,2021-01-02T00:00:00Z,t,1.4
// ,,0,m,2021-01-03T00:00:00Z,t,2.2
// "
//
// outData = "
// #datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,double
// #group,false,false,true,true,true,true,false
// #default,_result,,,,,,
// ,result,table,_start,_stop,_measurement,_field,_value
// ,,0,2021-01-01T00:00:00Z,2021-01-03T01:00:00Z,m,t,4.8
// "
//
// t_sum = (table=<-) =>
//   (table
//     |> range(start:2021-01-01T00:00:00Z, stop:2021-01-03T01:00:00Z)
//     |> sum())
//
// test _sum = () =>
//   ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_sum})
//
// testing.benchmark(case: _sum)
// ```
benchmark = (case) => {
    tc = case()

    return tc.input |> tc.fn()
}
