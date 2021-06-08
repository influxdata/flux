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

// # Days of the week
//
// The days of the week are represented as integers in the range `[0-6]`
Sunday = 0
Monday = 1
Tuesday = 2
Wednesday = 3
Thursday = 4
Friday = 5
Saturday = 6

// # Months of the year
//
// Months are represented as integers in the range `[1-12]`
January = 1
February = 2
March = 3
April = 4
May = 5
June = 6
July = 7
August = 8
September = 9
October = 10
November = 11
December = 12
