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
