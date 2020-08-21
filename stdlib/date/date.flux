package date

builtin second : (t: T) => int where T: Timeable
builtin minute : (t: T) => int where T: Timeable
builtin hour : (t: T) => int where T: Timeable
builtin weekDay : (t: T) => int where T: Timeable
builtin monthDay : (t: T) => int where T: Timeable
builtin yearDay : (t: T) => int where T: Timeable
builtin month : (t: T) => int where T: Timeable
builtin year : (t: T) => int where T: Timeable
builtin week : (t: T) => int where T: Timeable
builtin quarter : (t: T) => int where T: Timeable
builtin millisecond : (t: T) => int where T: Timeable
builtin microsecond : (t: T) => int where T: Timeable
builtin nanosecond : (t: T) => int where T: Timeable
builtin truncate : (t: T) => int where T: Timeable

Sunday    = 0
Monday    = 1
Tuesday   = 2
Wednesday = 3
Thursday  = 4
Friday    = 5
Saturday  = 6

January   = 1
February  = 2
March     = 3
April     = 4
May       = 5
June      = 6
July      = 7
August    = 8
September = 9
October   = 10
November  = 11
December  = 12
