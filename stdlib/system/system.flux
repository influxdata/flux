// Package system provides functions for reading values from the system.
package system


// time is a function that returns the current system time
//
// ## Example
// ```
// import "system"
// import "array"
//
// array.from(rows:[{time: system.time()}])
// ```
builtin time : () => time
