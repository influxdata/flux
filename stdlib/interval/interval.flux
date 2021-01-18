package interval


builtin intervals : (
    every: A,
    period: B,
    offset: C,
) => (
    start: D,
    stop: E,
) => [{
    start: time,
    stop: time,
}] where
    A: Timeable,
    B: Timeable,
    C: Timeable,
    D: Timeable,
    E: Timeable
