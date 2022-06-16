// Package testing provides functions for testing Flux operations.
//
// ## Metadata
// introduced: 0.14.0
//
package testing


import "array"
import c "csv"

// assertEquals tests whether two streams of tables are identical.
//
// If equal, the function outputs the tested data stream unchanged.
// If unequal, the function returns an error.
//
// assertEquals can be used to perform in-line tests in a query.
//
// ## Parameters
// - name: Unique assertion name.
// - got: Data to test. Default is piped-forward data (`<-`).
// - want: Expected data to test against.
//
// ## Examples
//
// ### Test if streams of tables are different
// ```
// import "sampledata"
// import "testing"
//
// want = sampledata.int()
// got = sampledata.float() |> toInt()
//
// testing.assertEquals(name: "test_equality", got: got, want: want)
// ```
//
// ### Test if streams of tables are different mid-script
// ```no_run
// import "testing"
//
// want = from(bucket: "backup-example-bucket")
//     |> range(start: -5m)
//
// from(bucket: "example-bucket")
//     |> range(start: -5m)
//     |> testing.assertEquals(want: want)
// ```
//
// ## Metadata
// tags: tests
//
builtin assertEquals : (name: string, <-got: stream[A], want: stream[A]) => stream[A]

// assertEmpty tests if an input stream is empty. If not empty, the function returns an error.
//
// assertEmpty can be used to perform in-line tests in a query.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Check if there is a difference between streams
// This example uses `testing.diff()` to output the difference between two streams of tables.
// `testing.assertEmpty()` checks to see if the difference is empty.
//
// ```
// import "sampledata"
// import "testing"
//
// want = sampledata.int()
// got = sampledata.float() |> toInt()
//
// got
//     |> testing.diff(want: want)
//     |> testing.assertEmpty()
// ```
//
// ## Metadata
// introduced: 0.18.0
// tags: tests
//
builtin assertEmpty : (<-tables: stream[A]) => stream[A]

builtin _diff : (
        <-got: stream[A],
        want: stream[A],
        ?verbose: bool,
        ?epsilon: float,
        ?nansEqual: bool,
    ) => stream[{A with _diff: string}]

// diff produces a diff between two streams.
//
// The function matches tables from each stream based on group keys.
// For each matched table, it produces a diff.
// Any added or removed rows are added to the table as a row.
// An additional string column with the name diff is created and contains a
// `-` if the row was present in the `got` table and not in the `want` table or
// `+` if the opposite is true.
//
// `diff()` function emits at least one row if the tables are
// different and no rows if the tables are the same.
// The exact diff produced may change.
// `diff()` can be used to perform in-line diffs in a query.
//
// ## Parameters
// - got: Stream containing data to test. Default is piped-forward data (`<-`).
// - want: Stream that contains data to test against.
// - epsilon: Specify how far apart two float values can be, but still considered equal. Defaults to 0.000000001.
// - verbose: Include detailed differences in output. Default is `false`.
// - nansEqual: Consider `NaN` float values equal. Default is `false`.
//
// ## Examples
//
// ### Output a diff between two streams of tables
// ```
// import "sampledata"
// import "testing"
//
// want = sampledata.int()
// got = sampledata.int()
//     |> map(fn: (r) => ({r with _value: if r._value > 15 then r._value + 1 else r._value }))
//
// < testing.diff(got: got, want: want)
// ```
//
// ### Return a diff between a stream of tables an the expected output
// ```no_run
// import "testing"
//
// want = from(bucket: "backup-example-bucket") |> range(start: -5m)
//
// from(bucket: "example-bucket")
//     |> range(start: -5m)
//     |> testing.diff(want: want)
// ```
//
// ## Metadata
// introduced: 0.18.0
// tags: tests
//
diff = (
    got=<-,
    want,
    verbose=false,
    epsilon=0.000001,
    nansEqual=false,
) =>
    {
        return
            _diff(
                got,
                want,
                verbose,
                epsilon,
                nansEqual,
            )
                |> yield(name: "error")
    }

// load loads test data from a stream of tables.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Load a raw stream of tables in a test case
// The following test uses `array.from()` to create two streams of tables to
// compare in the test.
//
// ```
// import "testing"
// import "array"
//
// got =
//     array.from(
//         rows: [
//             {_time: 2021-01-01T00:00:00Z, _measurement: "m", _field: "t", _value: 1.2},
//             {_time: 2021-01-01T01:00:00Z, _measurement: "m", _field: "t", _value: 0.8},
//             {_time: 2021-01-01T02:00:00Z, _measurement: "m", _field: "t", _value: 3.2},
//         ],
//     )
//
// want =
//     array.from(
//         rows: [
//             {_time: 2021-01-01T00:00:00Z, _measurement: "m", _field: "t", _value: 1.2},
//             {_time: 2021-01-01T01:00:00Z, _measurement: "m", _field: "t", _value: 0.8},
//             {_time: 2021-01-01T02:00:00Z, _measurement: "m", _field: "t", _value: 3.1},
//         ],
//     )
//
// testing.load(tables: got)
//     |> testing.diff(want: want)
// ```
//
// ## Metadata
// introduced: 0.112.0
//
option load = (tables=<-) => tables

// inspect returns information about a test case.
//
// ## Parameters
// - case: Test case to inspect.
//
// ## Examples
//
// ### Define and inspect a test case
// ```no_run
// import "testing"
//
// inData =
//     "
// #datatype,string,long,string,dateTime:RFC3339,string,double
// #group,false,false,true,false,true,false
// #default,_result,,,,,
// ,result,table,_measurement,_time,_field,_value
// ,,0,m,2021-01-01T00:00:00Z,t,1.2
// ,,0,m,2021-01-02T00:00:00Z,t,1.4
// ,,0,m,2021-01-03T00:00:00Z,t,2.2
// "
//
// outData =
//     "
// #datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,double
// #group,false,false,true,true,true,true,false
// #default,_result,,,,,,
// ,result,table,_start,_stop,_measurement,_field,_value
// ,,0,2021-01-01T00:00:00Z,2021-01-03T01:00:00Z,m,t,4.8
// "
//
// t_sum = (table=<-) =>
//     table
//         |> range(start: 2021-01-01T00:00:00Z, stop: 2021-01-03T01:00:00Z)
//         |> sum()
//
// test _sum = () => ({input: csv.from(csv: inData), want: csv.from(csv: outData), fn: t_sum})
//
// testing.inpsect(case: _sum)
//
// // Returns: {
// //     fn: (<-table: [{_time: time | t10997}]) -> [t10996],
// //     input: fromCSV -> range -> map,
// //     want: fromCSV -> yield,
// //     got: fromCSV -> range -> map -> range -> sum -> yield,
// //     diff: ( fromCSV; fromCSV -> range -> map -> range -> sum;  ) -> diff -> yield
// // }
// ```
//
// ## Metadata
// introduced: 0.18.0
//
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
// - case: Test case to run.
//
// ## Examples
//
// ### Define and execute a test case
// ```
// import "csv"
// import "testing"
//
// inData =
//     "
// #datatype,string,long,string,dateTime:RFC3339,string,double
// #group,false,false,true,false,true,false
// #default,_result,,,,,
// ,result,table,_measurement,_time,_field,_value
// ,,0,m,2021-01-01T00:00:00Z,t,1.2
// ,,0,m,2021-01-02T00:00:00Z,t,1.4
// ,,0,m,2021-01-03T00:00:00Z,t,2.2
// "
//
// outData =
//     "
// #datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,double
// #group,false,false,true,true,true,true,false
// #default,_result,,,,,,
// ,result,table,_start,_stop,_measurement,_field,_value
// ,,0,2021-01-01T00:00:00Z,2021-01-03T01:00:00Z,m,t,4.8
// "
//
// t_sum = (table=<-) =>
//     table
//         |> range(start: 2021-01-01T00:00:00Z, stop: 2021-01-03T01:00:00Z)
//         |> sum()
//
// test _sum = () => ({input: csv.from(csv: inData), want: csv.from(csv: outData), fn: t_sum})
//
// testing.run(case: _sum)
// ```
//
// ## Metadata
// introduced: 0.20.0
//
run = (case) => {
    return inspect(case: case).diff |> assertEmpty()
}

// benchmark executes a test case without comparing test output with the expected test output.
// This lets you accurately benchmark a test case without the added overhead of
// comparing test output that occurs in `testing.run()`.
//
// ## Parameters
// - case: Test case to benchmark.
//
// ## Examples
//
// ### Define and benchmark a test case
// The following script defines a test case for the sum() function and enables
// profilers to measure query performance.
//
// ```
// import "csv"
// import "testing"
// import "profiler"
//
// option profiler.enabledProfilers = ["query", "operator"]
//
// inData =
//     "
// #datatype,string,long,string,dateTime:RFC3339,string,double
// #group,false,false,true,false,true,false
// #default,_result,,,,,
// ,result,table,_measurement,_time,_field,_value
// ,,0,m,2021-01-01T00:00:00Z,t,1.2
// ,,0,m,2021-01-02T00:00:00Z,t,1.4
// ,,0,m,2021-01-03T00:00:00Z,t,2.2
// "
//
// outData =
//     "
// #datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,double
// #group,false,false,true,true,true,true,false
// #default,_result,,,,,,
// ,result,table,_start,_stop,_measurement,_field,_value
// ,,0,2021-01-01T00:00:00Z,2021-01-03T01:00:00Z,m,t,4.8
// "
//
// t_sum = (table=<-) =>
//     table
//         |> range(start: 2021-01-01T00:00:00Z, stop: 2021-01-03T01:00:00Z)
//         |> sum()
//
// test _sum = () => ({input: csv.from(csv: inData), want: csv.from(csv: outData), fn: t_sum})
//
// testing.benchmark(case: _sum)
// ```
//
// ## Metadata
// introduced: 0.49.0
//
benchmark = (case) => {
    tc = case()

    return tc.input |> tc.fn()
}

// assertEqualValues tests whether two values are equal.
//
// ## Parameters
// - got: Value to test.
// - want: Expected value to test against.
//
// ## Examples
//
// ### Test if two values are equal
// ```
// import "testing"
//
// < testing.assertEqualValues(got: 5, want: 12)
// ```
//
// ## Metadata
// introduced: 0.141.0
// tags: tests
//
assertEqualValues = (got, want) => {
    return diff(got: array.from(rows: [{v: got}]), want: array.from(rows: [{v: want}]))
}
