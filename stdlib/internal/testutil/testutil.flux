// Package testutil provides helper function for writing test cases.
//
// ## Metadata
// introduced: 0.68.0
package testutil


// fail causes the current script to fail.
builtin fail : () => bool

// yield is the identity function.
//
// ## Parameters
// - v: Any value.
builtin yield : (<-v: A) => A

// makeRecord is the identity function, but breaks the type connection from input to output.
//
// ## Parameters
// - o: Record value.
builtin makeRecord : (o: A) => B where A: Record, B: Record

// makeAny constructs any value based on a type description as a string.
//
// ## Parameters
// - typ: Description of the type to create.
builtin makeAny : (typ: string) => A
