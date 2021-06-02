package date


// date.second is a function that returns the second of the specified time.
// Results can range from [0 - 59].
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
builtin second : (t: T) => int where T: Timeable

// date.minute is a function that returns the minute of a specified time.
// Results range from [0 - 59].
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
builtin minute : (t: T) => int where T: Timeable

// date.hour is a function that returns the hour of the specified time.
// Results range from [0 - 23].
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
builtin hour : (t: T) => int where T: Timeable

// date.weekDay is a funcion that returns the day of the week for a specified time.
// Results range from [0 - 6].
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
builtin weekDay : (t: T) => int where T: Timeable

// date.monthDay is a function that returns the day of the month for a specified time.
// Results range from [1 - 31].
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
builtin monthDay : (t: T) => int where T: Timeable

// date.yearDay is a function that returns the day of the year for a specified time.
// Results include leap days, and range from [1 - 366].
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
builtin yearDay : (t: T) => int where T: Timeable

// date.month is a function that returns the month of a specified time.
// Results range from [1 - 12].
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
builtin month : (t: T) => int where T: Timeable

// date.year is a function that returns the year of a specified time.
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
builtin year : (t: T) => int where T: Timeable

// date.week is a function that returns the ISO week of the year for a
// specified time. Results range from [1 - 53].
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
builtin week : (t: T) => int where T: Timeable

// date.quarter is a function that returns the quarter of the year for a
// specified time. Results range from [1 - 4].
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
builtin quarter : (t: T) => int where T: Timeable

// date.millisecond is a function that returns the millisecond of a
// specified time. Results range from [0 - 999].
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
builtin millisecond : (t: T) => int where T: Timeable

// date.microsecond is a function that returns the microsecond of a
// specified time. Results range from [0 - 999999].
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
builtin microsecond : (t: T) => int where T: Timeable

// date.nanosecond is a function that returns the nanosecond of a
// specified time. Results range from [0 - 999999999].
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
builtin nanosecond : (t: T) => int where T: Timeable

// date.truncate returns a time truncated to the specified duration unit.
//
// - `t` is the time to operate on. Use an absolute time, relative duration,
//	or integer. Durations are relative to now().
//
// - `unit` is the unit of time to truncate to.
//
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
