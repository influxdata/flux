package generate

builtin from : (start: A, stop: A, count: int, fn: (n: int) => int) => [{ _start: time , _stop: time , _time: time , _value:int }] where A: Timeable
