package generate


// From generates a table with count rows using fn to determine the value of each row.
builtin from : (
    start: A,
    stop: A,
    count: int,
    fn: (n: int) => int,
) => [{
    _start: time,
    _stop: time,
    _time: time,
    _value: int,
}] where
    A: Timeable
