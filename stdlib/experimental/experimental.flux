package experimental


builtin addDuration : (d: duration, to: time) => time
builtin subDuration : (d: duration, from: time) => time

// An experimental version of group that has mode: "extend"
builtin group : (<-tables: [A], mode: string, columns: [string]) => [A] where A: Record

// objectKeys produces a list of the keys existing on the object
builtin objectKeys : (o: A) => [string] where A: Record

// set adds the values from the object onto each row of a table
builtin set : (<-tables: [A], o: B) => [C] where A: Record, B: Record, C: Record

// An experimental version of "to" that:
// - Expects pivoted data
// - Any column in the group key is made a tag in storage
// - All other columns are fields
// - An error will be thrown for incompatible data types
builtin to : (
    <-tables: [A],
    ?bucket: string,
    ?bucketID: string,
    ?org: string,
    ?orgID: string,
    ?host: string,
    ?token: string,
) => [A] where
    A: Record

// An experimental version of join.
builtin join : (left: [A], right: [B], fn: (left: A, right: B) => C) => [C] where A: Record, B: Record, C: Record
builtin chain : (first: [A], second: [B]) => [B] where A: Record, B: Record

// Aligns all tables to a common start time by using the same _time value for
// the first record in each table and incrementing all subsequent _time values
// using time elapsed between input records.
// By default, it aligns to tables to 1970-01-01T00:00:00Z UTC.
alignTime = (tables=<-, alignTo=time(v: 0)) => tables
    |> stateDuration(
        fn: (r) => true,
        column: "timeDiff",
        unit: 1ns,
    )
    |> map(fn: (r) => ({r with _time: time(v: int(v: alignTo) + r.timeDiff)}))
    |> drop(columns: ["timeDiff"])

// An experimental version of window.
builtin window : (
    <-tables: [{T with _start: time, _stop: time, _time: time}],
    ?every: duration,
    ?period: duration,
    ?offset: duration,
    ?createEmpty: bool,
) => [{T with _start: time, _stop: time, _time: time}]

// An experimental version of integral.
builtin integral : (<-tables: [{T with _time: time, _value: B}], ?unit: duration, ?interpolate: string) => [{T with _value: B}]

// An experimental version of count.
builtin count : (<-tables: [{T with _value: A}]) => [{T with _value: int}]

// An experimental version of histogramQuantile
builtin histogramQuantile : (<-tables: [{T with _value: float, le: float}], ?quantile: float, ?minValue: float) => [{T with _value: float}]

// An experimental version of mean.
builtin mean : (<-tables: [{T with _value: float}]) => [{T with _value: float}]

// An experimental version of mode.
builtin mode : (<-tables: [{T with _value: A}]) => [{T with _value: A}]

// An experimental version of quantile.
builtin quantile : (<-tables: [{T with _value: float}], q: float, ?compression: float, ?method: string) => [{T with _value: float}]

// An experimental version of skew.
builtin skew : (<-tables: [{T with _value: float}]) => [{T with _value: float}]

// An experimental version of spread.
builtin spread : (<-tables: [{T with _value: A}]) => [{T with _value: A}] where A: Numeric

// An experimental version of stddev.
builtin stddev : (<-tables: [{T with _value: float}], ?mode: string) => [{T with _value: float}]

// An experimental version of sum.
builtin sum : (<-tables: [{T with _value: A}]) => [{T with _value: A}] where A: Numeric

// An experimental version of kaufmansAMA.
builtin kaufmansAMA : (<-tables: [{T with _value: A}], n: int) => [{T with _value: float}] where A: Numeric

// An experimental version of distinct
builtin distinct : (<-tables: [{T with _value: A}]) => [{T with _value: A}]

// An experimental version of fill
builtin fill : (<-tables: [{T with _value: A}], ?value: A, ?usePrevious: bool) => [{T with _value: A}]

// An experimental version of first
builtin first : (<-tables: [{T with _value: A}]) => [{T with _value: A}]

// An experimental version of last
builtin last : (<-tables: [{T with _value: A}]) => [{T with _value: A}]

// An experimental version of max
builtin max : (<-tables: [{T with _value: A}]) => [{T with _value: A}]

// An experimental version of min
builtin min : (<-tables: [{T with _value: A}]) => [{T with _value: A}]

// An experimental version of unique
builtin unique : (<-tables: [{T with _value: A}]) => [{T with _value: A}]

// An experimental version of histogram
builtin histogram : (<-tables: [{T with _value: float}], bins: [float], ?normalize: bool) => [{T with _value: float, le: float}]
