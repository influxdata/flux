// Package system provides functions for reading values from the system.
//
// introduced: 0.18.0
//
package system


// time returns the current system time.
//
// ## Examples
//
// ### Return a stream of tables with the current system time
// ```
// import "array"
// import "system"
//
// array.from(rows:[{time: system.time()}])
// ```
//
// tags: date/time
//
builtin time : () => time
