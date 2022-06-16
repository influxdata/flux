// Package join is under active development and is not yet ready for public consumption.
package join


// tables is under active development and is not yet ready for public consumption.
//
// ## Parameters
// - left:
// - right:
// - on:
// - as:
// - method:
builtin tables : (
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

// time is under active development and is not yet ready for public consumption.
//
// ## Parameters
// - left:
// - right:
// - as:
// - method:
time = (left=<-, right, as, method="inner") =>
    tables(
        left: left,
        right: right,
        on: (l, r) => l._time == r._time,
        as: as,
        method: method,
    )

// inner is under active development and is not yet ready for public consumption.
//
// ## Parameters
// - left:
// - right:
// - on:
// - as:
inner = (left=<-, right, on, as) =>
    tables(
        left: left,
        right: right,
        on: on,
        as: as,
        method: "inner",
    )

// full is under active development and is not yet ready for public consumption.
//
// ## Parameters
// - left:
// - right:
// - on:
// - as:
full = (left=<-, right, on, as) =>
    tables(
        left: left,
        right: right,
        on: on,
        as: as,
        method: "full",
    )

// left is under active development and is not yet ready for public consumption.
//
// ## Parameters
// - left:
// - right:
// - on:
// - as:
left = (left=<-, right, on, as) =>
    tables(
        left: left,
        right: right,
        on: on,
        as: as,
        method: "left",
    )

// right is under active development and is not yet ready for public consumption.
//
// ## Parameters
// - left:
// - right:
// - on:
// - as:
right = (left=<-, right, on, as) =>
    tables(
        left: left,
        right: right,
        on: on,
        as: as,
        method: "right",
    )
