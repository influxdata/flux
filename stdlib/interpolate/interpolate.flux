package interpolate


// Linear inserts rows at regular intervals using linear interpolation
// to determine the value of any missing rows.
//
// Example
//
//    import "interpolate"
//    import "array"
//    
//    array.from(
//        rows: [
//            {_time: 2021-01-01T00:00:00Z, _value: 10.0},
//            {_time: 2021-01-02T00:00:00Z, _value: 20.0},
//            {_time: 2021-01-04T00:00:00Z, _value: 40.0},
//            {_time: 2021-01-05T00:00:00Z, _value: 50.0},
//            {_time: 2021-01-08T00:00:00Z, _value: 80.0},
//            {_time: 2021-01-09T00:00:00Z, _value: 90.0},
//        ],
//    )
//        |> interpolate.linear(every: 1d)
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
