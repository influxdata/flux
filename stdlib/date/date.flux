// Package date provides date and time constants and functions.
package date


// second is a function that returns the second of a specified time. Results
//  range from [0 - 59].
//
// ## Parameters
// - `t` is the time to operate on.
//
//   Use an absolute time, relative duration, or integer. durations are
//   relative to `now()`.
//
// ## Return the second of a time value
//
// ```
// import "date"
//
// date.second(t: 2020-02-11T12:21:03.293534940Z)
// ```
//
// ## Return the second of a relative duration
//
// ```
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.second(t: -50s)
// ```
builtin second : (t: T) => int where T: Timeable

// minute is a function that returns the minute of a specified time. Results
//  range from [0 - 59].
//
// ## Parameters
// - `t` is the time to operate on.
//
//    Use an absolute time, relative duration, or integer. durations are
//    relative to `now()`.
//
// ## Return the minute of a time value
//
// ```
// import "date"
//
// date.minute(t: 2020-02-11T12:21:03.293534940Z)
// ```
//
// ## Return the minute of a relative duration
//
// ```
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.minute(t: -45m)
// ```
builtin minute : (t: T) => int where T: Timeable

// hour is a function that returns the hour of a specified time. Results
//  range from [0 - 23].
//
// ## Parameters
// - `t` is the time to operate on.
//
//   Use an absolute time, relative duration, or integer. durations are
//   relative to `now()`.
//
// ## Return the hour of a time value
//
// ```
// import "date"
//
// date.hour(t: 2020-02-11T12:21:03.293534940Z)
// ```
//
// ## Return the hour of a relative duration
//
// ```
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.hour(t: -8h)
// ```
builtin hour : (t: T) => int where T: Timeable

// weekDay is a function that returns the day of the week for a specified time.
//  Results range from [0 - 6].
//
// ## Parameters
// - `t` is the time to operate on.
//
//   Use an absolute time, relative duration, or integer. durations are
//   relative to `now()`.
//
// ## Return the day of the week for a time value
//
// ```
// import "date"
//
// date.weekDay(t: 2020-02-11T12:21:03.293534940Z)
// ```
//
// ## Return the day of the week for a relative duration
//
// ```
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.weekDay(t: -84h)
// ```
builtin weekDay : (t: T) => int where T: Timeable

// monthDay is a function that returns the day of the month for a specified
//  time. Results range from [1 - 31].
//
// ## Parameters
// - `t` is the time to operate on.
//
//   Use an absolute time, relative duration, or integer. durations are
//   relative to `now()`.
//
// ## Return the day of the month for a time value
//
// ```
// import "date"
//
// date.monthDay(t: 2020-02-11T12:21:03.293534940Z)
// ```
//
// ## Return the day of the month for a relative duration
//
// ```
// import "date"
//
//option now = () => 2020-02-11T12:21:03.293534940Z
//
//date.monthDay(t: -8d)
// ```
builtin monthDay : (t: T) => int where T: Timeable

// yearDay is a function that returns the day of the year for a specified time
//  Results can include leap days and range from [ 1 - 366].
//
// ## Parameters
// - `t` is the time to operate on.
//
//   Use an absolute time, relative duration, or integer. durations are
//   relative to `now()`.
//
// ## Return the day of the year for a time value
//
// ```
// import "date"
//
//date.yearDay(t: 2020-02-11T12:21:03.293534940Z)
// ```
//
// ## Return the day of the year for a relative duration
//
// ```
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.yearDay(t: -1mo)
// ```
builtin yearDay : (t: T) => int where T: Timeable

// month is a function that returns the month of a specified time.
//  Results range from [1 - 12].
//
// ## Parameters
// - `t` is the time to operate on.
//
//   Use an absolute time, relative duration, or integer. durations are
//   relative to `now()`.
//
// ## Return the month of a time value
//
// ```
// import "date"
//
//date.month(t: 2020-02-11T12:21:03.293534940Z)
// ```
//
// ## Retrun the month of a relative duration
//
// ```
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.month(t: -3mo)
// ```
builtin month : (t: T) => int where T: Timeable

// year is a function that returns the year of a specified time.
//
// ## Parameters
// - `t` is the time to operate on.
//
//   Use an absolute time, relative duration, or integer. durations are
//   relative to `now()`.
//
// ## Return the year for a time value
//
// ```
// import "date"
//
//date.year(t: 2020-02-11T12:21:03.293534940Z)
// ```
//
// ## Return the year for a relative duration
//
// ```
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.year(t: -14y)
// ```
builtin year : (t: T) => int where T: Timeable

// week is a function that returns the ISO week of the year for a specified time.
//  Results range from [1 - 53].
//
// ## Parameters
// - `t` is the time to operate on.
//
//   Use an absolute time, relative duration, or integer. durations are
//   relative to `now()`. 
//
// ## Return the week of the year
//
// ```
// import "date"
//
// date.week(t: 2020-02-11T12:21:03.293534940Z)
// ```
//
// ## Return the week of the year using a relative duration
//
// ```
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.week(t: -12d)
// ```
builtin week : (t: T) => int where T: Timeable

// Quarter returns the quarter for a specified time. Results range 
//  from [1-4].
//
// ## Parameters
// - `t` is the time to operate on.
//
//   Use an absolute time, relative duration, or integer. durations are
//   relative to `now()`.
//
// ## Return the quarter for a time value
//
// ```
// import "date"
//
// date.quarter(t: 2020-02-11T12:21:03.293534940Z)
// ```
//
// ## Return the quarter for a relative duration
//
// ```
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.quarter(t: -7mo)
// ```
builtin quarter : (t: T) => int where T: Timeable

// Millisecond returns the milliseconds for a specified time.
//  Results range from [0-999].
//
// ## Parameters
// - `t` is the time to operate on.
//
//   Use an absolute time, relative duration, or integer. durations are
//   relative to `now()`.
//
// ## Return the millisecond of the time value
//
// ```
// import "date"
//
// date.millisecond(t: 2020-02-11T12:21:03.293534940Z)
// ```
//
// ## Return the millisecond of a relative duration
//
// ```
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.millisecond(t: -150ms)
// ```
builtin millisecond : (t: T) => int where T: Timeable

// Microsecond returns the microseconds for a specified time.
//  Results range from [0-999999].
//
// ## Parameters
// - `t` is the time to operate on.
//
//   Use an absolute time, relative duration, or integer. durations are
//   relative to `now()`.
//
// ## Return the microsecond of a time value
//
// ```
// import "date"
//
// date.microsecond(t: 2020-02-11T12:21:03.293534940Z)
// ```
//
// ## Return the microsecond of a relative duration
//
// ```
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.microsecond(t: -1890us)
// ```
builtin microsecond : (t: T) => int where T: Timeable

// Nanosecond returns the nanoseconds for a specified time.
// Results range from [0-999999999].
//
// ## Parameters
// - `t` is the time to operate on.
//
//   Use an absolute time, relative duration, or integer. durations are
//   relative to `now()`.
//
// ## Return the nanosecond for a time value
//
// ```
// import "date"
//
// date.nanosecond(t: 2020-02-11T12:21:03.293534940Z)
// ```
//
// ## Return the nanosecond for a relative duration
//
// ```
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.nanosecond(t: -2111984ns)
// ```
builtin nanosecond : (t: T) => int where T: Timeable

// Truncate returns a time truncated to the specified duration unit.
//
// ## Parameters
// - `t` is the time to operate on.
//
//   Use an absolute time, relative duration, or integer. durations are
//   relative to `now()`.
//
// - `unit` is the unit of time to truncate to
//
//   Only use 1 and the unit of time to specify the unit. For example:
//   1s, 1m, 1h.
//
// ## Example
//
// ```
// import "date"
//
// date.truncate(
//   t: 2019-07-17T12:05:21.012Z
//   unit: 1s
// )
// ```
//
// ## Truncate time values
//
// ```
// import "date"
//
// date.truncate(t: 2019-06-03T13:59:01.000000000Z, unit: 1s)
// // Returns 2019-06-03T13:59:01.000000000Z
//
// date.truncate(t: 2019-06-03T13:59:01.000000000Z, unit: 1m)
// // Returns 2019-06-03T13:59:00.000000000Z
//
// date.truncate(t: 2019-06-03T13:59:01.000000000Z, unit: 1h)
// // Returns 2019-06-03T13:00:00.000000000Z
// ```
//
// ## Truncate time values using durations
//
// ```
// import "date"
//
// option now = () => 2020-01-01T00:00:30.500000000Z
//
// date.truncate(t: -30s, unit: 1s)
// // Returns 2019-12-31T23:59:30.000000000Z
//
// date.truncate(t: -1m, unit: 1m)
// // Returns 2019-12-31T23:59:00.000000000Z
//
// date.truncate(t: -1h, unit: 1h)
// // Returns 2019-12-31T23:00:00.000000000Z
// ```
builtin truncate : (t: T, unit: duration) => time where T: Timeable

// Sunday is a constant that represents Sunday as a day of the week
Sunday = 0

// Monday is a constant that represents Monday as a day of the week.
Monday = 1

// Tuesday is a constant that represents Tuesday as a day of the week.
Tuesday = 2

// Wednesday is a constant that represents Wednesday as a day of the week.
Wednesday = 3

// Thursday is a constant that represents Thursday as a day of the week.
Thursday = 4

// Friday is a constant that represents Friday as a day of the week.
Friday = 5

// Saturday is a constant that represents Saturday as a day of the week.
Saturday = 6

// January is a constant that represents the month of January.
January = 1

// February is a constant that represents the month of February.
February = 2

// March is a constant that represents the month of March.
March = 3

// April is a constant that represents the month of April.
April = 4

// May is a constant that represents the month of May.
May = 5

// June is a constant that represents the month of June.
June = 6

// July is a constant that represents the month of July.
July = 7

// August is a constant that represents the month of August.
August = 8

// September is a constant that represents the month of September.
September = 9

// October is a constant that represents the month of October.
October = 10

// November is a constant that represents the month of November.
November = 11

// December is a constant that represents the month of December.
December = 12
