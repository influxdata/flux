package date


// Second returns the seconds of a specified time. Results range from [0-59].
builtin second : (t: T) => int where T: Timeable

// Minute returns the minutes of a specified time. Results range from [0-59].
builtin minute : (t: T) => int where T: Timeable

// Hour returns the hours of a specified time. Results range from [0-23].
builtin hour : (t: T) => int where T: Timeable

// WeekDay returns the day of the week for a specified time. Results range from [0-6].
builtin weekDay : (t: T) => int where T: Timeable

// MonthDay returns the day of the month for a specified time. Results range from [1-31].
builtin monthDay : (t: T) => int where T: Timeable

// YearDay returns the day of the year for a specified time. Results range from [1-366].
builtin yearDay : (t: T) => int where T: Timeable

// Month returns the month for a specified time. Results range from [1-12].
builtin month : (t: T) => int where T: Timeable

// Year returns the year for a specified time.
builtin year : (t: T) => int where T: Timeable

// Week returns the ISO week of the year for a specified time. Results range from [1-53].
builtin week : (t: T) => int where T: Timeable

// Quarter returns the quarter for a specified time. Results range from [1-4].
builtin quarter : (t: T) => int where T: Timeable

// Millisecond returns the milliseconds for a specified time. Results range from [0-999].
builtin millisecond : (t: T) => int where T: Timeable

// Microsecond returns the microseconds for a specified time. Results range from [0-999999].
builtin microsecond : (t: T) => int where T: Timeable

// Nanosecond returns the nanoseconds for a specified time. Results range from [0-999999999].
builtin nanosecond : (t: T) => int where T: Timeable

// Truncate returns a time truncated to the specified duration unit.
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
