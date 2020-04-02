package experimental

builtin addDuration
builtin subDuration

// An experimental version of group that has mode: "extend"
builtin group

// objectKeys produces a list of the keys existing on the object
builtin objectKeys

// set adds the values from the object onto each row of a table
builtin set

// An experimental version of "to" that:
// - Expects pivoted data
// - Any column in the group key is made a tag in storage
// - All other columns are fields
// - An error will be thrown for incompatible data types
builtin to

// An experimental version of join.
builtin join

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
