// Package join
package join


// join
//
// ## Parameters
// - left:
// - right:
// - on:
// - as:
// - method:
builtin join : (
        <-left: stream[L],
        right: stream[R],
        on: (l: L, r: R) => bool,
        as: (l: L, r: R) => A,
        method: string,
    ) => stream[A]
    where
    A: Record,
    L: Record,
    R: Record
