// Package boundaries provides operators for finding the boundaries around certain days, months, and weeks.
//
// ## Metadata
// introduced: NEXT
package boundaries


import "system"
import "date"
import "math"
import "strings"
import "regexp"
import "experimental/table"

// yesterday returns a record with `start` and `stop` boundary timestamps for yesterday.
//
// Yesterday is relative to `now()`.
//
// ## Examples
//
// ### Return start and stop timestamps of yesterday
// ```no_run
// import "date/boundaries"
//
// option now = () => 2022-01-02T13:45:28Z
//
// boundaries.yesterday()
// // Returns {start: 2022-01-01T00:00:00.000000000Z, stop: 2022-01-02T00:00:00.000000000Z}
// ```
//
// ### Query data from yesterday
// ```no_run
// import "date/boundaries"
//
//  day = boundaries.yesterday()
//
// from(bucket: "example-bucket")
//     |> range(start: day.start, stop: day.stop )
// ```
//
// ## Metadata
// introduced: NEXT
// tags: date/time
yesterday = () => {
    ret = {start: date.sub(d: 1d, from: today()), stop: today()}

    return ret
}

_timezone_convert = (s) => {
    t = int(v: location.offset)
    as = if t == 0 then 0h else duration(v: -t)

    return
        if t >= 0 then
            date.add(d: as, to: s)
        else
            date.add(d: duration(v: int(v: as) - int(v: 1d)), to: s)
}

_day_finder = (td, func, offset=0h) => {
    day_date = today()
    today_date = if int(v: offset) != 0 then date.add(d: offset, to: day_date) else day_date
    cur_day = date.weekDay(t: today_date)

    val =
        if cur_day == date.Sunday then
            7 - td
        else if td >= cur_day then
            7 - (td - cur_day)
        else
            cur_day - td

    scaled_offset =
        if cur_day == td then
            date.scale(d: 1w, n: 1)
        else
            date.scale(d: 1d, n: val)

    day_calc = date.sub(d: scaled_offset, from: today_date)

    return func(s: _timezone_convert(s: day_calc))
}

_day_formatter = (s) => {
    return {start: s, stop: date.add(d: 1d, to: s)}
}

_week_formatter = (s) => {
    return {start: s, stop: date.add(d: 1w, to: s)}
}

// monday returns a record with `start` and `stop` boundary timestamps of last Monday.
// Last Monday is relative to `now()`. If today is Monday, the function returns boundaries for the previous Monday.
//
// ## Examples
//
// ### Return start and stop timestamps of last Monday
//
// ```no_run
// import "date/boundaries"
//
// option location = timezone.fixed(offset: -8h)
// option now = () => 2021-12-30T00:40:44Z
//
// boundaries.monday()
// // Returns {start: 2021-12-27T08:00:00Z, stop:2021-12-28T08:00:00Z }
// ```
//
//
//
// ### Query data collected last Monday
// ```no_run
// import "date/boundaries"
//
// day = boundaries.monday()
//
// from(bucket: "example-bucket")
//     |> range(start: day.start, stop: day.stop)
// ```
// ## Metadata
// tags: date/time
//
monday = () => {
    return _day_finder(td: date.Monday, func: _day_formatter)
}

// tuesday returns a record with `start` and `stop` boundary timestamps of last Tuesday.
//
// Last Tuesday is relative to `now()`. If today is Tuesday, the function returns boundaries for the previous Tuesday.
//
// ## Examples
//
// ### Return start and stop timestamps of last Tuesday
//
// ```no_run
// import "date/boundaries"
//
// option location = timezone.fixed(offset: -8h)
// option now = () => 2021-12-30T00:40:44Z
//
// boundaries.tuesday()
// // Returns {start: 2021-12-28T08:00:00Z, stop:2021-12-29T08:00:00Z }
// ```
//
//
// ### Query data collected last Tuesday
// ```no_run
// import "date/boundaries"
//
// day = boundaries.tuesday()
//
// from(bucket: "example-bucket")
//     |> range(start: day.start, stop: day.stop)
// ```
//
// ## Metadata
// tags: date/time
//
tuesday = () => {
    return _day_finder(td: date.Tuesday, func: _day_formatter)
}

// wednesday returns a record with `start` and `stop` boundary timestamps for last Wednesday.
//
// Last Wednesday is relative to `now()`. If today is Wednesday, the function returns boundaries for the previous Wednesday.
//
// ## Examples
//
// // ### Return start and stop timestamps of last Wednesday
//
// ```no_run
// import "date/boundaries"
//
// option location = timezone.fixed(offset: -8h)
// option now = () => 2021-12-30T00:40:44Z
//
// boundaries.wednesday()
// // Returns {start: 2021-12-29T08:00:00Z, stop:2021-12-30T08:00:00Z }
// ```
//
//
// ### Query data collected last Wednesday
// ```no_run
// import "date/boundaries"
//
// day = boundaries.wednesday()
//
// from(bucket: "example-bucket")
//     |> range(start: day.start, stop: day.stop)
// ```
//
// This will return all records from Wednesday this week
//
// ## Metadata
// tags: date/time
//
wednesday = () => {
    return _day_finder(td: date.Wednesday, func: _day_formatter)
}

// thursday returns a record with `start` and `stop` boundary timestamps for last Thursday.
//
// Last Thursday is relative to `now()`. If today is Thursday, the function returns boundaries for the previous Thursday.
// ## Examples
//
// ### Return start and stop timestamps of last Thursday
//
// ```no_run
// import "date/boundaries"
//
// option location = timezone.fixed(offset: -8h)
// option now = () => 2021-12-30T00:40:44Z
//
// boundaries.thursday()
// // Returns {start: 2021-12-23T08:00:00Z, stop:2021-12-24T08:00:00Z }
// ```
//
// ### Query data collected last Thursday
// ```no_run
// import "date/boundaries"
//
// day = boundaries.thursday()
//
// from(bucket: "example-bucket")
//     |> range(start: day.start, stop: day.stop)
// ```
//
// ## Metadata
// tags: date/time
//
thursday = () => {
    return _day_finder(td: date.Thursday, func: _day_formatter)
}

// friday returns a record with `start` and `stop` boundary timestamps for last Friday.
//
// Last Friday is relative to `now()`. If today is Friday, the function returns boundaries for the previous Friday.
// ## Examples
//
// ### Return start and stop timestamps of last Friday
//
// ```no_run
// import "date/boundaries"
//
// option location = timezone.fixed(offset: -8h)
// option now = () => 2021-12-30T00:40:44Z
//
// boundaries.friday()
// // Returns {start: 2021-12-24T08:00:00Z, stop:2022-12-25T08:00:00Z }
// ```
//
// ### Query data collected last Friday
// ```no_run
// import "date/boundaries"
//
// day = boundaries.friday()
//
// from(bucket: "example-bucket")
//     |> range(start: day.start, stop: day.stop)
// ```
//
// ## Metadata
// tags: date/time
//
friday = () => {
    return _day_finder(td: date.Friday, func: _day_formatter)
}

// saturday returns a record with `start` and `stop` boundary timestamps for last Saturday.
//
// Last Saturday is relative to `now()`. If today is Saturday, the function returns boundaries for the previous Saturday.
//
// ## Examples
//
// ### Return start and stop timestamps of last Saturday
//
// ```no_run
// import "date/boundaries"
//
// option location = timezone.fixed(offset: -8h)
// option now = () => 2021-12-30T00:40:44Z
//
// boundaries.saturday()
// // Returns {start: 2022-12-25T08:00:00Z, stop:2022-12-26T08:00:00Z }
// ```
//
// ### Query data collected last Saturday
// ```no_run
// import "date/boundaries"
//
// day = boundaries.saturday()
//
// from(bucket: "example-bucket")
//     |> range(start: day.start, stop: day.stop)
// ```
//
// ## Metadata
// tags: date/time
//
saturday = () => {
    return _day_finder(td: date.Saturday, func: _day_formatter)
}

// sunday returns a record with `start` and `stop` boundary timestamps for last Sunday.
//
// Last Sunday is relative to `now()`. If today is Sunday, the function returns boundaries for the previous Sunday.
//
// ## Examples
//
// ### Return start and stop timestamps of last Sunday
//
// ```no_run
// import "date/boundaries"
//
// option location = timezone.fixed(offset: -8h)
// option now = () => 2021-12-30T00:40:44Z
//
// boundaries.sunday()
// // Returns {start: 2021-12-26T08:00:00Z, stop:2021-12-27T08:00:00Z }
// ```
//
// ### Query data collected last Sunday
// ```no_run
// import "date/boundaries"
//
// day = boundaries.sunday()
//
// from(bucket: "example-bucket")
//     |> range(start: day.start, stop: day.stop)
// ```
//
// ## Metadata
// tags: date/time
//
sunday = () => {
    return _day_finder(td: date.Sunday, func: _day_formatter)
}

// month returns a record with `start` and `stop` boundary timestamps for the current month.
//
// `now()` determines the current month.
//
// ## Parameters
// - month_offset: Number of months to offset from the current month. Default is `0`.
//
//   Use a negative offset to return boundaries from previous months.
//   Use a positive offset to return boundaries for future months.
//
// ## Examples
//
// ### Return start and stop timestamps for the current month
//
// ```no_run
// import "date/boundaries"
//
// option now = () => 2022-05-10T10:10:00Z
//
// boundaries.month()
// // Returns {start:2022-05-01T00:00:00.000000000Z, stop:2022-06-01T00:00:00.000000000Z}
// ```
//
//
// ### Query data from this month
// ```no_run
// import "date/boundaries"
//
// thisMonth = boundaries.month()
//
// from(bucket: "example-bucket")
//     |> range(start: thisMonth.start, stop: thisMonth.stop)
// ```
//
// ### Query data from last month
//
// ```no_run
// import "date/boundaries"
//
// lastMonth = boundaries.month(month_offset: -1)
//
// from(bucket: "example-bucket")
//     |> range(start: lastMonth.start, stop: lastMonth.stop)
// ```
//
// ## Metadata
// tags: date/time
//
month = (month_offset=0) => {
    s = date.truncate(t: today(), unit: 1mo)
    as = _timezone_convert(s: s)
    start = date.add(d: date.scale(d: 1mo, n: month_offset), to: as)

    return {start: start, stop: date.add(d: 1mo, to: start)}
}

// week returns a record with `start` and `stop` boundary timestamps of the current week.
// By default, weeks start on Monday.
//
// ## Parameters
// - start_sunday: Indicate if the week starts on Sunday. Default is `false`.
//
//   When set to `false`, the week starts on Monday.
//
// - week_offset: Number of weeks to offset from the current week. Default is `0`.
//
//   Use a negative offset to return boundaries from previous weeks.
//   Use a positive offset to return boundaries for future weeks.
//
// ## Examples
//
// ### Return start and stop timestamps of the current week starting on Monday
//
// ```no_run
// import "date/boundaries"
//
// option now = () => 2022-05-10T00:00:00.000010000Z
//
// boundaries.week()
// // Returns {start: 2022-05-09T00:00:00.000000000Z, stop: 2022-05-16T00:00:00.000000000Z}
// ```
//
// ### Return start and stop timestamps of the current week starting on Sunday
//
// ```no_run
// import "date/boundaries"
//
// option now = () => 2022-05-10T10:10:00Z
//
// boundaries.week(start_sunday:true)
// // Returns {start: 2022-05-08T00:00:00.000000000Z, stop: 2022-05-14T00:00:00.000000000Z}
// ```
//
// ### Query data from current week
//
// ```no_run
// import "date/boundaries"
//
// thisWeek = boundaries.week()
//
// from(bucket: "example-bucket")
//     |> range(start: thisWeek.start, stop: thisWeek.stop)
// ```
//
// ### Query data from last week
//
// ```no_run
// import "date/boundaries"
//
// lastWeek = boundaries.week(week_offset: -1)
//
// from(bucket: "example-bucket")
//     |> range(start: lastWeek.start, stop: lastWeek.stop)
// ```
//
// ## Metadata
// tags: date/time
//
week = (week_offset=0, start_sunday=false) => {
    return
        if start_sunday then
            _day_finder(td: date.Sunday, func: _week_formatter, offset: date.scale(d: 1w, n: week_offset))
        else
            _day_finder(td: date.Monday, func: _week_formatter, offset: date.scale(d: 1w, n: week_offset))
}
