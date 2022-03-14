// Package interpolate provides functions that insert rows for missing data
// at regular intervals and estimate values using different interpolation methods.
//
// ## Metadata
// introduced: 0.87.0
//
package interpolate


// linear inserts rows at regular intervals using linear interpolation to
// determine values for inserted rows.
//
// ### Function requirements
// - Input data must have `_time` and `_value` columns.
// - All columns other than `_time` and `_value` must be part of the group key.
//
// ## Parameters
// - every: Duration of time between interpolated points.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Interpolate missing data by day
// ```
// # import "array"
// import "interpolate"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, _value: 10.0},
// #         {_time: 2021-01-02T00:00:00Z, _value: 20.0},
// #         {_time: 2021-01-04T00:00:00Z, _value: 40.0},
// #         {_time: 2021-01-05T00:00:00Z, _value: 50.0},
// #         {_time: 2021-01-08T00:00:00Z, _value: 80.0},
// #         {_time: 2021-01-09T00:00:00Z, _value: 90.0},
// #     ],
// # )
//
// < data
// >     |> interpolate.linear(every: 1d)
// ```
//
// ## Metadata
// tags: transformations
//
builtin linear : (
        <-tables: stream[{T with _time: time, _value: float}],
        every: duration,
    ) => stream[{T with _time: time, _value: float}]
