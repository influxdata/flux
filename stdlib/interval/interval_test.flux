// Three rules:
// 1. When you add a duration  to a time you expect all smaller units of time to remain the same
// 2. When you add a duration to a time you expect the unit being added to change by the amount specified.
// 3. When you cannot achieve rule 1 then follow rule 2.
// Examples:
// Adding a day you expect the time of day to be the same
// Adding a month you expect the day of the month and the time of day to be the same
//
// In cases where that is not possible follow rule 2.
// For example adding 1 month to Jan 31 the day of the month cannot be the same so use the last day of February.
// Using any days in march would violate rule 2 because it would increase the month unit by 2 even though you only added 1 month.
// Constant time lookup
// Mixed durations do not commute or assiosicate so when they are involved in the alegbra of determining a window they make the math hard/impossible.
// Therefore we have certain restrictions in order to make window lookup constant
//
// Possbile equations for window definitions:
//
//    Option 1
//       window_start_i = (epoch + offset) + every * i
//       window_stop_i  = window_start_i + period
//
//   Option 2
//       window_start_i = epoch + every * i + offset
//       window_stop_i  = epoch + every * i + period + offset
// Edge cases
// When you can't follow rule 1
// When the every duration is on the same order of magnitude as the gap in time
// i.e. group by 1h or 2h or 3h over day light savings change
// When a mixed time adjust things across boundaries
// When a period is negative
package main


import "experimental"
import "interval"

experimental.addDuration(d: 1d, to: 2020-01-01T00:00:00Z) == 2020-01-02T00:00:00Z or die(msg: "day addition")
experimental.addDuration(d: 1mo, to: 2020-01-01T00:00:00Z) == 2020-02-01T00:00:00Z or die(msg: "month addition")
experimental.addDuration(d: 1mo, to: 2020-01-31T00:00:00Z) == 2020-02-29T00:00:00Z or die(msg: "month addition end of month")
experimental.addDuration(d: 1mo, to: 2020-02-28T00:00:00Z) == 2020-03-28T00:00:00Z or die(msg: "month addition ??")

// per minute intervals
interval.intervals(every: 1m, period: 1m, offset: 0s)(start: 2020-10-30T00:00:00Z, stop: 2020-10-30T00:10:00Z) == [
    {start: 2020-10-30T00:00:00Z, stop: 2020-10-30T00:01:00Z},
    {start: 2020-10-30T00:01:00Z, stop: 2020-10-30T00:02:00Z},
    {start: 2020-10-30T00:02:00Z, stop: 2020-10-30T00:03:00Z},
    {start: 2020-10-30T00:03:00Z, stop: 2020-10-30T00:04:00Z},
    {start: 2020-10-30T00:04:00Z, stop: 2020-10-30T00:05:00Z},
    {start: 2020-10-30T00:05:00Z, stop: 2020-10-30T00:06:00Z},
    {start: 2020-10-30T00:06:00Z, stop: 2020-10-30T00:07:00Z},
    {start: 2020-10-30T00:07:00Z, stop: 2020-10-30T00:08:00Z},
    {start: 2020-10-30T00:08:00Z, stop: 2020-10-30T00:09:00Z},
    {start: 2020-10-30T00:09:00Z, stop: 2020-10-30T00:10:00Z},
] or die(msg: "per minute intervals")

// daily
interval.intervals(every: 1d, period: 1d, offset: 11h)(start: 2020-10-30T11:00:00Z, stop: 2020-11-05T11:00:00Z) == [
    {start: 2020-10-30T11:00:00Z, stop: 2020-10-31T11:00:00Z},
    {start: 2020-10-31T11:00:00Z, stop: 2020-11-01T11:00:00Z},
    {start: 2020-11-01T11:00:00Z, stop: 2020-11-02T11:00:00Z},
    {start: 2020-11-02T11:00:00Z, stop: 2020-11-03T11:00:00Z},
    {start: 2020-11-03T11:00:00Z, stop: 2020-11-04T11:00:00Z},
    {start: 2020-11-04T11:00:00Z, stop: 2020-11-05T11:00:00Z},
] or die(msg: "per day intervals")

// daily 9-5
interval.intervals(every: 1d, period: 8h, offset: 17h)(start: 2020-10-30T00:00:00Z, stop: 2020-11-05T00:00:00Z) == [
    {start: 2020-10-30T09:00:00Z, stop: 2020-10-30T17:00:00Z},
    {start: 2020-10-31T09:00:00Z, stop: 2020-10-31T17:00:00Z},
    {start: 2020-11-01T09:00:00Z, stop: 2020-11-01T17:00:00Z},
    {start: 2020-11-02T09:00:00Z, stop: 2020-11-02T17:00:00Z},
    {start: 2020-11-03T09:00:00Z, stop: 2020-11-03T17:00:00Z},
    {start: 2020-11-04T09:00:00Z, stop: 2020-11-04T17:00:00Z},
] or die(msg: "per day 9AM-5PM intervals")
