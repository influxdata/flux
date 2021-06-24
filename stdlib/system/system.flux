package system


// time is a function that returns the current system time
//
// ## Example
// ```
// import "system"
//
// data
//   |> set(key: "processed_at", value: string(v: system.time() ))
// ```
builtin time : () => time
