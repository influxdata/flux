// Package generate provides functions for generating data.
//
// ## Metadata
// introduced: 0.17.0
// tags: generate
package generate


// from generates data using the provided parameter values.
//
// ## Parameters
// - count: Number of rows to generate.
// - fn: Function used to generate values.
//
//   The function takes an `n` parameter that represents the row index, operates
//   on `n`, and then returns an integer value. Rows use zero-based indexing.
//
// - start: Beginning of the time range to generate values in.
// - stop: End of the time range to generate values in.
//
// ## Examples
//
// ### Generate sample data
// ```
// import "generate"
//
// generate.from(
//     count: 6,
//     fn: (n) => (n + 1) * (n + 2),
//     start: 2021-01-01T00:00:00Z,
//     stop: 2021-01-02T00:00:00Z,
// )
// ```
//
// ## Metadata
// tags: inputs
builtin from : (
        start: A,
        stop: A,
        count: int,
        fn: (n: int) => int,
    ) => stream[{_start: time, _stop: time, _time: time, _value: int}]
    where
    A: Timeable
