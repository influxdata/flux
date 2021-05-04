package debug


// pass will pass any incoming tables directly next to the following transformation.
// It is best used to interrupt any planner rules that rely on a specific ordering.
builtin pass : (<-tables: [A]) => [A] where A: Record

// slurp will read the incoming tables and concatenate buffers with the same group key
// into a single table. This is useful for testing the performance impact of multiple
// buffers versus a single buffer.
builtin slurp : (<-tables: [A]) => [A] where A: Record

// sink will discard all data that comes into it.
builtin sink : (<-tables: [A]) => [A] where A: Record
