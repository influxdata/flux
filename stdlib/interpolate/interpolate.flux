// Flux interpolate package provides functions that insert rows for missing data
// at regular intervals and estimate values using different interpolation methods.
package interpolate


// linear is a function that inserts rows at regular intervals using linear
//  interpolation to determine values for inserted rows. 
//
// ## Function Requirements
// - Input data must have _time and _value columns.
// - All columns other than _time and _value must be part of the group key.
//
// ## Parameters
// - `every` is the duration of time between interpolated points.
//
// Interpolate missing data by day
//
// ```
// import "interpolate"
//
// data
//   |> interpolate.linear(every: 1d)
// ```
// # Input
// _time | _value
// --- | ---
// 2021-01-01T00:00:00Z | 10.0
// 2021-01-02T00:00:00Z | 20.0
// 2021-01-04T00:00:00Z | 40.0
// 2021-01-05T00:00:00Z | 50.0
// 2021-01-08T00:00:00Z | 80.0
// 2021-01-09T00:00:00Z | 90.0
//
// # Output
// _time | _value
// --- | ---
// 2021-01-01T00:00:00Z | 10.0
// 2021-01-02T00:00:00Z | 20.0
// 2021-01-04T00:00:00Z | 40.0
// 2021-01-05T00:00:00Z | 50.0
// 2021-01-06T00:00:00Z | 60.0
// 2021-01-07T00:00:00Z | 70.0
// 2021-01-08T00:00:00Z | 80.0
// 2021-01-09T00:00:00Z | 90.0
//
builtin linear : (
    <-tables: [{T with
        _time: time,
        _value: float,
    }],
    every: duration,
) => [{T with
    _time: time,
    _value: float,
}]
