// Package date provides date and time constants and functions.
//
// introduced: 0.37.0
// tags: date/time
package date


import "timezone"

// location is a function option whose default behaviour is to return linear clock and no offset
option location = timezone.utc

// second returns the second of a specified time. Results range from `[0 - 59]`.
//
// ## Parameters
// - t: Time to operate on.
//
//   Use an absolute time, relative duration, or integer.
//   Durations are relative to `now()`.
//
// ## Examples
//
// ### Return the second of a time value
//
// ```no_run
// import "date"
//
// date.second(t: 2020-02-11T12:21:03.293534940Z)
//
// // Returns 3
// ```
//
// ## Return the second of a relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.second(t: -50s)
//
// // Returns 28
// ```
builtin second : (t: T) => int where T: Timeable

// builtin _minute used by minute
builtin _minute : (t: T, location: {zone: string, offset: duration}) => int where T: Timeable

// minute returns the minute of a specified time. Results range from `[0 - 59]`.
//
// ## Parameters
// - t: Time to operate on.
//
//    Use an absolute time, relative duration, or integer.
//    Durations are relative to `now()`.
// - location: location loads a timezone based on a location name.
//
// ## Examples
//
// ### Return the minute of a time value
//
// ```no_run
// import "date"
//
// date.minute(t: 2020-02-11T12:21:03.293534940Z)
// // Returns 21
// ```
//
// ### Return the minute of a relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.minute(t: -45m)
//
// // Returns 6
// ```
minute = (t, location=location) => _minute(t, location)

// builtin _hour used by hour
builtin _hour : (t: T, location: {zone: string, offset: duration}) => int where T: Timeable

// hour returns the hour of a specified time. Results range from `[0 - 23]`.
//
// ## Parameters
// - t: Time to operate on.
//
//   Use an absolute time, relative duration, or integer.
//   Durations are relative to `now()`.
// - location: location loads a timezone based on a location name.
//
// ## Examples
//
// ### Return the hour of a time value
//
// ```no_run
// import "date"
//
// date.hour(t: 2020-02-11T12:21:03.293534940Z)
//
// // Returns 12
// ```
//
// ### Return the hour of a relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.hour(t: -8h)
//
// // Returns 7
// ```
hour = (t, location=location) => _hour(t, location)

// builtin _weekDay used by weekDay
builtin _weekDay : (t: T, location: {zone: string, offset: duration}) => int where T: Timeable

// weekDay returns the day of the week for a specified time.
// Results range from `[0 - 6]`.
//
// ## Parameters
// - t: Time to operate on.
//
//   Use an absolute time, relative duration, or integer.
//   Durations are relative to `now()`.
// - location: location loads a timezone based on a location name.
//
// ## Examples
//
// ### Return the day of the week for a time value
//
// ```no_run
// import "date"
//
// date.weekDay(t: 2020-02-11T12:21:03.293534940Z)
//
// // Returns 2
// ```
//
// ### Return the day of the week for a relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.weekDay(t: -84h)
//
// // Returns 6
// ```
weekDay = (t, location=location) => _weekDay(t, location)

// builtin _monthDay used by monthDay
builtin _monthDay : (t: T, location: {zone: string, offset: duration}) => int where T: Timeable

// monthDay returns the day of the month for a specified time.
// Results range from `[1 - 31]`.
//
// ## Parameters
// - t: Time to operate on.
//
//   Use an absolute time, relative duration, or integer.
//   Durations are relative to `now()`.
// - location: location loads a timezone based on a location name.
//
// ## Examples
//
// ### Return the day of the month for a time value
//
// ```no_run
// import "date"
//
// date.monthDay(t: 2020-02-11T12:21:03.293534940Z)
//
// // Returns 11
// ```
//
// ### Return the day of the month for a relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.monthDay(t: -8d)
//
// // Returns 25
// ```
monthDay = (t, location=location) => _monthDay(t, location)

// builtin _yearDay used by yearDay
builtin _yearDay : (t: T, location: {zone: string, offset: duration}) => int where T: Timeable

// yearDay returns the day of the year for a specified time.
// Results can include leap days and range from `[1 - 366]`.
//
// ## Parameters
// - t: Time to operate on.
//
//   Use an absolute time, relative duration, or integer.
//   Durations are relative to `now()`.
// - location: location loads a timezone based on a location name.
//
// ## Examples
//
// ### Return the day of the year for a time value
//
// ```no_run
// import "date"
//
// date.yearDay(t: 2020-02-11T12:21:03.293534940Z)
//
// // Returns 42
// ```
//
// ### Return the day of the year for a relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.yearDay(t: -1mo)
//
// // Returns 276
// ```
yearDay = (t, location=location) => _yearDay(t, location)

// builtin _month used by month
builtin _month : (t: T, location: {zone: string, offset: duration}) => int where T: Timeable

// month returns the month of a specified time. Results range from `[1 - 12]`.
//
// ## Parameters
// - t: Time to operate on.
//
//   Use an absolute time, relative duration, or integer.
//   Durations are relative to `now()`.
// - location: location loads a timezone based on a location name.
//
// ## Examples
//
// ### Return the month of a time value
//
// ```no_run
// import "date"
//
// date.month(t: 2020-02-11T12:21:03.293534940Z)
//
// // Returns 2
// ```
//
// ### Return the month of a relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.month(t: -3mo)
//
// // Returns 8
// ```
month = (t, location=location) => _month(t, location)

// builtin _year used by year
builtin _year : (t: T, location: {zone: string, offset: duration}) => int where T: Timeable

// year returns the year of a specified time.
//
// ## Parameters
// - t: Time to operate on.
//
//   Use an absolute time, relative duration, or integer.
//   Durations are relative to `now()`.
// - location: location loads a timezone based on a location name.
//
// ## Examples
//
// ### Return the year for a time value
//
// ```no_run
// import "date"
//
// date.year(t: 2020-02-11T12:21:03.293534940Z)
//
// // Returns 2020
// ```
//
// ### Return the year for a relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.year(t: -14y)
//
// // Returns 2007
// ```
year = (t, location=location) => _year(t, location)

// builtin _week used by week
builtin _week : (t: T, location: {zone: string, offset: duration}) => int where T: Timeable

// week returns the ISO week of the year for a specified time.
// Results range from `[1 - 53]`.
//
// ## Parameters
// - t: Time to operate on.
//
//   Use an absolute time, relative duration, or integer.
//   Durations are relative to `now()`.
// - location: location loads a timezone based on a location name.
//
// ## Examples
//
// ### Return the week of the year
//
// ```no_run
// import "date"
//
// date.week(t: 2020-02-11T12:21:03.293534940Z)
//
// // Returns 7
// ```
//
// ### Return the week of the year using a relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.week(t: -12d)
//
// // Returns 42
// ```
week = (t, location=location) => _week(t, location)

// builtin _quarter used by quarter
builtin _quarter : (t: T, location: {zone: string, offset: duration}) => int where T: Timeable

// quarter returns the quarter for a specified time. Results range from `[1-4]`.
//
// ## Parameters
// - t: Time to operate on.
//
//   Use an absolute time, relative duration, or integer.
//   Durations are relative to `now()`.
// - location: location loads a timezone based on a location name.
//
// ## Examples
//
// ### Return the quarter for a time value
//
// ```no_run
// import "date"
//
// date.quarter(t: 2020-02-11T12:21:03.293534940Z)
//
// // Returns 1
// ```
//
// ### Return the quarter for a relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.quarter(t: -7mo)
//
// // Returns 2
// ```
quarter = (t, location=location) => _quarter(t, location)

// millisecond returns the milliseconds for a specified time.
// Results range from `[0-999]`.
//
// ## Parameters
// - t: Time to operate on.
//
//   Use an absolute time, relative duration, or integer.
//   Durations are relative to `now()`.
//
// ## Examples
//
// ### Return the millisecond of the time value
//
// ```no_run
// import "date"
//
// date.millisecond(t: 2020-02-11T12:21:03.293534940Z)
//
// // Returns 293
// ```
//
// ### Return the millisecond of a relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.millisecond(t: -150ms)
//
// // Returns 127
// ```
builtin millisecond : (t: T) => int where T: Timeable

// microsecond returns the microseconds for a specified time.
// Results range `from [0-999999]`.
//
// ## Parameters
// - t: Time to operate on.
//
//   Use an absolute time, relative duration, or integer.
//   Durations are relative to `now()`.
//
// ## Examples
//
// ### Return the microsecond of a time value
//
// ```no_run
// import "date"
//
// date.microsecond(t: 2020-02-11T12:21:03.293534940Z)
//
// // Returns 293534
// ```
//
// ### Return the microsecond of a relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.microsecond(t: -1890us)
//
// // Returns 322661
// ```
builtin microsecond : (t: T) => int where T: Timeable

// nanosecond returns the nanoseconds for a specified time.
// Results range from `[0-999999999]`.
//
// ## Parameters
// - t: Time to operate on.
//
//   Use an absolute time, relative duration, or integer.
//   Durations are relative to `now()`.
//
// ## Examples
//
// ### Return the nanosecond for a time value
//
// ```no_run
// import "date"
//
// date.nanosecond(t: 2020-02-11T12:21:03.293534940Z)
//
// // Returns 293534940
// ```
//
// ### Return the nanosecond for a relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2020-02-11T12:21:03.293534940Z
//
// date.nanosecond(t: -2111984ns)
//
// // Returns 128412016
// ```
builtin nanosecond : (t: T) => int where T: Timeable

// truncate returns a time truncated to the specified duration unit.
//
// ## Parameters
// - t: Time to operate on.
//
//   Use an absolute time, relative duration, or integer.
//   Durations are relative to `now()`.
//
// - unit: Unit of time to truncate to.
//
//   Only use 1 and the unit of time to specify the unit.
//   For example: `1s`, `1m`, `1h`.
//
// ## Examples
//
// ### Truncate time values
//
// ```no_run
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
// ### Truncate time values using relative durations
//
// ```no_run
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
