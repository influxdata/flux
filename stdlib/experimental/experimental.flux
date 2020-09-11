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
builtin to : (<-tables: [A], ?bucket: string, ?bucketID: string, ?org: string, ?orgID: string, ?host: string, ?token: string) => [A] where A: Record

// An experimental version of join.
builtin join : (left: [A], right: [B], fn: (left: A, right: B) => C) => [C] where A: Record, B: Record, C: Record

builtin chain : (first: [A], second: [B]) => [B] where A: Record, B: Record

// Aligns all tables to a common start time by using the same _time value for
// the first record in each table and incrementing all subsequent _time values
// using time elapsed between input records.
// By default, it aligns to tables to 1970-01-01T00:00:00Z UTC.
alignTime = (tables=<-, alignTo=time(v: 0)) =>
  tables
    |> stateDuration(
      fn: (r) => true,
      column: "timeDiff",
      unit: 1ns
    )
    |> map(fn: (r) =>
      ({ r with _time: time(v: (int(v: alignTo ) + r.timeDiff)) })
    )
    |> drop(columns: ["timeDiff"])
