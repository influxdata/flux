// Package testing provides functions for testing Flux operations.
//
// ## Metadata
// introduced: 0.14.0
//
package testing


import "array"
import c "csv"
import "experimental"

//  tags is a list of tags that will be applied to a test case.
//
//  The test harness allows filtering based on included tags.
//
//  Tags are expected to be overridden per test file and test case
//  using normal option semantics.
option tags = []

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
            experimental.diff(got, want)
                |> yield(name: "errorOutput")
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

// shouldError calls a function that catches any error and checks that the error matches the expected value.
//
// ## Parameters
// - fn: Function to call.
// - want: Regular expression to match the expected error.
//
// ## Examples
//
// ### Test die function errors
//
// ```no_run
// import "testing"
//
// testing.shouldError(fn: () => die(msg: "error message"), want: /error message/)
// ```
//
// ## Metadata
// introduced: 0.174.0
// tags: tests
//
shouldError = (fn, want) => {
    got = experimental.catch(fn)

    return
        if exists got.msg then
            array.from(rows: [{v: got.msg}])
                |> filter(fn: (r) => r.v !~ want)
                |> yield(name: "errorOutput")
        else
            die(msg: "shouldError expected an error")
}
