package interpolate

builtin linear : (<-tables: [{ T with
    _time:  time,
    _value: float }], every: duration) => [{ T with
    _time:  time,
    _value: float }]
