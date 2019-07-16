package date

builtin second
builtin minute
builtin hour
builtin weekDay
builtin monthDay
builtin yearDay
builtin month
builtin truncate

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

// hack to simulate an imported strings package
date = {
  second:second,
  minute:minute,
  hour:hour,
  weekDay:weekDay,
  monthDay:monthDay,
  yearDay:yearDay,
  month:month,
  Sunday:Sunday,
  Monday:Monday,
  Tuesday:Tuesday,
  Wednesday:Wednesday,
  Thursday:Thursday,
  Friday:Friday,
  Saturday:Saturday,
  January:January,
  February:February,
  March:March,
  April:April,
  May:May,
  June:June,
  July:July,
  August:August,
  September:September,
  October:October,
  November:November,
  December:December,
  truncate:truncate,
}