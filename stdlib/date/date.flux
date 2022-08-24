// Package date provides date and time constants and functions.
//
// ## Metadata
// introduced: 0.37.0
// tags: date/time
package date


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
// ### Return the second of a relative duration
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
//
// ### Return the current second
//
// ```no_run
// import "date"
//
// date.second(t: now())
// ```
//
builtin second : (t: T) => int where T: Timeable

// builtin _time used by time
builtin _time : (t: T, location: {zone: string, offset: duration}) => time where T: Timeable

// time returns the time value of a specified relative duration or time.
//
// `date.time` assumes duration values are relative to `now()`.
//
// ## Parameters
// - t: Duration or time value.
//
//   Use an absolute time or relative duration.
//   Durations are relative to `now()`.
//
// - location: Location used to determine timezone.
//   Default is the `location` option.
//
// ## Examples
//
// ### Return the time for a given time
//
// ```no_run
// import "date"
//
// date.time(t: 2020-02-11T12:21:03.293534940Z)
// // Returns 2020-02-11T12:21:03.293534940Z
// ```
//
// ### Return the time for a given relative duration
//
// ```no_run
// import "date"
//
// option now = () => 2022-01-01T00:00:00.000000000Z
//
// date.time(t: -1h)
//
// // Returns 2021-12-31T23:00:00.000000000Z
// ```
//
// ## Metadata
// introduced: 0.172.0
//
time = (t, location=location) => _time(t, location)

// builtin _minute used by minute
builtin _minute : (t: T, location: {zone: string, offset: duration}) => int where T: Timeable

// minute returns the minute of a specified time. Results range from `[0 - 59]`.
//
// ## Parameters
// - t: Time to operate on.
//
//    Use an absolute time, relative duration, or integer.
//    Durations are relative to `now()`.
//
// - location: Location used to determine timezone.
//   Default is the `location` option.
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
//
// ### Return the current minute
//
// ```no_run
// import "date"
//
// date.minute(t: now())
// ```
//
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
//
// - location: Location used to determine timezone.
//   Default is the `location` option.
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
//
// ### Return the current hour
//
// ```no_run
// import "date"
//
// date.hour(t: now())
// ```
//
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
//
// - location: Location used to determine timezone.
//   Default is the `location` option.
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
//
// ### Return the current day of the week
//
// ```no_run
// import "date"
//
// date.weekDay(t: now())
// ```
//
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
//
// - location: Location used to determine timezone.
//   Default is the `location` option.
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
//
// ### Return the current day of the month
//
// ```no_run
// import "date"
//
// date.monthDay(t: now())
// ```
//
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
//
// - location: Location used to determine timezone.
//   Default is the `location` option.
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
//
// ### Return the current day of the year
//
// ```no_run
// import "date"
//
// date.yearDay(t: now())
// ```
//
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
//
// - location: Location used to determine timezone.
//   Default is the `location` option.
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
//
// ### Return the current numeric month
//
// ```no_run
// import "date"
//
// date.month(t: now())
// ```
//
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
//
// - location: Location used to determine timezone.
//   Default is the `location` option.
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
//
// ### Return the current year
//
// ```no_run
// import "date"
//
// date.year(t: now())
// ```
//
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
//
// - location: Location used to determine timezone.
//   Default is the `location` option.
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
//
// ### Return the current week of the year
//
// ```no_run
// import "date"
//
// date.week(t: now())
// ```
//
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
//
// - location: Location used to determine timezone.
//   Default is the `location` option.
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
//
// ### Return the current quarter
//
// ```no_run
// import "date"
//
// date.quarter(t: now())
// ```
//
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
//
// ### Return the current millisecond unit
//
// ```no_run
// import "date"
//
// date.millisecond(t: now())
// ```
//
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
//
// ### Return the current microsecond unit
//
// ```no_run
// import "date"
//
// date.microsecond(t: now())
// ```
//
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
//
// ### Return the current nanosecond unit
//
// ```no_run
// import "date"
//
// date.nanosecond(t: now())
// ```
//
builtin nanosecond : (t: T) => int where T: Timeable

// builtin _add used by add
builtin _add : (d: duration, to: T, location: {zone: string, offset: duration}) => time
    where
    T: Timeable

// add adds a duration to a time value and returns the resulting time value.
//
// ## Parameters
// - d: Duration to add.
// - to: Time to add the duration to.
// - location: Location to use for the time value.
//
//   Use an absolute time or a relative duration.
//   Durations are relative to `now()`.
//
// ## Examples
//
// ### Add six hours to a timestamp
// ```no_run
// import "date"
//
// date.add(
//     d: 6h,
//     to: 2019-09-16T12:00:00Z,
// )
//
// // Returns 2019-09-16T18:00:00.000000000Z
// ```
//
// ### Add one month to yesterday
//
// A time may be represented as either an explicit timestamp
// or as a relative time from the current `now` time. add can
// support either type of value.
//
// ```no_run
// import "date"
//
// option now = () => 2021-12-10T16:27:40Z
//
// date.add(d: 1mo, to: -1d)
//
// // Returns 2022-01-09T16:27:40Z
// ```
//
// ### Add six hours to a relative duration
// ```no_run
// import "date"
//
// option now = () => 2022-01-01T12:00:00Z
//
// date.add(d: 6h, to: 3h)
//
// // Returns 2022-01-01T21:00:00.000000000Z
// ```
//
// ## Metadata
// tags: date/time
// introduced: 0.162.0
//
add = (d, to, location=location) => _add(d, to, location)

// builtin _sub used by sub
builtin _sub : (from: T, d: duration, location: {zone: string, offset: duration}) => time
    where
    T: Timeable

// sub subtracts a duration from a time value and returns the resulting time value.
//
// ## Parameters
// - from: Time to subtract the duration from.
//
//   Use an absolute time or a relative duration.
//   Durations are relative to `now()`.
//
// - d: Duration to subtract.
// - location: Location to use for the time value.
//
// ## Examples
//
// ### Subtract six hours from a timestamp
// ```no_run
// import "date"
//
// date.sub(from: 2019-09-16T12:00:00Z, d: 6h)
//
// // Returns 2019-09-16T06:00:00.000000000Z
// ```
//
// ### Subtract six hours from a relative duration
// ```no_run
// import "date"
//
// option now = () => 2022-01-01T12:00:00Z
//
// date.sub(d: 6h, from: -3h)
//
// // Returns 2022-01-01T03:00:00.000000000Z
// ```
//
// ### Subtract two days from one hour ago
//
// A time may be represented as either an explicit timestamp
// or as a relative time from the current `now` time. sub can
// support either type of value.
//
// ```no_run
// import "date"
//
// option now = () => 2021-12-10T16:27:40Z
//
// date.sub(
//     from: -1h,
//     d: 2d,
// )
//
// // Returns 2021-12-08T15:27:40Z
// ```
//
// ## Metadata
// tags: date/time
// introduced: 0.162.0
//
sub = (d, from, location=location) => _sub(d, from, location)

// builtin _truncate used by truncate
builtin _truncate : (t: T, unit: duration, location: {zone: string, offset: duration}) => time
    where
    T: Timeable

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
// - location: Location used to determine timezone.
//   Default is the `location` option.
//
// ## Examples
//
// ### Truncate time values
//
// ```no_run
// import "date"
// import "timezone"
//
// option location = timezone.location(name: "Europe/Madrid")
//
// date.truncate(t: 2019-06-03T13:59:01.000000000Z, unit: 1s)
// // Returns 2019-06-03T13:59:01.000000000Z
//
// date.truncate(t: 2019-06-03T13:59:01.000000000Z, unit: 1m)
// // Returns 2019-06-03T13:59:00.000000000Z
//
// date.truncate(t: 2019-06-03T13:59:01.000000000Z, unit: 1h)
// // Returns 2019-06-03T13:00:00.000000000Z
//
// date.truncate(t: 2019-06-03T13:59:01.000000000Z, unit: 1d)
// // Returns 2019-06-02T22:00:00.000000000Z
//
// date.truncate(t: 2019-06-03T13:59:01.000000000Z, unit: 1mo)
// // Returns 2019-05-31T22:00:00.000000000Z
//
// date.truncate(t: 2019-06-03T13:59:01.000000000Z, unit: 1y)
// // Returns 2018-12-31T23:00:00.000000000Z
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
//
// ### Query data from this year
//
// ```no_run
// import "date"
//
// from(bucket: "example-bucket")
//     |> range(start: date.truncate(t: now(), unit: 1y))
// ```
//
// ### Query data from this calendar month
//
// ```no_run
// import "date"
//
// from(bucket: "example-bucket")
//     |> range(start: date.truncate(t: now(), unit: 1mo))
// ```
//
truncate = (t, unit, location=location) => _truncate(t, unit, location)

// scale will multiply the duration by the given value.
//
// ## Parameters
// - d: Duration to scale.
// - n: Amount to scale the duration by.
//
// ## Examples
//
// ### Add n hours to a time
//
// ```no_run
// import "date"
//
// n = 5
// d = date.scale(d: 1h, n: n)
// date.add(d: d, to: 2022-05-10T00:00:00Z)
//
// // Returns 2022-05-10T00:00:00.000000000Z
// ```
//
// ### Add scaled mixed duration to a time
//
// ```no_run
// import "date"
//
// n = 5
// d = date.scale(d: 1mo1h, n: 5)
// date.add(d: d, to: 2022-01-01T00:00:00Z)
//
// // Returns 2022-06-01T05:00:00.000000000Z
// ```
//
// ## Metadata
// tags: date/time
//
builtin scale : (d: duration, n: int) => duration

// Sunday is a constant that represents Sunday as a day of the week.
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
