// Package debug provides methods for debugging the Flux engine.
package debug


// pass will pass any incoming tables directly next to the following transformation.
// It is best used to interrupt any planner rules that rely on a specific ordering.
//
// ## Parameters
// - tables: Stream to pass unmodified to next transformation.
//
builtin pass : (<-tables: stream[A]) => stream[A] where A: Record

// opaque works like `pass` in that it passes any incoming tables directly to the
// following transformation, save for its type signature does not indicate that the
// input type has any correlation with the output type.
//
// ## Parameters
// - tables: Stream to pass unmodified to next transformation.
//
builtin opaque : (<-tables: stream[A]) => stream[B] where A: Record, B: Record

// slurp will read the incoming tables and concatenate buffers with the same group key
// into a single in memory table buffer. This is useful for testing the performance impact of multiple
// buffers versus a single buffer.
//
// ## Parameters
// - tables: Stream to consume into single buffers per table.
//
builtin slurp : (<-tables: stream[A]) => stream[A] where A: Record

// sink will discard all data that comes into it.
//
// ## Parameters
// - tables: Stream to discard.
//
builtin sink : (<-tables: stream[A]) => stream[A] where A: Record

// getOption gets the value of an option using a form of reflection.
//
// ## Parameters
// - pkg: Full path of the package.
// - name: Option name.
//
builtin getOption : (pkg: string, name: string) => A

// feature returns the value associated with the given feature flag.
//
// ## Parameters
// - key: Feature flag name.
//
builtin feature : (key: string) => A

// vectorize controls whether the compiler attempts to vectorize Flux functions.
option vectorize = false

// unpivot removes the `_time` column and any other column not in the group key and outputs a new table with `_field` and `_value` columns pairs.
// The output stream retains the group key and all group key columns of the input stream.
// Specialized to transform the pivoted output from `iox.from()` into the unpivoted format.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
builtin unpivot : (<-tables: stream[{A with _time: time}]) => stream[{B with _field: string, _value: C}]
    where
    A: Record,
    B: Record
